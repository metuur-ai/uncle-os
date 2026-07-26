//go:build darwin || linux

package main

import (
	"os"
	"syscall"
	"unsafe"
)

// isTerminal is Python's sys.stdin.isatty() (bin/company-os:1960), which
// `init`'s wizard branches on (GPF-R-1.3).
//
// It has to be a real ioctl. The obvious stdlib approximation — Stat() and a
// test for os.ModeCharDevice — is WRONG, and wrong in exactly the case that
// matters: /dev/null is a character device too, so a CI run redirecting stdin
// from /dev/null would be treated as interactive, print "Company name
// [My Company]: " to stdout and then fail on EOF, where the oracle refuses
// immediately and names the missing flag. Measured against the 0.3 differential
// harness, which runs every invocation with stdin closed.
//
// The termios read is what isatty(3) itself does. Only its success matters, so
// the buffer is opaque: 128 bytes covers struct termios on both platforms
// (72 on darwin, 60 on linux).
func isTerminal(f *os.File) bool {
	var termios [128]byte
	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, f.Fd(),
		ioctlReadTermios, uintptr(unsafe.Pointer(&termios[0])))
	return errno == 0
}
