package cmd

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/dustin/go-humanize"
	"golang.org/x/term"
)

const (
	// downloadWorkersAuto selects the number of concurrent file downloads
	// automatically from the measured download throughput.
	downloadWorkersAuto = 0

	// autoDownloadWorkersInitial is the concurrency the automatic tuner starts
	// from before it has measured any throughput.
	autoDownloadWorkersInitial = 4

	// autoDownloadWorkersMax caps the concurrency the automatic tuner may reach.
	// Dropbox rate limits make very high connection counts counterproductive.
	autoDownloadWorkersMax = 16

	// autoTuneGainThreshold is the relative throughput improvement a higher
	// concurrency level must deliver to be kept. Smaller gains are treated as
	// noise, which means the link is saturated and the tuner settles.
	autoTuneGainThreshold = 0.10
)

var (
	// downloadMonitorInterval controls how often the status line is redrawn
	// and throughput samples are collected.
	downloadMonitorInterval = 500 * time.Millisecond

	// downloadAutoTuneWindow is the sampling window used to measure throughput
	// for one concurrency level before deciding whether to change it.
	downloadAutoTuneWindow = 2 * time.Second
)

// concurrencyLimiter bounds the number of concurrently running download jobs.
// Unlike a channel semaphore, the limit can be raised or lowered while jobs
// are running; lowering it only delays new jobs and never interrupts running
// ones.
type concurrencyLimiter struct {
	mu      sync.Mutex
	cond    *sync.Cond
	limit   int
	active  int
	waiting int
	waited  bool
}

func newConcurrencyLimiter(limit int) *concurrencyLimiter {
	if limit < 1 {
		limit = 1
	}
	l := &concurrencyLimiter{limit: limit}
	l.cond = sync.NewCond(&l.mu)
	return l
}

// acquire blocks until a slot is available or ctx is done.
func (l *concurrencyLimiter) acquire(ctx context.Context) error {
	stop := context.AfterFunc(ctx, func() {
		l.mu.Lock()
		l.cond.Broadcast()
		l.mu.Unlock()
	})
	defer stop()

	l.mu.Lock()
	defer l.mu.Unlock()
	for l.active >= l.limit {
		if err := ctx.Err(); err != nil {
			return err
		}
		l.waited = true
		l.waiting++
		l.cond.Wait()
		l.waiting--
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	l.active++
	return nil
}

func (l *concurrencyLimiter) release() {
	l.mu.Lock()
	l.active--
	l.cond.Broadcast()
	l.mu.Unlock()
}

func (l *concurrencyLimiter) setLimit(limit int) {
	if limit < 1 {
		limit = 1
	}
	l.mu.Lock()
	l.limit = limit
	l.cond.Broadcast()
	l.mu.Unlock()
}

func (l *concurrencyLimiter) currentLimit() int {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.limit
}

// takeSaturated reports whether the limit was binding at any point since the
// previous call, meaning a job had to wait for a slot. The flag is reset.
func (l *concurrencyLimiter) takeSaturated() bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	saturated := l.waited || (l.waiting > 0 && l.active >= l.limit)
	l.waited = false
	return saturated
}

// adaptiveConcurrency is a hill-climbing controller that discovers how many
// concurrent downloads the available bandwidth can use. It raises the
// concurrency step by step while each step improves aggregate throughput by
// at least autoTuneGainThreshold, and settles on the last level that did as
// soon as a step stops paying off, which indicates the link is saturated.
type adaptiveConcurrency struct {
	limit     int
	max       int
	best      float64
	bestLimit int
	warmup    bool
	settled   bool
}

func newAdaptiveConcurrency(initial, max int) *adaptiveConcurrency {
	if max < 1 {
		max = 1
	}
	if initial < 1 {
		initial = 1
	}
	if initial > max {
		initial = max
	}
	return &adaptiveConcurrency{
		limit:     initial,
		max:       max,
		bestLimit: initial,
		settled:   initial >= max,
	}
}

// observe feeds one throughput sample (bytes per second) measured at the
// current limit and returns the limit to use for the next window. saturated
// reports whether the limit was actually binding during the window; samples
// taken while it was not are uninformative and leave the limit unchanged.
func (c *adaptiveConcurrency) observe(bytesPerSecond float64, saturated bool) int {
	if c.settled || !saturated || bytesPerSecond <= 0 {
		return c.limit
	}
	if c.warmup {
		// The first window after a change still contains transfers that
		// started under the old limit, so skip it.
		c.warmup = false
		return c.limit
	}

	if c.best == 0 || bytesPerSecond >= c.best*(1+autoTuneGainThreshold) {
		c.best = bytesPerSecond
		c.bestLimit = c.limit
		if c.limit >= c.max {
			c.settled = true
			return c.limit
		}
		c.limit = nextConcurrencyStep(c.limit, c.max)
		c.warmup = true
		return c.limit
	}

	// The last increase did not buy a meaningful gain: the link is saturated.
	c.limit = c.bestLimit
	c.settled = true
	return c.limit
}

func nextConcurrencyStep(limit, max int) int {
	next := limit + limit/2
	if next <= limit {
		next = limit + 1
	}
	if next > max {
		next = max
	}
	return next
}

// downloadStatusWriter serialises stderr output from concurrent downloads and
// keeps a single live status line at the bottom when stderr is a terminal.
type downloadStatusWriter struct {
	mu      sync.Mutex
	w       io.Writer
	live    bool
	lastLen int
}

func newDownloadStatusWriter(w io.Writer) *downloadStatusWriter {
	if w == nil {
		w = io.Discard
	}
	live := false
	if f, ok := w.(*os.File); ok {
		live = term.IsTerminal(int(f.Fd()))
	}
	return &downloadStatusWriter{w: w, live: live}
}

func (s *downloadStatusWriter) clearLineLocked() {
	if s.lastLen == 0 {
		return
	}
	_, _ = fmt.Fprint(s.w, "\r"+strings.Repeat(" ", s.lastLen)+"\r")
	s.lastLen = 0
}

// message prints a complete line, temporarily clearing the live status line.
func (s *downloadStatusWriter) message(format string, args ...any) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.clearLineLocked()
	_, _ = fmt.Fprintf(s.w, format, args...)
}

// status redraws the live status line. It is a no-op unless stderr is a
// terminal, so logs and pipes only receive complete lines.
func (s *downloadStatusWriter) status(line string) {
	if !s.live {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	// Pad to the widest line drawn so far so shorter updates leave no
	// stale characters behind, and remember that width for clearing.
	padded := line
	if len(padded) < s.lastLen {
		padded += strings.Repeat(" ", s.lastLen-len(padded))
	}
	s.lastLen = len(padded)
	_, _ = fmt.Fprint(s.w, "\r"+padded)
}

// finish clears the live status line, if any, and prints a final summary
// line in its place.
func (s *downloadStatusWriter) finish(summary string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.clearLineLocked()
	_, _ = fmt.Fprintln(s.w, summary)
}

// downloadPool runs file download jobs with bounded, optionally self-tuning,
// concurrency and tracks aggregate progress across all of them.
type downloadPool struct {
	limiter    *concurrencyLimiter
	tuner      *adaptiveConcurrency
	status     *downloadStatusWriter
	totalFiles int
	totalBytes int64
	bytes      atomic.Int64
	done       atomic.Int64
	peakLimit  int
	started    time.Time
	meter      *transferRateMeter
}

// newDownloadPool creates a pool for jobs files. workers selects a fixed
// concurrency, or downloadWorkersAuto to tune it from measured throughput.
func newDownloadPool(workers int, totalFiles int, totalBytes int64, status *downloadStatusWriter) *downloadPool {
	if status == nil {
		status = newDownloadStatusWriter(io.Discard)
	}
	p := &downloadPool{
		status:     status,
		totalFiles: totalFiles,
		totalBytes: totalBytes,
	}
	if workers == downloadWorkersAuto {
		p.tuner = newAdaptiveConcurrency(
			min(autoDownloadWorkersInitial, max(totalFiles, 1)),
			min(autoDownloadWorkersMax, max(totalFiles, 1)),
		)
		workers = p.tuner.limit
	}
	p.limiter = newConcurrencyLimiter(workers)
	p.peakLimit = workers
	return p
}

// concurrent reports whether more than one download may run at a time.
func (p *downloadPool) concurrent() bool {
	if p.tuner != nil {
		return p.tuner.max > 1
	}
	return p.limiter.currentLimit() > 1
}

// addBytes records newly committed download bytes for throughput tracking.
func (p *downloadPool) addBytes(n int64) {
	if n > 0 {
		p.bytes.Add(n)
	}
}

// fileProgress returns a progress callback for a single file that feeds only
// newly committed bytes into the pool's aggregate counter.
func (p *downloadPool) fileProgress() downloadProgressFunc {
	var last int64
	var mu sync.Mutex
	return func(committed, _ int64) {
		mu.Lock()
		defer mu.Unlock()
		if committed > last {
			p.addBytes(committed - last)
			last = committed
		}
	}
}

// run executes jobs in order, starting each as soon as a slot is free, and
// returns when all of them have finished. Jobs are never skipped: when ctx is
// cancelled the remaining jobs receive ctx's error through onError.
func (p *downloadPool) run(ctx context.Context, jobs []func(), onError func(index int, err error)) {
	p.started = time.Now()
	p.meter = newTransferRateMeter(p.started)

	// Sequential downloads keep their own per-file progress bar, so the
	// aggregate status line and tuner only run in concurrent mode.
	concurrent := p.concurrent()
	stop := make(chan struct{})
	var monitor sync.WaitGroup
	if concurrent {
		monitor.Add(1)
		go func() {
			defer monitor.Done()
			p.monitor(stop)
		}()
	}

	var wg sync.WaitGroup
	for i, job := range jobs {
		if err := p.limiter.acquire(ctx); err != nil {
			onError(i, err)
			p.done.Add(1)
			continue
		}
		wg.Add(1)
		go func() {
			defer wg.Done()
			defer p.limiter.release()
			defer p.done.Add(1)
			job()
		}()
	}
	wg.Wait()

	close(stop)
	monitor.Wait()
	if concurrent {
		p.status.finish(p.summary())
	}
}

func (p *downloadPool) monitor(stop <-chan struct{}) {
	ticker := time.NewTicker(downloadMonitorInterval)
	defer ticker.Stop()

	windowStart := p.started
	windowBytes := p.bytes.Load()
	for {
		select {
		case <-stop:
			return
		case now := <-ticker.C:
			total := p.bytes.Load()
			p.meter.update(now, total)
			if p.tuner != nil && now.Sub(windowStart) >= downloadAutoTuneWindow {
				elapsed := now.Sub(windowStart).Seconds()
				rate := float64(total-windowBytes) / elapsed
				limit := p.tuner.observe(rate, p.limiter.takeSaturated())
				p.limiter.setLimit(limit)
				if limit > p.peakLimit {
					p.peakLimit = limit
				}
				windowStart, windowBytes = now, total
			}
			p.status.status(p.statusLine(now, total))
		}
	}
}

// statusLine renders the live aggregate line, for example:
//
//	Downloading 12/40 files  45%|=========           | 1.2 GiB/2.7 GiB [00:12<00:15, 98 MiB/s, 6 workers]
func (p *downloadPool) statusLine(now time.Time, bytes int64) string {
	return fmt.Sprintf("Downloading %d/%d files %s",
		p.done.Load(), p.totalFiles,
		formatTransferProgress(bytes, p.totalBytes, p.meter.elapsed(now), p.meter.rate(),
			fmt.Sprintf(", %d workers", p.limiter.currentLimit())))
}

// summary renders the final line printed once every job has finished, with
// the average throughput over the whole run.
func (p *downloadPool) summary() string {
	elapsed := time.Since(p.started)
	bytes := max(p.bytes.Load(), 0)
	average := 0.0
	if seconds := elapsed.Seconds(); seconds > 0 {
		average = float64(bytes) / seconds
	}
	return fmt.Sprintf("Downloaded %d/%d files, %s in %s (%s), up to %d workers",
		p.done.Load(), p.totalFiles,
		humanize.IBytes(uint64(bytes)), formatTransferClock(elapsed), formatTransferRate(average),
		p.peakLimit)
}
