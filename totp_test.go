package main

import (
	"bytes"
	"crypto"
	"crypto/sha1"
	"testing"
)

func TestHotp(t *testing.T) {
	// test vectors from RFC 4336
	const SECRET = "12345678901234567890"

	tests := []struct {
		counter int64
		output  uint32
	}{
		{0, 755224},
		{3, 969429},
		{9, 520489},
	}

	for _, test := range tests {
		totp := NewTotp([]byte(SECRET), DEFAULT_PERIOD, 6, sha1.New)

		got := totp.hotp(test.counter)
		if got != test.output {
			t.Errorf("HOTP(K, %d) expected %v got %v\n", test.counter, test.output, got)
		}
	}

}

func TestTotp(t *testing.T) {
	// test vectors from RFC 6248
	const SECRET1 = "12345678901234567890"
	const SECRET256 = SECRET1 + "123456789012"
	const SECRET512 = SECRET1 + SECRET1 + SECRET1 + "1234"

	tests := []struct {
		secret string
		hash_  crypto.Hash
		time   int64
		output uint32
	}{
		{SECRET1, crypto.SHA1, 0x0000001, 94287082},
		{SECRET1, crypto.SHA1, 0x23523EC, 7081804},
		{SECRET1, crypto.SHA1, 0x23523ED, 14050471},

		{SECRET256, crypto.SHA256, 0x0000001, 46119246},
		{SECRET256, crypto.SHA256, 0x23523EC, 68084774},
		{SECRET256, crypto.SHA256, 0x23523ED, 67062674},

		{SECRET512, crypto.SHA512, 0x0000001, 90693936},
		{SECRET512, crypto.SHA512, 0x23523EC, 25091201},
		{SECRET512, crypto.SHA512, 0x23523ED, 99943326},
	}

	for _, test := range tests {
		totp := NewTotp([]byte(test.secret), DEFAULT_PERIOD, 8, test.hash_.New)

		got := totp.totp(test.time)
		if got != test.output {
			t.Errorf("TOTP(K, %d)expected %v got %v\n", test.time, test.output, got)
		}
	}
}

func TestSecret(t *testing.T) {
	const ENCODED_SECRET = "AAAQ EAYE AUDA OCAJ"
	var SHORT_SECRET []byte = []byte{
		0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09,
	}
	var SECRET []byte = []byte{
		0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	}

	secret, err := secretFromBase64(ENCODED_SECRET)
	if err != nil {
		t.Errorf("Unable to extract secret: %v", err)
	} else if !bytes.Equal(SECRET, secret) {
		t.Errorf("Incorrect secret, wanted %v got %v", SECRET, secret)
	}

	secret, err = secretFromBytes(SHORT_SECRET)
	if err != nil {
		t.Errorf("Unable to extract secret: %v", err)
	} else if !bytes.Equal(SECRET, secret) {
		t.Errorf("Incorrect secret, wanted %v got %v", SECRET, secret)
	}
}
