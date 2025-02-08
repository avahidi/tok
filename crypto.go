package main

import (
    "hash"
	"crypto/hmac"
//	"crypto/sha256"
    "crypto/rand" // not math/rand :)
    "encoding/binary"
    "log"

)

func secureRandom(size int) []byte {
    data := make([]byte, size)
    if _, err := rand.Read(data); err != nil {
        log.Fatalf("Internal error: %v\n", err)
    }
    return data
}

// RFC 2898 PBKDF2 implementation
func PBKDF2(pass, salt []byte, count, dkLen int,
    hasher func () hash.Hash) []byte {
    h := hasher()
    hLen := h.Size()

    l := (dkLen + hLen - 1) / hLen
    r := dkLen - (l -1) * hLen


    var tmp [4]byte
    T := make([]byte, hLen)
    buffer := make([]byte, 0, r * hLen)

    for index := 0; index < l; index ++ {
        // T := pbkdf2_F(pass, salt, count, index)

         // Step 0:
        h := hmac.New(hasher, pass)
        h.Write(salt)
        binary.BigEndian.PutUint32(tmp[:], uint32(index + 1))
        h.Write( tmp[:])
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
