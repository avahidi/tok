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
}

// NewEntry creates a new entry from given data. Note that secret is not sanity checked.
func NewEntry(name, secret string, period uint16, digits uint8, hash crypto.Hash) *Entry {
	return &Entry{
		Added:  time.Now().UnixMicro(),
		Period: period,
		Digits: digits,
		Hash:   hash,
		Name:   name,
		Secret: secret,
	}
}

func (e Entry) WriteTo(w io.Writer) error {
	return WriteMultiple(w, BYTE_ORDER, e.Added, e.Period, e.Digits, uint64(e.Hash), e.Name, e.Secret)
}

func (e *Entry) ReadFrom(r io.Reader) error {
	var tmp uint64
	err := ReadMultiple(r, BYTE_ORDER, &e.Added, &e.Period, &e.Digits, &tmp, &e.Name, &e.Secret)
	e.Hash = crypto.Hash(tmp) // Read cant handle crypto.Hash=uint, but uint64 works fine
	return err
}

func (e Entry) String() string {
	return fmt.Sprintf("(%s)\t[%d:%d:%s]\t%s", e.Date(), e.Period, e.Digits, e.Hash, e.Name)
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
