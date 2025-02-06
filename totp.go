package main

import (
	"crypto/hmac"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base32"
	"encoding/binary"
	"fmt"
	"hash"
	"log"
	"strings"
	"time"
)

type Totp struct {
	Secret []byte
	Period int64
	hasher func() hash.Hash
}

func NewTotp(secret []byte) *Totp {

	hasher := algoFromSecretLength(len(secret))
	return &Totp{Secret: secret, Period: 30, hasher: hasher}
}

// HOTP according to RFC 4226
func (t Totp) hotp(counter int64, digits int) uint32 {

	// Step 1: Generate an HMAC-SHA-x value Let HS = HMAC-SHA-x(K,C)  // HS
	mac := hmac.New(t.hasher, t.Secret)

	asBin := make([]byte, 8)
	binary.BigEndian.PutUint64(asBin, uint64(counter))
	mac.Write(asBin)

	hash := mac.Sum(nil)

	// Step 2: dynamic truncation to 4 bytes
	offset := hash[len(hash)-1] & 15
	hash = hash[offset : offset+4]

	// Step 3: Convert to number:
	mod := digitsToMod(digits)
	num := binary.BigEndian.Uint32(hash) & 0x7FFF_FFFF
	return num % mod
}

func (t Totp) totp(counter int64, digits int) uint32 {
	return t.hotp(counter, digits)
}

func (t Totp) Generate(digits int) string {
	counter := tx() / t.Period
	kod := fmt.Sprintf("%d", t.totp(counter, digits))
	for len(kod) < digits {
		kod = "0" + kod
	}
	return kod
}

// helper to return the mod value for a number of digits
func digitsToMod(d int) uint32 {
	var POW10 = []uint32{1, 10, 100, 1000, 10000, 100000, 1000000, 10000000, 100000000}
	return POW10[d]
}

// helper function to get the required time
func tx() int64 {
	return time.Now().Unix()
}

// give the size of the data, figure out what algorithm we should use:
func algoFromSecretLength(n int) func() hash.Hash {
	switch n * 8 {
	case 160:
		return sha1.New
	case 256:
		return sha256.New
	case 512:
		return sha512.New

	default:
		log.Fatalf("Unknown algorithm for secret size %d\n", n)
		return nil
	}
}

func secretFromBase64(str string) ([]byte, error) {
	str = strings.Replace(str, " ", "", -1)
	for len(str)%8 != 0 {
		str += "="
	}

	key, err := base32.StdEncoding.DecodeString(str)
	if err != nil {
		return nil, err
	}
	return secretFromBytes(key)
}

func secretFromBytes(bs []byte) ([]byte, error) {
	for len(bs) != 20 && len(bs) != 32 && len(bs) != 64 {
		bs = append(bs, 0)
	}
	return bs, nil
}
