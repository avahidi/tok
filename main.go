package main

import (
	"crypto"
	"flag"
	"fmt"
	"log"
	"os"
	"path"
	"strings"
	"time"
)

// Config is the current application configuration
// Most of this comes from command-line options
type Config struct {
	DatabaseFilename string
	Period           int
	Digits           int
	HashAlgorithm    crypto.Hash
	password         string
}

// Password returns the database password.
// If empty it will try to read it from env variable or from terminal
func (c *Config) Password() (string, error) {
	if c.password == "" {
		c.password = os.Getenv("TOK_PASSWORD")
	}
	if c.password == "" {
		password, err := ReadPassword("Please enter database password: ")
		if err != nil {
			return "", err
		}
		c.password = password
	}
	return c.password, nil
}

func usage() {
	name := os.Args[0]
	out := flag.CommandLine.Output()
	fmt.Fprintf(out, "Usage:\n"+
		"    %s [OPTIONS] COMMAND\n", name)

	fmt.Fprintf(out, "OPTIONS:\n")
	flag.PrintDefaults()
	fmt.Fprintf(out, "COMMANDS::\n"+
		"    add <NAME> <KEY>\n"+
		"    rm <name>\n"+
		"    ls\n"+
		"    show <NAME>\n"+
		"    <name> (same as show <NAME>)\n",
	)
	fmt.Fprintf(out, "ENVIRONMENT:\n"+
		"    TOK_PASSWORD: database password, if you don't want to read it from command line\n",
	)
}

func parseParams() (*Config, []string) {
	// Get the default database filename
	home, err := os.UserHomeDir()
	if err != nil {
		log.Fatalf("You are homeless: %v\n", err)
	}
	defaultDbFile := path.Join(home, DATABASE_FILENAME)

	filename := flag.String("db", defaultDbFile, "the database")
	period := flag.Int("period", DEFAULT_PERIOD, "token period")
	digits := flag.Int("digits", DEFAULT_DIGITS, "token digits")
	hashName := flag.String("hash", DEFAULT_HASH, "totp hash algorithm")
	flag.Usage = usage
	flag.Parse()

	args := flag.Args()
	if len(args) == 0 {
		flag.Usage()
		os.Exit(20)
	}

	hash_, err := hashFromName(*hashName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to parse hash: %s\n", err)
		flag.Usage()
		os.Exit(20)
	}

	cfg := &Config{
		DatabaseFilename: *filename,
		Period:           *period,
		Digits:           *digits,
		HashAlgorithm:    hash_,
	}

	return cfg, args
}

func getDatabase(cfg *Config, allowCreate bool) (*Database, error) {
	password, err := cfg.Password()
	if err != nil {
		return nil, err
	}

	db, err := LoadDatabase(cfg.DatabaseFilename, password)
	if err != nil {
		if !allowCreate {
			return nil, err
		}
		log.Printf("Warning: could not load old database: %v\n", err)
		db, err = CreateDatabase(cfg.DatabaseFilename, password)
	}
	return db, err
}

// messageWithProgress creates a colored progress bar with the message in the middle,
// color changes to red when we are past 75% progress
func messageWithProgress(msg string, curr, max int) (string, int) {
	const WIDTH = 60
	colorFG, colorBG := TERM_FG+TERM_BLACK, TERM_BG+TERM_YELLOW
	s1 := (WIDTH - len(msg)) / 2
	s2 := WIDTH - s1 - len(msg)
	str := fmt.Sprintf("%s%s%s", strings.Repeat(" ", s1), msg, strings.Repeat(" ", s2))
	pos := curr * WIDTH / max

	if curr*4 > max*3 {
		colorFG = TERM_FG + TERM_WHITE
		colorBG = TERM_BG + TERM_RED
	}
	return fmt.Sprintf("%s%s%s%s%s%s",
		TextControl(colorFG), TextControl(colorBG),
		str[:pos],
		TextControl(TERM_BG+TERM_DEFAULT),
		str[pos:],
		TextControl(TERM_RESET),
	), WIDTH
}

// showEntry will try to show the token in a somewhat readable way
func showEntry(entry *Entry) error {
	totp, err := entry.Totp()
	if err != nil {
		return err
	}

	fmt.Printf("\nToken '%s', added %s:\n\n",
		TextWrap(TERM_BOLD, TERM_NORMAL, entry.Name),
		entry.Date())

	// TODO: change this to continue until Ctrl-C
	written := 0
	for i := 0; i < SHOWN_TIME; i++ {
		timeleft, kod := totp.Generate()
		line, size := messageWithProgress(kod, int(totp.Period)-timeleft, int(totp.Period))
		written = size
		fmt.Printf("  [%s] \r", line)
		time.Sleep(time.Second)
	}

	// at exit write over visible text.
	// XXX: this will not run if user presses Ctrl-C
	fmt.Printf("\r  [%s]  \n", strings.Repeat("*", written))
	return nil
}

func cmdAdd(cfg *Config, name, secret string) error {
	// first check if the secret is reasonable
	if _, err := secretFromBase64(secret); err != nil {
		return err
	}
	entry := NewEntry(name, secret, uint16(cfg.Period), uint8(cfg.Digits), cfg.HashAlgorithm)
	db, err := getDatabase(cfg, true)
	if err != nil {
		return err
	}

	if entry := db.FindExact(name); entry != nil {
		return fmt.Errorf("Item '%s' already exists, remove it first", name)
	}

	db.Add(entry)
	if err := db.Save(); err != nil {
		return err
	}

	return showEntry(entry)
}

func cmdRemove(cfg *Config, name string) error {
	db, err := getDatabase(cfg, false)
	if err != nil {
		log.Fatalf("Internal error: %v\n", err)
	}

	e := db.Delete(name)
	if e == nil {
		return fmt.Errorf("entry '%s' does not exist", name)
	}

	fmt.Printf("Removed %s\n", e)
	return db.Save()
}

func cmdSearch(cfg *Config, name string) error {
	db, err := getDatabase(cfg, false)
	if err != nil {
		log.Fatalf("Internal error: %v\n", err)
	}

	entry := db.FindExact(name)
	if entry != nil {
		return showEntry(entry)
	}

	entries := db.FindFuzzy(name)
	switch len(entries) {
	case 0:
		return fmt.Errorf("unable to find '%s'", name)
	case 1:
		return showEntry(entries[0])
	default:
		for i, e := range entries {
			fmt.Printf("\t%d\t%v\n", i+1, e)
		}
		return fmt.Errorf("multiple items matching '%s'", name)
	}
}

func cmdList(cfg *Config) error {
	db, err := getDatabase(cfg, false)
	if err != nil {
		log.Fatalf("Internal error: %v\n", err)
	}

	for i, entry := range db.Entries {
		fmt.Printf("%03d %v\n", i+1, entry)
	}
	return nil
}

func main() {
	cfg, args := parseParams()
	cmd, params := args[0], args[1:]

	// check if we received the correct number of parameters
	cmdParams := map[string]int{"add": 2, "rm": 1, "ls": 0, "show": 1}
	if count, ok := cmdParams[cmd]; (ok && count != len(params)) || (!ok && len(params) != 1) {
		usage()
		os.Exit(20)
	}

	var err error
	switch cmd {
	case "add":
		err = cmdAdd(cfg, params[0], params[1])
	case "rm":
		err = cmdRemove(cfg, params[0])
	case "ls":
		err = cmdList(cfg)
	case "show":
		err = cmdSearch(cfg, params[0])
	default:
		// also allow user to not use any command and just use the token name
		err = cmdSearch(cfg, cmd)
	}

	if err != nil {
		log.Fatalf("Failed: %v\n", err)
	}
}
