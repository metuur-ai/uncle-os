//go:build !darwin && !linux

package main

import "os"

// isTerminal is conservatively false wherever the termios ioctl is not
// implemented. Erring towards "not a terminal" is the safe direction: the
// wizard refuses and names the flags to pass, which is the documented
// non-interactive contract, rather than blocking on a read nobody will answer.
func isTerminal(*os.File) bool { return false }
