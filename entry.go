package main

import (
	"crypto"
	"fmt"
	"io"
	"time"
)

// Entry represents one item in the database
type Entry struct {
	Added  int64
	Period uint16
	Digits uint8
	Hash   crypto.Hash
	Name   string
	Secret string
	Note   string
}

// NewEntry creates a new entry from given data, sanity checks period, digits, secret and hashname
func NewEntry(name, secret, hashname, note string, period, digits int) (*Entry, error) {
	// check secret:
	if _, err := secretFromBase64(secret); err != nil {
		return nil, fmt.Errorf("Invalid secret: %v", err)
	}

	// check hash
	hash, err := hashFromName(hashname)
	if err != nil {
		return nil, fmt.Errorf("Invalid algorithm: %v", err)
	}

	// check period
	if period < 1 {
		return nil, fmt.Errorf("Invalid period: %v", period)
	}

	// check digits
	if digits < 1 || digits > 9 {
		return nil, fmt.Errorf("Invalid digits: %v", digits)
	}

	return &Entry{
		Added:  time.Now().UnixMicro(),
		Period: uint16(period),
		Digits: uint8(digits),
		Hash:   hash,
		Name:   name,
		Secret: secret,
		Note:   note,
	}, nil
}

func (e Entry) Serial(w io.Writer) error {
	return WriteMultiple(w, BYTE_ORDER, e.Added, e.Period, e.Digits, uint64(e.Hash), e.Name, e.Secret, e.Note)
}

func (e *Entry) Deserial(r io.Reader) error {
	var tmp uint64
	err := ReadMultiple(r, BYTE_ORDER, &e.Added, &e.Period, &e.Digits, &tmp, &e.Name, &e.Secret, &e.Note)
	e.Hash = crypto.Hash(tmp) // Read cant handle crypto.Hash=uint, but uint64 works fine
	return err
}

func (e Entry) Date() string {
	t := time.UnixMicro(e.Added)
	return t.Format("2006-01-02 15:04:05")
}

// Totp creates a TOTP object from this entry, to be queried for a code
func (e Entry) Totp() (*Totp, error) {
	secret, err := secretFromBase64(e.Secret)
	if err != nil {
		return nil, err
	}

	return NewTotp(secret, int(e.Period), int(e.Digits), e.Hash.New), nil
}
