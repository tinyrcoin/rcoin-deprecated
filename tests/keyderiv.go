package main

import (
	"fmt"
	"os"
	"bytes"
	"crypto/sha256"
	"golang.org/x/crypto/ed25519"
)

func main() {
	pass := []byte(os.Args[1])
	hashed := sha256.Sum256(pass)
	_, keys, _ := ed25519.GenerateKey(bytes.NewReader(hashed[:]))
	fmt.Printf("Public:\n%x\n", keys[32:])
	fmt.Printf("Private:\n%x\n", keys[:32])
	
}
