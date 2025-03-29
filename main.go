package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path"
)

const (
	DATABASE_FILENAME = ".tokdb"
	DEFAULT_HASH      = "sha1"
	DEFAULT_DIGITS    = 6
	DEFAULT_PERIOD    = 30
	DEFAULT_TIME      = 30
)

// Config is the current application configuration
// Most of this comes from command-line options
type Config struct {
	DatabaseFilename string
	Period           int
	Digits           int
	HashAlgorithm    string
	password         string
	Verbose          bool
	Time             int
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
	fmt.Fprintf(out, "COMMANDS:\n"+
		"    add <NAME> <KEY> [NOTE]\n"+
		"    add otpauth://totp/...\n"+
		"    rm <name>\n"+
		"    ls\n"+
		"    show <NAME>\n"+
		"    show \\#<NUMBER>\n"+
		"    <NAME> (same as show <NAME>)\n",
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
	time := flag.Int("time", DEFAULT_TIME, "display time")
	period := flag.Int("period", DEFAULT_PERIOD, "token period")
	digits := flag.Int("digits", DEFAULT_DIGITS, "token digits")
	hashname := flag.String("hash", DEFAULT_HASH, "TOTP hash algorithm")
	verbose := flag.Bool("v", false, "verbose output")

	flag.Usage = usage
	flag.Parse()

	args := flag.Args()
	if len(args) == 0 {
		flag.Usage()
		os.Exit(20)
	}

	cfg := &Config{
		DatabaseFilename: *filename,
		Period:           *period,
		Digits:           *digits,
		HashAlgorithm:    *hashname,
		Verbose:          *verbose,
		Time:             *time,
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

func addEntry(cfg *Config, entry *Entry) error {
	db, err := getDatabase(cfg, true)
	if err != nil {
		return err
	}

	if err := db.Add(entry); err != nil {
		return err
	}

	db.Add(entry)
	if err := db.Save(); err != nil {
		return err
	}
	return showEntry(cfg.Time, entry)
}

func cmdAdd(cfg *Config, name, secret, note string) error {
	entry, err := NewEntry(name, secret, cfg.HashAlgorithm, note, cfg.Period, cfg.Digits)
	if err != nil {
		return err
	}
	return addEntry(cfg, entry)
}

func cmdAddUri(cfg *Config, uristr string) error {
	entry, err := EntryFromUri(uristr)
	if err != nil {
		return err
	}
	return addEntry(cfg, entry)
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

	fmt.Printf("Removed %s\n", e.Name)
	return db.Save()
}

func cmdSearch(cfg *Config, name string) error {
	db, err := getDatabase(cfg, false)
	if err != nil {
		log.Fatalf("Internal error: %v\n", err)
	}

	entries, err := db.Find(name)
	if err != nil {
		return err
	}

	switch len(entries) {
	case 0:
		return fmt.Errorf("unable to find '%s'", name)
	case 1:
		return showEntry(cfg.Time, entries[0])
	default:
		showEntries(false, entries)
		return fmt.Errorf("multiple items matching '%s'", name)
	}
}

func cmdList(cfg *Config) error {
	db, err := getDatabase(cfg, false)
	if err != nil {
		log.Fatalf("Internal error: %v\n", err)
	}

	showEntries(cfg.Verbose, db.Entries)
	return nil
}

func main() {
	cfg, args := parseParams()
	cmd, params := args[0], args[1:]

	// check if we received the correct number of parameters
	n := len(params)
	if (cmd == "add" && (n < 1 || n > 3)) ||
		(cmd == "rm" && n != 1) ||
		(cmd == "ls" && n != 0) ||
		(cmd == "show" && n != 1) {
		usage()
		os.Exit(20)
	}

	var err error
	switch cmd {
	case "add":
		if len(params) == 1 {
			err = cmdAddUri(cfg, params[0])
		} else {
			params = append(params, "") // add empty note if it was not provided
			err = cmdAdd(cfg, params[0], params[1], params[2])
		}
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
