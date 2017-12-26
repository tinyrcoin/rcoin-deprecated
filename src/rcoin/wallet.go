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
	return StringToAddress(s)
}
func (w *Wallet) Balance(c *Chain) float64 {
	return AmountToFloat(c.GetBalance(w.Public))
}
func (w *Wallet) Send(c *Chain, a Address, amt float64, comment string) {
	tx := NewTransaction()
	tx.Amount = FloatToAmount(amt)
	tx.To = a
	tx.From = w.Public
	tx.Sign(w)
	if !tx.Verify() { return }
	c.AddTransaction(tx)
	Broadcast(Command{Type:CMD_TX,TX:*tx})
}
func (w *Wallet) Sign(b []byte) []byte {
	return ed25519.Sign(w.Private, b)
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
	w.Public = StringToAddress(a)
	w.Private = []byte(StringToAddress(b))
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
