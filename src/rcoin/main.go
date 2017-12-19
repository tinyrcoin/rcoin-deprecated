package main

import (
	"log"
	"golang.org/x/crypto/ed25519"
)

func main() {
	log.Println("Starting rcoind...")
	pub, priv, _ := ed25519.GenerateKey(nil)
	b := NewBlock()
	b.Amount = 23
	copy(b.To, pub)
	b.HashSign(priv)
	log.Println(b.Verify())
}
