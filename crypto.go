package main

import (
	"crypto"
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand" // not math/rand :)
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"hash"
	"log"
	"strings"

	// we must import what we plan to use or hash.New won't find them
	_ "crypto/sha1"
	_ "crypto/sha512"
)

const (
	PBKDF2_ITERATIONS = 45821
)

// secureRandom is a helper function for generating crypto-safe random
func secureRandom(size int) []byte {
	data := make([]byte, size)
	if _, err := rand.Read(data); err != nil {
		log.Fatalf("Internal error: %v\n", err)
	}
	return data
}

// HKDF is the RFC-5869 hash key derivation ufnction.
// Note that SHA-512 is required in the CNSA 2.0 approved version
func HKDF(key, salt, info []byte, kLen int, hasher func() hash.Hash) ([]byte, error) {

	// prk = Extract(salt, key_in)
	h := hmac.New(hasher, salt)
	h.Write(key)
	prk := h.Sum(nil)

	// key_out = Expand(prk, info, length)
	var T, Tx []byte
	for i := 1; len(T) < kLen; i++ {
		h = hmac.New(hasher, prk)
		h.Write(Tx)
		h.Write(info)
		h.Write([]byte{byte(i)})
		Tx = h.Sum(nil)

		T = append(T, Tx...)
	}
	return T[:kLen], nil
}

// PBKDF2 is an implementation of RFC 2898 PBKDF2
func PBKDF2(pass, salt []byte, count, dkLen int,
	hasher func() hash.Hash) []byte {
	h := hasher()
	hLen := h.Size()

	l := (dkLen + hLen - 1) / hLen
	r := dkLen - (l-1)*hLen

	var tmp [4]byte
	T := make([]byte, hLen)
	buffer := make([]byte, 0, r*hLen)

	for index := 0; index < l; index++ {
		// T := pbkdf2_F(pass, salt, count, index)

		// Step 0:
		h := hmac.New(hasher, pass)
		h.Write(salt)
		binary.BigEndian.PutUint32(tmp[:], uint32(index+1))
		h.Write(tmp[:])
		dgst := h.Sum(nil)
		copy(T, dgst)

		// step N
		for i := 1; i < count; i++ {
			h = hmac.New(hasher, pass)
			h.Write(dgst)
			dgst = h.Sum(nil)
			for j, val := range dgst {
				T[j] ^= val
			}
		}
		buffer = append(buffer, T...)
	}
	return buffer[:dkLen]
}

// GenerateKeyFromPassword uses PBKDF2 to generate a key from a password
func GenerateKeyFromPassword(pass, salt []byte) []byte {
	return PBKDF2(pass, salt, PBKDF2_ITERATIONS, 32, sha256.New)
}

// encryptBytes is a helper for GCM-AES-256 encryption, result starts with nonce
func encryptBytes(key, data []byte) ([]byte, error) {
	aes, err := aes.NewCipher(key[:])
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(aes)
	if err != nil {
		return nil, err
	}
	nonce := secureRandom(gcm.NonceSize())
	ciphertext := gcm.Seal(nil, nonce, data, nil)
	return append(nonce, ciphertext...), nil
}

// decryptBytes is a helper for GCM-AES-256 decryption, assumes nonce is at start of the data
func decryptBytes(key, data []byte) ([]byte, error) {
	aes, err := aes.NewCipher(key[:])
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(aes)
	if err != nil {
		return nil, err
	}
	n := gcm.NonceSize()
	if len(data) < n {
		return nil, fmt.Errorf("Invalid data format")
	}
	nonce, ciphertext := data[:n], data[n:]
	return gcm.Open(nil, nonce, ciphertext, nil)
}

// hashFromName converts hash name to crypto.Hash
func hashFromName(name string) (crypto.Hash, error) {
	name = strings.ToLower(name)
	name = strings.Replace(name, "-", "", -1)
	name = strings.Replace(name, "_", "", -1)

	switch name {
	case "sha1":
		return crypto.SHA1, nil
	case "sha256":
		return crypto.SHA256, nil
	case "sha512":
		return crypto.SHA512, nil
	default:
		return 0, fmt.Errorf("unknown hash format: '%s'", name)
	}
}
