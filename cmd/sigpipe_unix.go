//go:build darwin || dragonfly || freebsd || illumos || linux || netbsd || openbsd || solaris

package cmd

import (
	"os/signal"
	"syscall"
)

func ignoreBrokenPipeSignal() {
	signal.Ignore(syscall.SIGPIPE)
}
