package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"strings"
	"time"
)

type Entry struct {
	Added  time.Time
	Name   string
	Secret string
}

func NewEntry(name, secret string) *Entry {
	return &Entry{
		Added:  time.Now(),
		Name:   name,
		Secret: secret,
	}
}

func (e Entry) String() string {
	return fmt.Sprintf("(%s) %s", e.Date(), e.Name)
}
func (e Entry) Totp() (*Totp, error) {
	secret, err := secretFromBase64(e.Secret)
	if err != nil {
		return nil, err
	}
	return NewTotp(secret), nil
}

func (e Entry) Date() string {
	return e.Added.Format("2006-01-02 15:04:05")
}

type Database struct {
	filename string
	Entries  []*Entry
}

func NewDatabase() (*Database, error) {
	filename, err := defaultDatabaseFilename()
	if err != nil {
		return nil, err
	}

	db := &Database{filename: filename}
	db.Load()
	return db, nil
}

func (db *Database) Add(name, secret string) *Entry {
	entry := NewEntry(name, secret)

	db.Entries = append(db.Entries, entry)
	return entry
}

func (db *Database) Find(name string) *Entry {
	name = strings.ToLower(name)
	// check for exact match first
	for _, e := range db.Entries {
		if strings.ToLower(e.Name) == name {
			return e
		}
	}
	// check for partial match
	for _, e := range db.Entries {
		if strings.Contains(strings.ToLower(e.Name), name) {
			return e
		}
	}
	return nil
}

func (db *Database) Load() error {
	data, err := os.ReadFile(db.filename)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, db)
}

func (db Database) Save() error {
	data, err := json.Marshal(&db)
	if err != nil {
		return err
	}
	return os.WriteFile(db.filename, data, 0700)
}

// the database is saved at a default location, this func will return it
func defaultDatabaseFilename() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	return path.Join(home, ".tok.db"), nil
}
