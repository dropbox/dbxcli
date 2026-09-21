package cmd

import (
	"bytes"
	"context"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox/files"
	"github.com/spf13/cobra"
)

func TestAdaptiveConcurrencyClimbsWhileThroughputImproves(t *testing.T) {
	c := newAdaptiveConcurrency(4, 16)

	if got := c.observe(100, true); got != 6 {
		t.Fatalf("first sample: limit = %d, want 6", got)
	}
	// The window right after a change is a warm-up and must not change anything.
	if got := c.observe(150, true); got != 6 {
		t.Fatalf("warm-up sample: limit = %d, want 6", got)
	}
	if got := c.observe(150, true); got != 9 {
		t.Fatalf("improved sample: limit = %d, want 9", got)
	}
	if got := c.observe(160, true); got != 9 {
		t.Fatalf("warm-up sample: limit = %d, want 9", got)
	}
	// Less than a 10% gain means the link is saturated: fall back and settle.
	if got := c.observe(155, true); got != 6 {
		t.Fatalf("plateau sample: limit = %d, want 6", got)
	}
	if !c.settled {
		t.Fatal("expected controller to settle after a plateau")
	}
	if got := c.observe(1000, true); got != 6 {
		t.Fatalf("settled controller changed limit to %d", got)
	}
}

func TestAdaptiveConcurrencyIgnoresUninformativeSamples(t *testing.T) {
	c := newAdaptiveConcurrency(4, 16)

	if got := c.observe(100, false); got != 4 {
		t.Fatalf("unsaturated sample: limit = %d, want 4", got)
	}
	if got := c.observe(0, true); got != 4 {
		t.Fatalf("zero sample: limit = %d, want 4", got)
	}
	if c.best != 0 {
		t.Fatalf("best = %v, want no measurement recorded", c.best)
	}
}

func TestAdaptiveConcurrencySettlesAtMax(t *testing.T) {
	c := newAdaptiveConcurrency(4, 6)

	if got := c.observe(100, true); got != 6 {
		t.Fatalf("limit = %d, want 6", got)
	}
	c.observe(100, true) // warm-up
	if got := c.observe(200, true); got != 6 {
		t.Fatalf("limit = %d, want 6", got)
	}
	if !c.settled {
		t.Fatal("expected controller to settle at max")
	}
}

func TestNewAdaptiveConcurrencyClampsInitialToMax(t *testing.T) {
	c := newAdaptiveConcurrency(4, 2)
	if c.limit != 2 || !c.settled {
		t.Fatalf("limit = %d, settled = %v, want 2 and settled", c.limit, c.settled)
	}

	c = newAdaptiveConcurrency(0, 0)
	if c.limit != 1 || c.max != 1 {
		t.Fatalf("limit = %d, max = %d, want 1 and 1", c.limit, c.max)
	}
}

func TestNextConcurrencyStep(t *testing.T) {
	steps := []int{4}
	for steps[len(steps)-1] < 16 {
		steps = append(steps, nextConcurrencyStep(steps[len(steps)-1], 16))
	}
	want := []int{4, 6, 9, 13, 16}
	if len(steps) != len(want) {
		t.Fatalf("steps = %v, want %v", steps, want)
	}
	for i := range want {
		if steps[i] != want[i] {
			t.Fatalf("steps = %v, want %v", steps, want)
		}
	}
	if got := nextConcurrencyStep(1, 16); got != 2 {
		t.Fatalf("nextConcurrencyStep(1) = %d, want 2", got)
	}
}

// concurrencyProbe records how many jobs run at once and holds the first
// `barrier` jobs until that many are in flight, which proves they overlap.
type concurrencyProbe struct {
	barrier int32
	active  atomic.Int32
	peak    atomic.Int32
	calls   atomic.Int32
	ready   chan struct{}
	once    sync.Once
	timeout atomic.Bool
}

func newConcurrencyProbe(barrier int) *concurrencyProbe {
	return &concurrencyProbe{barrier: int32(barrier), ready: make(chan struct{})}
}

func (p *concurrencyProbe) enter() {
	n := p.active.Add(1)
	for {
		peak := p.peak.Load()
		if n <= peak || p.peak.CompareAndSwap(peak, n) {
			break
		}
	}
	if p.calls.Add(1) >= p.barrier {
		p.once.Do(func() { close(p.ready) })
	}
	select {
	case <-p.ready:
	case <-time.After(10 * time.Second):
		p.timeout.Store(true)
	}
}

func (p *concurrencyProbe) leave() {
	p.active.Add(-1)
}

func (p *concurrencyProbe) check(t *testing.T, wantPeak int) {
	t.Helper()
	if p.timeout.Load() {
		t.Fatalf("jobs never reached %d in flight", p.barrier)
	}
	if got := int(p.peak.Load()); got != wantPeak {
		t.Fatalf("peak concurrency = %d, want %d", got, wantPeak)
	}
}

func TestDownloadPoolRunsJobsUpToFixedLimit(t *testing.T) {
	probe := newConcurrencyProbe(3)
	pool := newDownloadPool(3, 8, 0, nil)
	if !pool.concurrent() {
		t.Fatal("expected a fixed limit above 1 to be concurrent")
	}

	var ran atomic.Int32
	jobs := make([]func(), 8)
	for i := range jobs {
		jobs[i] = func() {
			probe.enter()
			defer probe.leave()
			ran.Add(1)
		}
	}
	pool.run(context.Background(), jobs, func(int, error) {
		t.Error("unexpected dispatch error")
	})

	probe.check(t, 3)
	if ran.Load() != 8 {
		t.Fatalf("ran %d jobs, want 8", ran.Load())
	}
	if pool.done.Load() != 8 {
		t.Fatalf("done = %d, want 8", pool.done.Load())
	}
}

func TestDownloadPoolSingleWorkerRunsSequentially(t *testing.T) {
	probe := newConcurrencyProbe(1)
	pool := newDownloadPool(1, 4, 0, nil)
	if pool.concurrent() {
		t.Fatal("expected a single worker not to be concurrent")
	}

	jobs := make([]func(), 4)
	for i := range jobs {
		jobs[i] = func() {
			probe.enter()
			defer probe.leave()
		}
	}
	pool.run(context.Background(), jobs, func(int, error) {
		t.Error("unexpected dispatch error")
	})
	probe.check(t, 1)
}

func TestDownloadPoolAutoStartsWithInitialConcurrency(t *testing.T) {
	pool := newDownloadPool(downloadWorkersAuto, 40, 0, nil)
	if pool.tuner == nil {
		t.Fatal("expected automatic tuner")
	}
	if got := pool.limiter.currentLimit(); got != autoDownloadWorkersInitial {
		t.Fatalf("initial limit = %d, want %d", got, autoDownloadWorkersInitial)
	}
	if pool.tuner.max != autoDownloadWorkersMax {
		t.Fatalf("max = %d, want %d", pool.tuner.max, autoDownloadWorkersMax)
	}

	// Small folders never need more workers than files.
	pool = newDownloadPool(downloadWorkersAuto, 2, 0, nil)
	if got := pool.limiter.currentLimit(); got != 2 {
		t.Fatalf("initial limit for 2 files = %d, want 2", got)
	}
	if pool.tuner.max != 2 {
		t.Fatalf("max for 2 files = %d, want 2", pool.tuner.max)
	}

	pool = newDownloadPool(downloadWorkersAuto, 1, 0, nil)
	if pool.concurrent() {
		t.Fatal("a single file must download sequentially with per-file progress")
	}
}

func TestDownloadPoolAutoTunesWhileRunning(t *testing.T) {
	restoreMonitorIntervals(t, time.Millisecond, 2*time.Millisecond)

	pool := newDownloadPool(downloadWorkersAuto, 64, 64<<20, nil)
	var maxSeen atomic.Int32
	jobs := make([]func(), 64)
	for i := range jobs {
		jobs[i] = func() {
			progress := pool.fileProgress()
			for step := int64(1); step <= 4; step++ {
				progress(step<<18, 1<<20)
				time.Sleep(time.Millisecond)
			}
			limit := int32(pool.limiter.currentLimit())
			for {
				seen := maxSeen.Load()
				if limit <= seen || maxSeen.CompareAndSwap(seen, limit) {
					break
				}
			}
		}
	}
	pool.run(context.Background(), jobs, func(int, error) {
		t.Error("unexpected dispatch error")
	})

	if pool.done.Load() != 64 {
		t.Fatalf("done = %d, want 64", pool.done.Load())
	}
	if got := pool.bytes.Load(); got != 64<<20 {
		t.Fatalf("bytes = %d, want %d", got, 64<<20)
	}
	if limit := int(maxSeen.Load()); limit < 1 || limit > autoDownloadWorkersMax {
		t.Fatalf("observed limit %d outside [1, %d]", limit, autoDownloadWorkersMax)
	}
	if pool.peakLimit < autoDownloadWorkersInitial || pool.peakLimit > autoDownloadWorkersMax {
		t.Fatalf("peak limit %d outside [%d, %d]", pool.peakLimit, autoDownloadWorkersInitial, autoDownloadWorkersMax)
	}
}

func TestDownloadPoolFileProgressCountsOnlyNewBytes(t *testing.T) {
	pool := newDownloadPool(2, 1, 0, nil)
	progress := pool.fileProgress()
	progress(10, 100)
	progress(10, 100)
	progress(35, 100)
	progress(20, 100) // never goes backwards
	if got := pool.bytes.Load(); got != 35 {
		t.Fatalf("bytes = %d, want 35", got)
	}
}

func TestDownloadPoolCancelledContextFailsRemainingJobs(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	pool := newDownloadPool(2, 3, 0, nil)
	var ran atomic.Int32
	jobs := []func(){
		func() { ran.Add(1) },
		func() { ran.Add(1) },
		func() { ran.Add(1) },
	}
	var mu sync.Mutex
	var failed []int
	pool.run(ctx, jobs, func(i int, err error) {
		if err != context.Canceled {
			t.Errorf("job %d error = %v, want context.Canceled", i, err)
		}
		mu.Lock()
		failed = append(failed, i)
		mu.Unlock()
	})

	if ran.Load() != 0 {
		t.Fatalf("ran %d jobs after cancellation, want 0", ran.Load())
	}
	if len(failed) != 3 {
		t.Fatalf("failed jobs = %v, want all three", failed)
	}
}

func TestConcurrencyLimiterReportsSaturation(t *testing.T) {
	l := newConcurrencyLimiter(1)
	if l.takeSaturated() {
		t.Fatal("fresh limiter reported saturation")
	}
	if err := l.acquire(context.Background()); err != nil {
		t.Fatal(err)
	}

	acquired := make(chan struct{})
	go func() {
		_ = l.acquire(context.Background())
		close(acquired)
	}()

	deadline := time.Now().Add(5 * time.Second)
	for !l.takeSaturated() {
		if time.Now().After(deadline) {
			t.Fatal("limiter never reported saturation while a job waited")
		}
		time.Sleep(time.Millisecond)
	}

	l.setLimit(2)
	select {
	case <-acquired:
	case <-time.After(5 * time.Second):
		t.Fatal("raising the limit did not wake the waiting job")
	}
	l.release()
	l.release()
}

func TestDownloadStatusWriterKeepsPipesLineOriented(t *testing.T) {
	var buf bytes.Buffer
	w := newDownloadStatusWriter(&buf)
	if w.live {
		t.Fatal("a buffer must not be treated as a terminal")
	}
	w.status("Downloading 1/2 files")
	w.message("Downloading %s -> %s\n", "/a", "a")
	w.finish("done")
	if got := buf.String(); got != "Downloading /a -> a\n" {
		t.Fatalf("output = %q, want only the message line", got)
	}
}

func TestDownloadStatusWriterClearsStatusLineBeforeMessages(t *testing.T) {
	var buf bytes.Buffer
	w := newDownloadStatusWriter(&buf)
	w.live = true

	w.status("abcdef")
	w.status("xy")
	w.message("hi\n")
	w.finish("done")

	want := "\rabcdef" + "\rxy    " + "\r      \r" + "hi\n" + "done\n"
	if got := buf.String(); got != want {
		t.Fatalf("output = %q, want %q", got, want)
	}
}

func TestParseGetOptionsDefaultsToAutomaticWorkers(t *testing.T) {
	opts, err := parseGetOptions(testGetCmd())
	if err != nil {
		t.Fatal(err)
	}
	if opts.workers != downloadWorkersAuto {
		t.Fatalf("workers = %d, want auto (%d)", opts.workers, downloadWorkersAuto)
	}

	opts, err = parseGetOptions(getCmd)
	if err != nil {
		t.Fatal(err)
	}
	if opts.workers != downloadWorkersAuto {
		t.Fatalf("registered flag default workers = %d, want auto (%d)", opts.workers, downloadWorkersAuto)
	}
}

func TestParseGetOptionsRejectsNegativeWorkers(t *testing.T) {
	cmd := testGetWorkersCmd(-1)
	_, err := parseGetOptions(cmd)
	if err == nil {
		t.Fatal("expected error for negative workers")
	}
	if code := jsonErrorCode(err); code != jsonErrorCodeInvalidArguments {
		t.Fatalf("error code = %q, want %q", code, jsonErrorCodeInvalidArguments)
	}
	if details := jsonErrorDetails(err); details["flag"] != "workers" {
		t.Fatalf("details = %#v, want flag=workers", details)
	}
}

func testGetWorkersCmd(workers int) *cobra.Command {
	cmd := testGetCmd()
	cmd.Flags().IntP("workers", "w", downloadWorkersAuto, "")
	if err := cmd.Flags().Set("recursive", "true"); err != nil {
		panic(err)
	}
	if err := cmd.Flags().Set("workers", strconv.Itoa(workers)); err != nil {
		panic(err)
	}
	return cmd
}

func restoreMonitorIntervals(t *testing.T, monitor, window time.Duration) {
	t.Helper()
	previousMonitor, previousWindow := downloadMonitorInterval, downloadAutoTuneWindow
	downloadMonitorInterval, downloadAutoTuneWindow = monitor, window
	t.Cleanup(func() {
		downloadMonitorInterval, downloadAutoTuneWindow = previousMonitor, previousWindow
	})
}

func parallelGetMock(t *testing.T, paths []string, probe *concurrencyProbe) *mockFilesClient {
	t.Helper()
	entries := []files.IsMetadata{getTestFolderMetadata("/remote")}
	for _, p := range paths {
		entries = append(entries, getTestFileMetadata(p, 4))
	}
	return &mockFilesClient{
		getMetadataFn: func(arg *files.GetMetadataArg) (files.IsMetadata, error) {
			return getTestFolderMetadata(arg.Path), nil
		},
		listFolderFn: func(arg *files.ListFolderArg) (*files.ListFolderResult, error) {
			return &files.ListFolderResult{Entries: entries}, nil
		},
		downloadFn: func(arg *files.DownloadArg) (*files.FileMetadata, io.ReadCloser, error) {
			if probe != nil {
				probe.enter()
				defer probe.leave()
			}
			return getTestFileMetadata(arg.Path, 4), io.NopCloser(strings.NewReader("data")), nil
		},
	}
}

func TestGetRecursiveDownloadsFilesInParallelByDefault(t *testing.T) {
	dst := filepath.Join(t.TempDir(), "out")
	paths := []string{"/remote/a.txt", "/remote/b.txt", "/remote/c.txt", "/remote/d.txt", "/remote/e.txt", "/remote/f.txt"}
	probe := newConcurrencyProbe(autoDownloadWorkersInitial)
	stubFilesClient(t, parallelGetMock(t, paths, probe))

	var stderr bytes.Buffer
	cmd := testGetCmd()
	cmd.SetErr(&stderr)
	if err := cmd.Flags().Set("recursive", "true"); err != nil {
		t.Fatal(err)
	}
	if err := get(cmd, []string{"/remote", dst}); err != nil {
		t.Fatalf("get error: %v", err)
	}

	probe.check(t, autoDownloadWorkersInitial)
	for _, p := range paths {
		got, err := os.ReadFile(filepath.Join(dst, filepath.Base(p)))
		if err != nil {
			t.Fatalf("read %s: %v", p, err)
		}
		if string(got) != "data" {
			t.Fatalf("%s content = %q, want data", p, got)
		}
		if !strings.Contains(stderr.String(), "Downloading "+p+" -> ") {
			t.Fatalf("stderr = %q, want a line for %s", stderr.String(), p)
		}
	}
	if strings.Contains(stderr.String(), "\r") {
		t.Fatalf("stderr = %q, want no progress-bar control characters on a pipe", stderr.String())
	}
}

func TestGetRecursiveWorkersFlagFixesConcurrency(t *testing.T) {
	dst := filepath.Join(t.TempDir(), "out")
	paths := []string{"/remote/a.txt", "/remote/b.txt", "/remote/c.txt", "/remote/d.txt", "/remote/e.txt"}
	probe := newConcurrencyProbe(2)
	stubFilesClient(t, parallelGetMock(t, paths, probe))

	cmd := testGetWorkersCmd(2)
	cmd.SetErr(io.Discard)
	if err := get(cmd, []string{"/remote", dst}); err != nil {
		t.Fatalf("get error: %v", err)
	}
	probe.check(t, 2)
	for _, p := range paths {
		if _, err := os.Stat(filepath.Join(dst, filepath.Base(p))); err != nil {
			t.Fatalf("%s not downloaded: %v", p, err)
		}
	}
}

func TestGetRecursiveSingleWorkerKeepsSequentialProgress(t *testing.T) {
	dst := filepath.Join(t.TempDir(), "out")
	paths := []string{"/remote/a.txt", "/remote/b.txt", "/remote/c.txt"}
	probe := newConcurrencyProbe(1)
	stubFilesClient(t, parallelGetMock(t, paths, probe))

	var stderr bytes.Buffer
	cmd := testGetWorkersCmd(1)
	cmd.SetErr(&stderr)
	if err := get(cmd, []string{"/remote", dst}); err != nil {
		t.Fatalf("get error: %v", err)
	}
	probe.check(t, 1)
	// The sequential path keeps the per-file progress bar on stderr.
	if !strings.Contains(stderr.String(), "Downloading 4 B/4 B") {
		t.Fatalf("stderr = %q, want per-file progress", stderr.String())
	}
}

func TestGetJSONRecursiveResultsKeepListingOrderUnderConcurrency(t *testing.T) {
	dst := filepath.Join(t.TempDir(), "out")
	paths := []string{"/remote/a.txt", "/remote/b.txt", "/remote/c.txt"}
	mock := parallelGetMock(t, paths, nil)

	// Make the first-listed file finish last so completion order differs
	// from listing order.
	lastStarted := make(chan struct{})
	var startOnce sync.Once
	download := mock.downloadFn
	mock.downloadFn = func(arg *files.DownloadArg) (*files.FileMetadata, io.ReadCloser, error) {
		switch arg.Path {
		case "/remote/c.txt":
			startOnce.Do(func() { close(lastStarted) })
		case "/remote/a.txt":
			select {
			case <-lastStarted:
			case <-time.After(10 * time.Second):
				t.Error("c.txt never started while a.txt waited")
			}
		}
		return download(arg)
	}
	stubFilesClient(t, mock)

	var stdout bytes.Buffer
	cmd := testGetJSONCmd(&stdout, nil)
	cmd.Flags().IntP("workers", "w", downloadWorkersAuto, "")
	if err := cmd.Flags().Set("recursive", "true"); err != nil {
		t.Fatal(err)
	}
	if err := cmd.Flags().Set("workers", "3"); err != nil {
		t.Fatal(err)
	}
	if err := get(cmd, []string{"/remote", dst}); err != nil {
		t.Fatalf("get error: %v", err)
	}

	got := decodeGetOutput(t, &stdout)
	wantTargets := []string{dst, filepath.Join(dst, "a.txt"), filepath.Join(dst, "b.txt"), filepath.Join(dst, "c.txt")}
	if len(got.Results) != len(wantTargets) {
		t.Fatalf("results = %+v, want %d entries", got.Results, len(wantTargets))
	}
	for i, want := range wantTargets {
		if got.Results[i].Input.Target != want {
			t.Fatalf("results[%d].target = %q, want %q (results: %+v)", i, got.Results[i].Input.Target, want, got.Results)
		}
	}
	if got.Results[0].Kind != getKindFolder || got.Results[1].Kind != getKindFile {
		t.Fatalf("kinds = %s, %s; want folder then file", got.Results[0].Kind, got.Results[1].Kind)
	}
}

func TestGetRecursiveConcurrentReportsErrorsInListingOrder(t *testing.T) {
	dst := filepath.Join(t.TempDir(), "out")
	paths := []string{"/remote/bad1.txt", "/remote/good.txt", "/remote/bad2.txt", "/remote/also-good.txt"}
	mock := parallelGetMock(t, paths, nil)
	download := mock.downloadFn
	mock.downloadFn = func(arg *files.DownloadArg) (*files.FileMetadata, io.ReadCloser, error) {
		if strings.Contains(arg.Path, "bad") {
			return nil, nil, &files.DownloadAPIError{}
		}
		return download(arg)
	}
	stubFilesClient(t, mock)

	var stderr bytes.Buffer
	cmd := testGetWorkersCmd(4)
	cmd.SetErr(&stderr)
	err := get(cmd, []string{"/remote", dst})
	if err == nil {
		t.Fatal("expected error for failed downloads")
	}
	if !strings.Contains(err.Error(), "2 error(s)") {
		t.Fatalf("error = %q, want 2 error(s)", err.Error())
	}
	first := strings.Index(stderr.String(), "Error: /remote/bad1.txt")
	second := strings.Index(stderr.String(), "Error: /remote/bad2.txt")
	if first < 0 || second < 0 || first > second {
		t.Fatalf("stderr = %q, want errors reported in listing order", stderr.String())
	}
	for _, name := range []string{"good.txt", "also-good.txt"} {
		if _, statErr := os.Stat(filepath.Join(dst, name)); statErr != nil {
			t.Fatalf("%s not downloaded: %v", name, statErr)
		}
	}
}

func TestGetRecursiveConcurrentExportOnlyFilesCountBytes(t *testing.T) {
	dst := filepath.Join(t.TempDir(), "out")
	exported := "exported paper"
	paper := getTestFileMetadata("/remote/doc.paper", uint64(len(exported)))
	paper.ExportInfo = &files.ExportInfo{ExportAs: "markdown"}
	entries := []files.IsMetadata{
		getTestFolderMetadata("/remote"),
		paper,
		getTestFileMetadata("/remote/a.txt", 4),
	}
	mock := &mockFilesClient{
		getMetadataFn: func(arg *files.GetMetadataArg) (files.IsMetadata, error) {
			return getTestFolderMetadata(arg.Path), nil
		},
		listFolderFn: func(arg *files.ListFolderArg) (*files.ListFolderResult, error) {
			return &files.ListFolderResult{Entries: entries}, nil
		},
		downloadFn: func(arg *files.DownloadArg) (*files.FileMetadata, io.ReadCloser, error) {
			return getTestFileMetadata(arg.Path, 4), io.NopCloser(strings.NewReader("data")), nil
		},
		exportFn: func(arg *files.ExportArg) (*files.ExportResult, io.ReadCloser, error) {
			meta := getTestFileMetadata(arg.Path, uint64(len(exported)))
			return &files.ExportResult{
				ExportMetadata: &files.ExportMetadata{Name: "doc.md", Size: uint64(len(exported))},
				FileMetadata:   meta,
			}, io.NopCloser(strings.NewReader(exported)), nil
		},
	}
	stubFilesClient(t, mock)

	cmd := testGetWorkersCmd(2)
	cmd.SetErr(io.Discard)
	if err := get(cmd, []string{"/remote", dst}); err != nil {
		t.Fatalf("get error: %v", err)
	}
	got, err := os.ReadFile(filepath.Join(dst, "doc.md"))
	if err != nil {
		t.Fatalf("read exported file: %v", err)
	}
	if string(got) != exported {
		t.Fatalf("exported content = %q, want %q", got, exported)
	}
	if _, err := os.Stat(filepath.Join(dst, "a.txt")); err != nil {
		t.Fatalf("a.txt not downloaded: %v", err)
	}
}
