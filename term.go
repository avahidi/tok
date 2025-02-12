/*
 * Minimal terminal support code, so we don't have to bring in an external package
 */
package main

import (
	"fmt"
	"os"
	"syscall"
	"unsafe"
)

// see https://en.wikipedia.org/wiki/ANSI_escape_code
const (
	TERM_RESET  = 0
	TERM_NORMAL = 0
	TERM_BOLD   = 1
	TERM_FAINT  = 2
	TERM_FG     = 30
	TERM_BG     = 40

	TERM_BLACK   = 0
	TERM_RED     = 1
	TERM_GREEN   = 2
	TERM_YELLOW  = 3
	TERM_BLUE    = 4
	TERM_WHITE   = 7
	TERM_DEFAULT = 9
)

func getTermState() (syscall.Termios, error) {
	var save syscall.Termios

	if _, _, err := syscall.Syscall6(
		syscall.SYS_IOCTL,
		uintptr(os.Stdin.Fd()),
		uintptr(syscall.TCGETS),
		uintptr(unsafe.Pointer(&save)),
		0, 0, 0); err != 0 {
		return save, fmt.Errorf("TCGETS failed: %d", err)
	}
	return save, nil
}

func setTermState(s syscall.Termios) error {
	if _, _, err := syscall.Syscall6(
		syscall.SYS_IOCTL,
		uintptr(os.Stdin.Fd()),
		uintptr(syscall.TCSETS),
		uintptr(unsafe.Pointer(&s)),
		0, 0, 0); err != 0 {
		return fmt.Errorf("TCSETS failed: %d", err)
	}
	return nil
}

func ReadPassword(prompt string) (string, error) {
	old, err := getTermState()
	if err != nil {
		return "", nil
	}

	new_ := old
	new_.Lflag &^= syscall.ECHO
	if err := setTermState(new_); err != nil {
		return "", nil
	}

	fmt.Printf("%s", prompt)

	var ret string
	fmt.Scanln(&ret)
	fmt.Println()

	if err := setTermState(old); err != nil {
		return "", nil
	}

	return ret, nil
}

func TextControl(code int) string {
	return fmt.Sprintf("\033[%dm", code)
}

func TextWrap(start, end int, text string) string {
	return fmt.Sprintf("%s%s%s", TextControl(start), text, TextControl(end))
}
