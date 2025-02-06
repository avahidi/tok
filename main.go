package main

import (
	"fmt"
	"log"
	"os"
)

func showEntry(entry *Entry) error {
	totp, err := entry.Totp()
	if err != nil {
		return err
	}

	fmt.Printf("%s - %s  \n", entry.Name, totp.Generate(6))
	return nil
}

func cmdAdd(name, secret string) error {
	_, err := secretFromBase64(secret)
	if err != nil {
		return err
	}

	db, err := NewDatabase()
	if err != nil {
		return err
	}

	entry := db.Add(name, secret)
	if err := db.Save(); err != nil {
		return err
	}

	return showEntry(entry)
}

func cmdSearch(name string) error {
	db, err := NewDatabase()
	if err != nil {
		return err
	}

	entry := db.Find(name)
	if entry == nil {
		return fmt.Errorf("unable to find '%s'", name)
	}

	return showEntry(entry)
}

func cmdLs() error {
	db, err := NewDatabase()
	if err != nil {
		return err
	}
	for i, entry := range db.Entries {
		fmt.Printf("%03d %v\n", i+1, entry)
	}
	return err
}

func usage() {
	name := os.Args[0]
	fmt.Fprintf(os.Stderr, "Usage:\n"+
		"    %s add <NAME> <KEY>\n"+
		"    %s ls\n"+
		"    %s show <NAME>\n",
		name, name, name,
	)
}

func main() {
	args := os.Args[1:]

	if len(args) == 3 && args[0] == "add" {
		if err := cmdAdd(args[1], args[2]); err != nil {
			log.Fatalf("Failed to add new entry: %v\n", err)
		}

	} else if len(args) == 1 && args[0] == "ls" {
		if err := cmdLs(); err != nil {
			log.Fatalf("Failed to list database: %v\n", err)
		}
	} else if len(args) == 1 {
		if err := cmdSearch(args[0]); err != nil {
			log.Fatalf("Failed to search database: %v\n", err)
		}
	} else {
		usage()
	}

	/*
	   encodedKey := "JBSW Y3DP EHPK 3PXP"
	   encodedKey = "ASDASDFF"

	   secret, err := secretFromBase64(encodedKey)

	   	if err != nil {
	   		log.Fatalf("Failed to decode secret: %v\n", err)
	   	}

	   totp := NewTotp(secret)

	   fmt.Printf("Token=%06d\n", totp.Generate(6))
	*/
}
