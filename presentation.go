package main

import (
	"fmt"
	"log"
	"strings"
	"time"
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

func TextControl(code int) string {
	return fmt.Sprintf("\033[%dm", code)
}

// codeWithProgress creates a progress bar with the kod
func codeWithProgress(kod string, curr, max int) string {
	// add some space in the middle and make it a bit prettier...
	if len(kod) > 4 {
		mid := len(kod) / 2
		kod = fmt.Sprintf("%s %s", kod[:mid], kod[mid:])
	}

	color := TERM_BG + TERM_GREEN
	if curr*8 > max*7 {
		color = TERM_BG + TERM_RED
	} else if curr*8 > max*6 {
		color = TERM_BG + TERM_YELLOW
	}

	const WIDTH = 20
	n := (WIDTH * 2 * (curr + 1)) / (max * 2) // *2 and +1 is to add 0.5 to curr
	return fmt.Sprintf(" [%s%s%s%s] - %s",
		TextControl(color),
		strings.Repeat("=", n),
		TextControl(TERM_RESET),
		strings.Repeat(" ", WIDTH-n),
		kod,
	)
}

// showEntry will try to show the token in a somewhat readable way
func showEntry(tim int, entry *Entry) error {
	totp, err := entry.Totp()
	if err != nil {
		return err
	}

	fmt.Printf("\nToken %s'%s'%s, added %s:\n",
		TextControl(TERM_BOLD), entry.Name, TextControl(TERM_NORMAL), entry.Date())
	if entry.Note != "" {
		fmt.Printf("%s\n", entry.Note)
	}
	fmt.Println()

	var line string
	for i := 0; i < tim; i++ {
		timeleft, kod := totp.Generate()
		line = codeWithProgress(kod, int(totp.Period)-timeleft, int(totp.Period))
		fmt.Printf("%s \r", line)
		time.Sleep(time.Second)
	}

	// at exit write over visible text.
	// XXX: this will not run if user presses Ctrl-C
	fmt.Printf("\r%s  \n", strings.Repeat("*", len(line)+2))
	return nil
}

// showEntries lists a set of entries but does not show the token
func showEntries(asuri, verbose bool, entries []*Entry) {
	for i, entry := range entries {
		if asuri {
			str, err := EntryToUri(entry)
			if err != nil {
				log.Fatalf("Unable to convert entry to uri: '%v'", err)
			}
			fmt.Printf("%3d - %s\n", i+1, str)

		} else if verbose {
			fmt.Printf("%3d - %s\n", i+1, entry.Name)
			if entry.Note != "" {
				fmt.Printf("\tNote: %s\n", entry.Note)
			}
			fmt.Printf("\tDate added: %s\n", entry.Date())
			fmt.Printf("\tPeriod: %d\n", entry.Period)
			fmt.Printf("\tDigits: %d\n", entry.Digits)
			fmt.Printf("\tHash: %s\n", hashToName(entry.Hash))
			fmt.Printf("\n")
		} else {
			fmt.Printf("%3d - %s\n", i+1, entry.Name)
		}
	}
}
