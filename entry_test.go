package main

import (
	"crypto"
	"testing"
)

func TestNewValid(t *testing.T) {
	e1, err := NewEntry("ok", "NZSXMZLSEBTW63TOME", "sha1", "", 30, 6)
	if err != nil {
		t.Fatalf("Could not create entry: %v", err)
	}
	if e1.Name != "ok" {
		t.Errorf("bad name: %s vs %s", "ok", e1.Name)
	}
	if e1.Secret != "NZSXMZLSEBTW63TOME" {
		t.Errorf("bad name: %s vs %s", "NZSXMZLSEBTW63TOME", e1.Secret)
	}
	if e1.Hash != crypto.SHA1 {
		t.Errorf("bad algorithm: %v vs %v", e1.Hash, crypto.SHA1)
	}

	if e1.Digits != 6 {
		t.Errorf("bad digits: %v vs %v", e1.Digits, 6)
	}
	if e1.Period != 30 {
		t.Errorf("bad period: %v vs %v", e1.Period, 30)
	}
}

func TestNewInvalid(t *testing.T) {
	_, err := NewEntry("ok", "Mary Had a Little Lamb", "sha1", "", 30, 6)
	if err == nil {
		t.Errorf("entry has invalid secret")
	}

	_, err = NewEntry("ok", "NZSXMZLSEBTW63TOME", "shenanigans-256", "", 30, 6)
	if err == nil {
		t.Errorf("entry has invalid algorithm")
	}
}

func TestNewUri(t *testing.T) {
	// test vectors from https://edent.codeberg.page/TOTP_Test_Suite/
	e1, err := EntryFromUri("otpauth://totp/issuer%3Aaccount%20name?secret=QWERTYUIOP&digits=6&issuer=issuer&algorithm=SHA1&period=30")
	if err != nil {
		t.Fatalf("Could not create entry: %v", err)
	}
	if e1.Name != "issuer:account name" {
		t.Errorf("bad name: %s vs %s", "issuer:account name", e1.Name)
	}
	if e1.Secret != "QWERTYUIOP" {
		t.Errorf("bad name: %s vs %s", "QWERTYUIOP", e1.Secret)
	}
	if e1.Hash != crypto.SHA1 {
		t.Errorf("bad algorithm: %v vs %v", e1.Hash, crypto.SHA1)
	}

	if e1.Digits != 6 {
		t.Errorf("bad digits: %v vs %v", e1.Digits, 6)
	}
	if e1.Period != 30 {
		t.Errorf("bad period: %v vs %v", e1.Period, 30)
	}

}
