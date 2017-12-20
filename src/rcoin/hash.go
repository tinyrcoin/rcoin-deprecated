package main

import (
	"crypto/sha512"
	"golang.org/x/crypto/scrypt"
)

func HashBytes(b []byte) []byte {
	seed := sha512.Sum512(b)
	scrypted, _ := scrypt.Key(b, seed[:], 8, 512, 1, 128)
	out := sha512.Sum512(scrypted)
	return out[:]
}
