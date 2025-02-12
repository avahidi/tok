package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
	"strings"
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
func (db *Database) Add(entry *Entry) {
	db.Entries = append(db.Entries, entry)
}

// Delete removes an entry from the database (but doesn't save it)
func (db *Database) Delete(name string) *Entry {
	name = strings.ToLower(name)

	for i, e := range db.Entries {
		if strings.ToLower(e.Name) == name {
			db.Entries = append(db.Entries[:i], db.Entries[i+1:]...)
			return e
		}
	}
	return nil
}

// FindExact will search for a token with this exact name
func (db Database) FindExact(name string) *Entry {
	name = strings.ToLower(name)
	// check for exact match first
	for _, e := range db.Entries {
		if strings.ToLower(e.Name) == name {
			return e
		}
	}
	return nil
}

// FindFuzzy will find a token with similar name.
// Currently this is done by a simple substring search
func (db Database) FindFuzzy(name string) []*Entry {
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
		return err
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
		if err := entry.ReadFrom(r2); err != nil {
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
		if err := entry.WriteTo(plain); err != nil {
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
