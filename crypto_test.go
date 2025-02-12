package main

import (
	"bytes"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"hash"
	"log"
	"testing"
)

func ensure[T any](val T, err error) T {
	if err != nil {
		log.Fatalf("Ensure failed: %v\n", err)
	}
	return val
}

// helper function to compare bytes in tests
func cmpbytes(t *testing.T, msg string, expected, got []byte) {
	if !bytes.Equal(expected, got) {
		t.Errorf("%s: Expected %s got %s",
			msg, hex.EncodeToString(expected), hex.EncodeToString(got))
	}
}

func TestRandom(t *testing.T) {
	r0 := secureRandom(4)
	r1 := secureRandom(4)

	// How do you test if random is random?
	// Well, we don't have time for that so let's just cheat
	if bytes.Equal(r0, r1) {
		t.Errorf("I think you are really unlucky: %v-%v\n", r0, r1)
	}
}

func TestHDK(t *testing.T) {
	// Test vectors from RFC 5869
	tests := []struct {
		hasher  func() hash.Hash
		key     string
		salt    string
		context string
		output  string
	}{
		{sha1.New, "0b0b0b0b0b0b0b0b0b0b0b", "000102030405060708090a0b0c", "f0f1f2f3f4f5f6f7f8f9",
			"085a01ea1b10f36933068b56efa5ad81a4f14b822f5b091568a9cdd4f155fda2c22e422478d305f3f896"},

		{
			sha256.New,
			"0b0b0b0b0b0b0b0b0b0b0b0b0b0b0b0b0b0b0b0b0b0b",
			"000102030405060708090a0b0c",
			"f0f1f2f3f4f5f6f7f8f9",
			"3cb25f25faacd57a90434f64d0362f2a2d2d0a90cf1a5a4c5db02d56ecc4c5bf34007208d5b887185865",
		},
		{
			sha256.New,
			"000102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f202122232425262728292a2b2c2d2e2f303132333435363738393a3b3c3d3e3f404142434445464748494a4b4c4d4e4f",
			"606162636465666768696a6b6c6d6e6f707172737475767778797a7b7c7d7e7f808182838485868788898a8b8c8d8e8f909192939495969798999a9b9c9d9e9fa0a1a2a3a4a5a6a7a8a9aaabacadaeaf",
			"b0b1b2b3b4b5b6b7b8b9babbbcbdbebfc0c1c2c3c4c5c6c7c8c9cacbcccdcecfd0d1d2d3d4d5d6d7d8d9dadbdcdddedfe0e1e2e3e4e5e6e7e8e9eaebecedeeeff0f1f2f3f4f5f6f7f8f9fafbfcfdfeff",
			"b11e398dc80327a1c8e7f78c596a49344f012eda2d4efad8a050cc4c19afa97c59045a99cac7827271cb41c65e590e09da3275600c2f09b8367793a9aca3db71cc30c58179ec3e87c14c01d5c1f3434f1d87",
		},
	}

	for _, test := range tests {
		key, _ := hex.DecodeString(test.key)
		salt, _ := hex.DecodeString(test.salt)
		context, _ := hex.DecodeString(test.context)
		output, _ := hex.DecodeString(test.output)
		got, err := HKDF(key, salt, context, len(output), test.hasher)
		if err != nil {
			t.Errorf("HKDF failed: %v", err)
		} else {
			cmpbytes(t, "HKDF", output, got)
		}
	}
}

func TestPBKDF2Sha1(t *testing.T) {
	// Test vectors from IETF
	testsSHA1 := []struct {
		pass string
		salt string
		iter int
		size int

		hexOutput string
	}{
		{"password", "salt", 1, 20, "0c60c80f961f0e71f3a9b524af6012062fe037a6"},
		{"password", "salt", 2, 20, "ea6c014dc72d6f8ccd1ed92ace1d41f0d8de8957"},
		{"password", "salt", 4096, 20, "4b007901b765489abead49d926f721d065a429c1"},
	}

	for _, test := range testsSHA1 {
		pass := []byte(test.pass)
		salt := []byte(test.salt)
		output, _ := hex.DecodeString(test.hexOutput)
		got := PBKDF2(pass, salt, test.iter, test.size, sha1.New)
		cmpbytes(t, "PBKDF2-SHA1", output, got)
	}
}

func TestPbkdf2Sha256(t *testing.T) {
	testsSha256 := []struct {
		pass      string
		salt      string
		iter      int
		size      int
		hexOutput string
	}{
		{"password", "salt", 1, 32, "120fb6cffcf8b32c43e7225256c4f837a86548c92ccc35480805987cb70be17b"},
		{"password", "salt", 2, 32, "ae4d0c95af6b46d32d0adff928f06dd02a303f8ef3c251dfd6e2d85a95474c43"},
		{"password", "salt", 654321, 20, "4d8814e8217b358bdf36a2937efb47fff7754a92"},
	}

	for _, test := range testsSha256 {
		pass := []byte(test.pass)
		salt := []byte(test.salt)
		output, _ := hex.DecodeString(test.hexOutput)
		got := PBKDF2(pass, salt, test.iter, test.size, sha256.New)
		cmpbytes(t, "PBKDF2-SHA256", output, got)
	}
}

func TestEncrypt(t *testing.T) {
	var KEY1 = secureRandom(32)
	var INPUT []byte = []byte{0, 1, 2, 70}

	enc1, err := encryptBytes(KEY1, INPUT)
	if err != nil {
		t.Fatalf("Unable to encrypt bytes: %v", err)
	}

	dec1, err := decryptBytes(KEY1, enc1)
	if err != nil {
		t.Fatalf("Unable to decrypt bytes: %v", err)
	}

	if bytes.Equal(INPUT, enc1) {
		t.Errorf("Data was not encrypted")
	}

	if !bytes.Equal(INPUT, dec1) {
		t.Errorf("Invalid encryption/decryption")
	}
}

func TestEncryptKey(t *testing.T) {
	var KEY1 = secureRandom(32)
	var INPUT []byte = []byte{0, 1, 2, 70}

	enc1, err := encryptBytes(KEY1, INPUT)
	if err != nil {
		t.Fatalf("Unable to encrypt bytes: %v", err)
	}

	// create a key that differs in one byte and check of the output is different
	for i := 0; i < len(KEY1); i++ {
		var key2 = make([]byte, len(KEY1))
		copy(key2, KEY1)
		key2[i] ^= 0x01

		enc2, err := encryptBytes(key2, INPUT)
		if err != nil {
			t.Fatalf("Unable to encrypt bytes: %v", err)
		}
		if bytes.Equal(enc1, enc2) {
			t.Errorf("data encrypted with different keys should result in different ciphertext")
		}
	}
}

func TestEncryptAuth(t *testing.T) {
	var KEY1 = secureRandom(32)
	var INPUT []byte = []byte{0, 1, 2, 70}

	enc1, err := encryptBytes(KEY1, INPUT)
	if err != nil {
		t.Fatalf("Unable to encrypt bytes: %v", err)
	}

	// Create a message that differs in one byte and check that if fails decryption
	for i := 0; i < len(enc1); i++ {
		badenc := make([]byte, len(enc1))
		copy(badenc, enc1)
		badenc[i] ^= 0x01

		if _, err := decryptBytes(KEY1, badenc); err == nil {
			t.Fatalf("decryption should fail due to message corruption at byte %d...", i)
		}
	}
}

func BenchmarkHDK_SHA1(b *testing.B) {
	key := secureRandom(20)
	salt := secureRandom(20)
	context := secureRandom(20)
	size := 40

	for i := 0; i < b.N; i++ {
		if _, err := HKDF(key, salt, context, size, sha1.New); err != nil {
			b.Errorf("HKDF failed: %v", err)
		}
	}
}

func BenchmarkHDK_256(b *testing.B) {
	key := secureRandom(32)
	salt := secureRandom(32)
	context := secureRandom(20)
	size := 40

	for i := 0; i < b.N; i++ {
		if _, err := HKDF(key, salt, context, size, sha256.New); err != nil {
			b.Errorf("HKDF failed: %v", err)
		}
	}
}

func BenchmarkHDK_512(b *testing.B) {
	key := secureRandom(64)
	salt := secureRandom(64)
	context := secureRandom(20)
	size := 40

	for i := 0; i < b.N; i++ {
		if _, err := HKDF(key, salt, context, size, sha512.New); err != nil {
			b.Errorf("HKDF failed: %v", err)
		}
	}
}

func BenchmarkPBKDF2Sha1(b *testing.B) {
	pass := []byte("password")
	salt := []byte("salt")
	iter := 10240
	size := 20
	for i := 0; i < b.N; i++ {
		PBKDF2(pass, salt, iter, size, sha1.New)
	}
}

func BenchmarkPBKDF2Sha256(b *testing.B) {
	pass := []byte("password")
	salt := []byte("salt")
	iter := 10240
	size := 20
	for i := 0; i < b.N; i++ {
		PBKDF2(pass, salt, iter, size, sha256.New)
	}
}

func BenchmarkPBKDF2Sha512(b *testing.B) {
	pass := []byte("password")
	salt := []byte("salt")
	iter := 10240
	size := 20
	for i := 0; i < b.N; i++ {
		PBKDF2(pass, salt, iter, size, sha512.New)
	}
}
