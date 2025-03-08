//go:build !linux

package main

import (
	"fmt"
)

// ReadPassword reads user password from terminal. In this version echo is not turned off.
//
//	We could use x/term package to handle that but we want to avoid any external package for now
func ReadPassword(prompt string) (string, error) {
	fmt.Printf("%s", prompt)

	var ret string
	fmt.Scanln(&ret)
	fmt.Println()
	return ret, nil
}
