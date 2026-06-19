//go:build !(darwin || dragonfly || freebsd || illumos || linux || netbsd || openbsd || solaris)

package cmd

func ignoreBrokenPipeSignal() {}
