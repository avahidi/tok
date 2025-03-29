package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
)

const (
	DATABASE_VERSION   uint32 = 1
	PASSWORD_SALT_SIZE        = 32
)

var (
	DATABASE_MAGIC [4]byte = [4]byte{'t', 'o', 'k', 'B'}
)

// databaseHeader represents the file format database header
type databaseHeader struct {
	Magic        [4]byte
	Version      uint32
	PasswordSalt [PASSWORD_SALT_SIZE]byte
	Length       uint32
}

// Database contains all tokens + some other information
type Database struct {
	Entries   []*Entry
	filename  string
	pass      []byte
	pass_salt []byte
}

// CreateDatabase create a new database for the given filename and password,
// A random password salt is generated when this function is called
func CreateDatabase(filename, password string) (*Database, error) {
	db := &Database{
		filename:  filename,
		pass:      []byte(password),
		pass_salt: secureRandom(PASSWORD_SALT_SIZE),
	}
	return db, nil
}

// LoadDatabase loads a database from file
func LoadDatabase(filename, password string) (*Database, error) {
	db := &Database{
		filename: filename,
		pass:     []byte(password),
	}
	if err := db.load(); err != nil {
		return nil, err
	}
	return db, nil
}

// Add adds an entry to the database (but doesn't save it)
func (db *Database) Add(entry *Entry) error {
	if db.findExact(entry.Name) != nil {
		return fmt.Errorf("Item '%s' already exists, remove it first", entry.Name)
	}

	db.Entries = append(db.Entries, entry)
	return nil
}

// Delete removes an entry from the database (but doesn't save it)
func (db *Database) Delete(name string) *Entry {
	entries, err := db.Find(name)
	if err != nil || len(entries) != 1 {
		fmt.Printf("Could not select unique item '%s'\n", name)
	} else if len(entries) == 1 {
		e1 := entries[0]
		for i, e := range db.Entries {
			if e == e1 {
				db.Entries = append(db.Entries[:i], db.Entries[i+1:]...)
				return e
			}
		}
	}
	return nil
}

// Find will try to find an entry, either by index, exact name or fuzzy search
func (db Database) Find(name string) ([]*Entry, error) {
	if entry, err := db.findIndex(name); err != nil {
		return nil, err
	} else if entry != nil {
		return []*Entry{entry}, nil
	}
	if entry := db.findExact(name); entry != nil {
		return []*Entry{entry}, nil
	}

	return db.findFuzzy(name), nil
}

// findIndex will find entry from index in format "#<number"
func (db Database) findIndex(name string) (*Entry, error) {
	if name[0] != '#' {
		return nil, nil
	}
	idx, err := strconv.ParseInt(name[1:], 10, 32)
	if err != nil {
		return nil, fmt.Errorf("invalid number %s: %v", name, err)
	}
	n := int(idx - 1)
	if n < 0 || n >= len(db.Entries) {
		return nil, fmt.Errorf("Entry %s does not exist", name)
	}
	return db.Entries[n], nil
}

// findExact will search for a token with this exact name
func (db Database) findExact(name string) *Entry {
	name = strings.ToLower(name)
	// check for exact match first
	for _, e := range db.Entries {
		if strings.ToLower(e.Name) == name {
			return e
		}
	}
	return nil
}

// findFuzzy will find a token with similar name.
// Currently this is done by a simple substring search
func (db Database) findFuzzy(name string) []*Entry {
	name = strings.ToLower(name)

	var ret []*Entry
	for _, e := range db.Entries {
		if strings.Contains(strings.ToLower(e.Name), name) {
			ret = append(ret, e)
		}
	}
	return ret
}

// load will attempt to load the database from file
func (db *Database) load() error {
	r1, err := os.Open(db.filename)
	if err != nil {
		return err
	}
	defer r1.Close()

	// 1. load the header
	var hdr databaseHeader
	if err := binary.Read(r1, BYTE_ORDER, &hdr); err != nil {
		return err
	}
	if hdr.Magic != DATABASE_MAGIC {
		return fmt.Errorf("Invalid database file")
	}
	if hdr.Version != DATABASE_VERSION {
		return fmt.Errorf("Invalid database version: %d", hdr.Version)
	}

	// 2. read the encrypted data
	enc, err := ReadExact(r1, int(hdr.Length))
	if err != nil {
		return err
	}

	// 3. decrpyt data
	key := GenerateKeyFromPassword(db.pass, hdr.PasswordSalt[:])

	dec, err := decryptBytes(key, enc)
	if err != nil {
		// the reason we don't return and error and instead terminate is that if password
		// was incorrect, we may end up creating a new one (with the bad password) and
		// overwrite the existing database
		log.Fatalf("Unable to decrypt database. Check your password!\n")
	}

	// 4. read the items from the decrypted buffer, start with number of enteries
	r2 := bytes.NewBuffer(dec)

	var count uint32
	if err := binary.Read(r2, BYTE_ORDER, &count); err != nil {
		return err
	}

	if count > 1000 {
		return fmt.Errorf("Too many items in database: %d", count)
	}

	// 5.a load each entry
	var entries []*Entry
	for i := 0; i < int(count); i++ {
		entry := &Entry{}
		if err := entry.Deserial(r2); err != nil {
			return err
		}
		entries = append(entries, entry)
	}

	// all data was loaded, update the database
	db.pass_salt = hdr.PasswordSalt[:]
	db.Entries = entries

	return nil
}

// Save will write the database to file
func (db Database) Save() error {
	key := GenerateKeyFromPassword(db.pass, db.pass_salt)

	// 1. write items to a plaintext buffer, start with item count
	plain := new(bytes.Buffer)
	if err := binary.Write(plain, BYTE_ORDER, uint32(len(db.Entries))); err != nil {
		return err
	}

	for _, entry := range db.Entries {
		if err := entry.Serial(plain); err != nil {
			return err
		}
	}

	// 2. encrypt the entire buffer
	enc, err := encryptBytes(key, plain.Bytes())
	if err != nil {
		return err
	}

	// 3. open file and write the header to it
	f, err := os.OpenFile(db.filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	defer f.Close()

	// write header to file
	hdr := databaseHeader{
		Magic:   DATABASE_MAGIC,
		Version: DATABASE_VERSION,
		Length:  uint32(len(enc)),
	}
	// ah golang...
	copy(hdr.PasswordSalt[:], db.pass_salt)

	if err := binary.Write(f, BYTE_ORDER, &hdr); err != nil {
		return err
	}

	// 4. add the encrypted data
	return WriteExact(f, enc)
}
