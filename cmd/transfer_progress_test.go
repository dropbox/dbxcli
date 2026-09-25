package cmd

import (
	"bytes"
	"strings"
	"testing"
	"time"
)

func TestFormatTransferProgress(t *testing.T) {
	got := formatTransferProgress(45<<20, 100<<20, 12*time.Second, 3.7*float64(1<<20), "")
	want := " 45%|=========           | 45 MiB/100 MiB [00:12<00:15, 3.7 MiB/s]"
	if got != want {
		t.Fatalf("progress = %q, want %q", got, want)
	}

	got = formatTransferProgress(2700<<20, 2700<<20, 90*time.Second, 30<<20, ", 6 workers")
	want = "100%|====================| 2.6 GiB/2.6 GiB [01:30<00:00, 30 MiB/s, 6 workers]"
	if got != want {
		t.Fatalf("complete progress = %q, want %q", got, want)
	}

	got = formatTransferProgress(0, 1<<30, 0, 0, "")
	want = "  0%|                    | 0 B/1.0 GiB [00:00<?, 0 B/s]"
	if got != want {
		t.Fatalf("initial progress = %q, want %q", got, want)
	}
}

func TestFormatTransferProgressWithoutTotal(t *testing.T) {
	got := formatTransferProgress(5<<20, 0, 3*time.Second, 1<<20, "")
	want := "5.0 MiB [00:03, 1.0 MiB/s]"
	if got != want {
		t.Fatalf("progress = %q, want %q", got, want)
	}
}

func TestFormatTransferClock(t *testing.T) {
	cases := map[time.Duration]string{
		0:                                     "00:00",
		59 * time.Second:                      "00:59",
		61 * time.Second:                      "01:01",
		3599 * time.Second:                    "59:59",
		time.Hour:                             "1:00:00",
		26*time.Hour + 5*time.Minute:          "26:05:00",
		1500 * time.Millisecond:               "00:02",
		-5 * time.Second:                      "00:00",
		12*time.Second + 400*time.Millisecond: "00:12",
	}
	for d, want := range cases {
		if got := formatTransferClock(d); got != want {
			t.Errorf("formatTransferClock(%v) = %q, want %q", d, got, want)
		}
	}
}

func TestFormatTransferETA(t *testing.T) {
	if got := formatTransferETA(0, 100); got != "00:00" {
		t.Fatalf("finished ETA = %q, want 00:00", got)
	}
	if got := formatTransferETA(100, 0); got != "?" {
		t.Fatalf("unknown ETA = %q, want ?", got)
	}
	if got := formatTransferETA(1500, 100); got != "00:15" {
		t.Fatalf("ETA = %q, want 00:15", got)
	}
	if got := formatTransferETA(1<<40, 1); got != ">99:59:59" {
		t.Fatalf("huge ETA = %q, want cap", got)
	}
}

func TestTransferRateMeterUsesSlidingWindow(t *testing.T) {
	start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	m := newTransferRateMeter(start)
	if m.rate() != 0 {
		t.Fatalf("initial rate = %v, want 0", m.rate())
	}

	// 1 MiB/s for the first ten seconds.
	for i := 1; i <= 10; i++ {
		m.update(start.Add(time.Duration(i)*time.Second), int64(i)<<20)
	}
	if got := m.rate(); got != float64(1<<20) {
		t.Fatalf("steady rate = %v, want %v", got, 1<<20)
	}

	// Then 10 MiB/s: after the window has rolled over, only the recent
	// speed should be reported, not the average since the start.
	bytes := int64(10 << 20)
	for i := 11; i <= 20; i++ {
		bytes += 10 << 20
		m.update(start.Add(time.Duration(i)*time.Second), bytes)
	}
	if got := m.rate(); got != float64(10<<20) {
		t.Fatalf("recent rate = %v, want %v", got, 10<<20)
	}
	if got := m.elapsed(start.Add(20 * time.Second)); got != 20*time.Second {
		t.Fatalf("elapsed = %v, want 20s", got)
	}
	if len(m.samples) > 8 {
		t.Fatalf("meter kept %d samples, want the window only", len(m.samples))
	}
}

func TestTransferRateMeterCoalescesRapidUpdates(t *testing.T) {
	start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	m := newTransferRateMeter(start)
	for i := 1; i <= 1000; i++ {
		m.update(start.Add(time.Duration(i)*time.Millisecond), int64(i)*1024)
	}
	if len(m.samples) > 3 {
		t.Fatalf("meter kept %d samples for sub-interval updates, want coalesced", len(m.samples))
	}
	if got := m.rate(); got != float64(1024*1000) {
		t.Fatalf("rate = %v, want %v", got, 1024*1000)
	}
}

func TestTransferProgressDrawerThrottlesAndReportsRate(t *testing.T) {
	var buf bytes.Buffer
	d := newTransferProgressDrawer(&buf, "Downloading ")
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	d.now = func() time.Time { return now }
	d.meter = newTransferRateMeter(now)

	total := int64(100 << 20)
	d.update(1<<20, total) // first update always draws
	now = now.Add(10 * time.Millisecond)
	d.update(2<<20, total) // too soon: skipped
	now = now.Add(time.Second)
	d.update(10<<20, total) // ~1s elapsed: drawn with rate and ETA
	now = now.Add(time.Second)
	d.update(total, total) // completion always draws
	d.finish()
	d.finish() // idempotent

	out := buf.String()
	frames := strings.Split(strings.TrimSuffix(out, "\n"), "\r")
	if len(frames) != 4 || frames[3] != "" {
		t.Fatalf("frames = %q, want three drawn frames and a trailing newline", frames)
	}
	if !strings.HasPrefix(frames[0], "Downloading   1%|") {
		t.Fatalf("first frame = %q", frames[0])
	}
	if want := "Downloading  10%|==                  | 10 MiB/100 MiB [00:01<00:09, 9.9 MiB/s]"; !strings.HasPrefix(frames[1], want) {
		t.Fatalf("second frame = %q, want prefix %q", frames[1], want)
	}
	if want := "Downloading 100%|====================| 100 MiB/100 MiB [00:02<00:00, 50 MiB/s]"; !strings.HasPrefix(frames[2], want) {
		t.Fatalf("final frame = %q, want prefix %q", frames[2], want)
	}
	if strings.Count(out, "\n") != 1 {
		t.Fatalf("output = %q, want exactly one newline from finish", out)
	}
}

func TestTransferProgressDrawerDrawFuncHandlesReaderProtocol(t *testing.T) {
	var buf bytes.Buffer
	d := newTransferProgressDrawer(&buf, "Downloading ")
	draw := d.drawFunc()
	if err := draw(512, 1024); err != nil {
		t.Fatal(err)
	}
	if err := draw(-1, -1); err != nil {
		t.Fatal(err)
	}
	if err := draw(1024, 1024); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	if !strings.Contains(out, "Downloading  50%|==========          | 512 B/1.0 KiB [") {
		t.Fatalf("output = %q, want a half-way frame", out)
	}
	if !strings.HasSuffix(out, "\n") || strings.Contains(out, "1.0 KiB/1.0 KiB") {
		t.Fatalf("output = %q, want nothing drawn after finish", out)
	}
}
