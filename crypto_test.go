package main

import (
    "crypto/sha1"
    "crypto/sha256"
    "testing"
    "bytes"
    "encoding/hex"
    "log"
)

func ensure[T any](val T, err error) T {
    if err != nil {
        log.Fatalf("Ensure failed: %v\n", err)
    }
    return val
}
func TestRandom(t *testing.T) {
	r0 := secureRandom(4)
	r1 := secureRandom(4)

	// How do you test if random is random?
	// Well, we don't have time for that so let's just cheat
	if bytes.Equal(r0,r1) {
		t.Errorf("I think you are really unlucky: %v-%v\n", r0, r1)
	}
}


func TestPbkdf2Sha1(t *testing.T) {
    // Test vectors from IETF
	testsSHA1 := []struct {
		pass string
		salt string
		iter int
		size int

		hexOutput string
	}{
	    {"password", "salt", 1, 20, "0c60c80f961f0e71f3a9b524af6012062fe037a6" },
	    {"password", "salt", 2, 20, "ea6c014dc72d6f8ccd1ed92ace1d41f0d8de8957" },
	    {"password", "salt", 4096, 20, "4b007901b765489abead49d926f721d065a429c1"},
	}

    for _, test := range testsSHA1 {
        pass := []byte(test.pass)
        salt := []byte(test.salt)
        output, _ := hex.DecodeString(test.hexOutput)
        got := PBKDF2(pass, salt, test.iter, test.size, sha1.New)

        if !bytes.Equal(got, output) {
        	t.Errorf("PBKDF2-SHA1 expected %v, got %v\n", output, got)
        }
    }
}

func TestPbkdf2Sha256(t *testing.T) {
	testsSha256 := []struct {
		pass string
		salt string
		iter int
		size int
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

        if !bytes.Equal(got, output) {
        	t.Errorf("PBKDF2-SHA256 expected %v, got %v\n", output, got)
        }
    }
}