package main

import "io/ioutil"
import "fmt"
import "golang.org/x/crypto/ed25519"
type Wallet struct {
	Public Address
	Private []byte
}
func DecodeWalletAddress(s string) Address {
	c := GetWallet(s)
	if c != nil { return c.Public }
	return DecodeAddress(s)
}
func (w *Wallet) Balance(c *Chain) float64 {
	return c.GetBalance(w.Public)
}
func (w *Wallet) Send(c *Chain, a Address, amt float64) {
	tx := NewTransaction()
	tx.SetAmount(amt)
	tx.To = a
	tx.Sign(w.Private)
	unconfirmed.Store(string(tx.Signature), tx)
	Broadcast(Command{Type:CMD_TX,TX:*tx},"")
}
func GenerateWallet() (w *Wallet) {
	w = new(Wallet)
	pu, pr, _ := ed25519.GenerateKey(nil)
	w.Public = Address(pu)
	w.Private = pr
	return
}

func (w *Wallet) Encode() string {
	return fmt.Sprintf("%s %s\n", w.Public.String(), Address(w.Private).String())
}

func DecodeWallet(s string) (w *Wallet) {
	w = new(Wallet)
	var a string
	var b string
	fmt.Sscanf(s, "%s %s", &a, &b)
	w.Public = DecodeAddress(a)
	w.Private = []byte(DecodeAddress(b))
	return
}

func (w *Wallet) Save(path string) {
	ioutil.WriteFile(path, []byte(w.Encode()), 0600)
}

func LoadWallet(path string) (w *Wallet) {
	d, e := ioutil.ReadFile(path)
	if e != nil { return nil }
	w = DecodeWallet(string(d))
	return
}

func GetWallet(name string) *Wallet {
	return LoadWallet(*datadir + "/" + name + ".wallet")
}
func PutWallet(name string, w *Wallet) {
	w.Save(*datadir + "/" + name + ".wallet")
}
func HasWallet(name string) bool {
	return GetWallet(name) != nil
}
