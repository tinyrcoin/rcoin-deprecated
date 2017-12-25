package main

import (
	"fmt"
	"os"
	"bytes"
	"crypto/sha256"
	"golang.org/x/crypto/ed25519"
	"golang.org/x/crypto/curve25519"
)
func getShared(public []byte, private []byte) []byte {
	var out [32]byte
	var pu, pr [32]byte
	copy(pu[:], public)
	copy(pr[:], private)
	curve25519.ScalarMult(&out, &pr, &pu)
	return out[:]
}
func main() {
	pass := []byte(os.Args[1])
	hashed := sha256.Sum256(pass)
	_, keys, _ := ed25519.GenerateKey(bytes.NewReader(hashed[:]))
	fmt.Printf("Public:\n%x\n", keys[32:])
	fmt.Printf("Private:\n%x\n", keys[:32])
	_, ot, _ := ed25519.GenerateKey(nil)
	fmt.Printf("%x\n%x\n", getShared(keys[32:], ot[:32]), getShared(ot[32:], keys[:32]))

}
