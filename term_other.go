//go:build !linux

package main

import (
	"bufio"
	"fmt"
	"os"
)

// ReadInput reads user input (possibly password) from terminal.
// In this version password echo is not turned off.
//
//	We could use x/term package to handle that but we want to avoid any external package for now
func ReadInput(prompt string, ispassword bool) (string, error) {
	fmt.Printf("%s", prompt)
	reader := bufio.NewReader(os.Stdin)
	return reader.ReadString('\n')
}
