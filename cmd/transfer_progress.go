package cmd

import (
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/mitchellh/ioprogress"
)

const (
	// transferRateWindow is how far back the rate meter looks when computing
	// the current throughput, so the figure tracks recent speed rather than
	// the average since the start.
	transferRateWindow = 5 * time.Second

	// transferRateSampleSpacing coalesces updates that arrive closer together
	// than this so the sample window stays small.
	transferRateSampleSpacing = 100 * time.Millisecond

	// transferProgressBarWidth is the width of the ASCII progress bar.
	transferProgressBarWidth = 20
)

// transferProgressDrawInterval throttles progress redraws so a fast transfer
// does not flood stderr.
var transferProgressDrawInterval = 100 * time.Millisecond

type transferRateSample struct {
	at    time.Time
	bytes int64
}

// transferRateMeter estimates throughput from cumulative byte counts using a
// sliding window, the way tqdm smooths its rate display.
type transferRateMeter struct {
	start   time.Time
	window  time.Duration
	samples []transferRateSample
}

func newTransferRateMeter(start time.Time) *transferRateMeter {
	return &transferRateMeter{
		start:   start,
		window:  transferRateWindow,
		samples: []transferRateSample{{at: start}},
	}
}

// update records the cumulative number of bytes transferred as of now.
func (m *transferRateMeter) update(now time.Time, bytes int64) {
	last := m.samples[len(m.samples)-1]
	if len(m.samples) > 1 && now.Sub(last.at) < transferRateSampleSpacing {
		m.samples[len(m.samples)-1] = transferRateSample{at: now, bytes: bytes}
	} else {
		m.samples = append(m.samples, transferRateSample{at: now, bytes: bytes})
	}

	// Drop samples that fell out of the window, keeping one older sample as
	// the anchor so the window always spans at least its full width.
	cutoff := now.Add(-m.window)
	for len(m.samples) > 2 && !m.samples[1].at.After(cutoff) {
		m.samples = m.samples[1:]
	}
}

// rate returns the recent throughput in bytes per second.
func (m *transferRateMeter) rate() float64 {
	first := m.samples[0]
	last := m.samples[len(m.samples)-1]
	seconds := last.at.Sub(first.at).Seconds()
	if seconds <= 0 || last.bytes <= first.bytes {
		return 0
	}
	return float64(last.bytes-first.bytes) / seconds
}

// elapsed returns the time since the transfer started.
func (m *transferRateMeter) elapsed(now time.Time) time.Duration {
	return now.Sub(m.start)
}

// formatTransferProgress renders a tqdm-style progress line:
//
//	45%|=========           | 45 MiB/100 MiB [00:12<00:15, 3.7 MiB/s]
//
// extra is appended inside the brackets, for example ", 6 workers". When the
// total is unknown the percentage, bar, and ETA are omitted.
func formatTransferProgress(done, total int64, elapsed time.Duration, rate float64, extra string) string {
	if done < 0 {
		done = 0
	}
	if total <= 0 {
		return fmt.Sprintf("%s [%s, %s%s]",
			humanize.IBytes(uint64(done)), formatTransferClock(elapsed), formatTransferRate(rate), extra)
	}
	if done > total {
		done = total
	}

	percent := int(done * 100 / total)
	filled := int(done * transferProgressBarWidth / total)
	bar := strings.Repeat("=", filled) + strings.Repeat(" ", transferProgressBarWidth-filled)

	return fmt.Sprintf("%3d%%|%s| %s/%s [%s<%s, %s%s]",
		percent, bar,
		humanize.IBytes(uint64(done)), humanize.IBytes(uint64(total)),
		formatTransferClock(elapsed), formatTransferETA(total-done, rate),
		formatTransferRate(rate), extra)
}

func formatTransferRate(rate float64) string {
	if rate < 0 {
		rate = 0
	}
	return humanize.IBytes(uint64(rate)) + "/s"
}

// formatTransferETA estimates the remaining time from the bytes left and the
// current rate, or "?" when no rate is known yet.
func formatTransferETA(remaining int64, rate float64) string {
	if remaining <= 0 {
		return formatTransferClock(0)
	}
	if rate <= 0 {
		return "?"
	}
	seconds := float64(remaining) / rate
	if seconds > 359999 { // cap at 99:59:59 to keep the line tidy
		return ">99:59:59"
	}
	return formatTransferClock(time.Duration(seconds * float64(time.Second)))
}

// formatTransferClock renders a duration as MM:SS, or H:MM:SS past an hour.
func formatTransferClock(d time.Duration) string {
	if d < 0 {
		d = 0
	}
	total := int64(d.Round(time.Second) / time.Second)
	hours, minutes, seconds := total/3600, (total%3600)/60, total%60
	if hours > 0 {
		return fmt.Sprintf("%d:%02d:%02d", hours, minutes, seconds)
	}
	return fmt.Sprintf("%02d:%02d", minutes, seconds)
}

// transferProgressDrawer draws a single-transfer progress line with throughput
// and ETA, throttled to transferProgressDrawInterval.
type transferProgressDrawer struct {
	prefix   string
	draw     ioprogress.DrawFunc
	meter    *transferRateMeter
	now      func() time.Time
	lastDraw time.Time
	finished bool
}

func newTransferProgressDrawer(w io.Writer, prefix string) *transferProgressDrawer {
	if w == nil {
		w = io.Discard
	}
	d := &transferProgressDrawer{prefix: prefix, now: time.Now}
	d.meter = newTransferRateMeter(d.now())
	d.draw = ioprogress.DrawTerminalf(w, func(progress, total int64) string {
		return d.line(progress, total)
	})
	return d
}

func (d *transferProgressDrawer) line(progress, total int64) string {
	now := d.now()
	return d.prefix + formatTransferProgress(progress, total, d.meter.elapsed(now), d.meter.rate(), "")
}

// update records progress and redraws the line if enough time has passed or
// the transfer is complete.
func (d *transferProgressDrawer) update(progress, total int64) {
	if d.finished {
		return
	}
	now := d.now()
	d.meter.update(now, progress)
	complete := total > 0 && progress >= total
	if !complete && !d.lastDraw.IsZero() && now.Sub(d.lastDraw) < transferProgressDrawInterval {
		return
	}
	d.lastDraw = now
	_ = d.draw(progress, total)
}

// finish ends the progress line. It is safe to call more than once.
func (d *transferProgressDrawer) finish() {
	if d.finished {
		return
	}
	d.finished = true
	_ = d.draw(-1, -1)
}

// drawFunc adapts the drawer to ioprogress.Reader, which performs its own
// throttling and signals completion with (-1, -1).
func (d *transferProgressDrawer) drawFunc() ioprogress.DrawFunc {
	return func(progress, total int64) error {
		if progress == -1 && total == -1 {
			d.finish()
			return nil
		}
		if d.finished {
			return nil
		}
		now := d.now()
		d.meter.update(now, progress)
		d.lastDraw = now
		return d.draw(progress, total)
	}
}
