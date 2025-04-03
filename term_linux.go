//go:build linux

/*
 * Minimal terminal support code, so we don't have to bring in an external package
 */
package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"syscall"
	"unsafe"
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

func ReadInput(prompt string, noecho bool) (string, error) {

	if noecho {
		old, err := getTermState()
		if err != nil {
			return "", nil
		}

		new_ := old
		new_.Lflag &^= syscall.ECHO
		if err := setTermState(new_); err != nil {
			return "", nil
		}

		defer setTermState(old)
	}

	fmt.Printf("%s", prompt)
	reader := bufio.NewReader(os.Stdin)
	ret, err := reader.ReadString('\n')
	ret = strings.Trim(ret, "\n\r")
	if noecho {
		fmt.Println()
	}
	return ret, err
}
