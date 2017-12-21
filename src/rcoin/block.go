
package main
import "time"
import "math/big"
import "encoding/base32"
import "fmt"
import "github.com/vmihailenco/msgpack"
import "golang.org/x/crypto/ed25519"
type Address []byte
func init() {
	tests["mine"] = func() {
		blk := NewBlock()
		blk.ProofOfWork(10000,2)
		fmt.Printf("%x\n", blk.Hash)
	}
	tests["verify"] = func() {
		blk := NewBlock()
		trs := NewTransaction()
		blk.AddTransaction(trs)
		pub, priv, _ := ed25519.GenerateKey(nil)
		trs.Sign(priv)
		println(trs.Verify())
		_ = pub
	}
}
func (a Address) String() string {
	return base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(a)
}
func (a Address) Equals(b Address) bool {
	return fmt.Sprintf("%x", a) == fmt.Sprintf("%x", b)
}
func DecodeAddress(s string) Address {
	a, _ := base32.StdEncoding.WithPadding(base32.NoPadding).DecodeString(s)
	return a
}
type Transaction struct {
	From Address
	To Address
	Signature Address
	Amount int64
}
func (t *Transaction) Encode() []byte {
	ret, _ := msgpack.Marshal(t)
	return ret
}
func (t *Transaction) Sign(key ed25519.PrivateKey) {
	t.From = Address(key[32:])
	t.Signature = make([]byte, 64)
	t.Signature = ed25519.Sign(key, t.Encode())
}
func (t *Transaction) Verify() bool {
	sig := t.Signature
	t.Signature = make([]byte, 64)
	ret := ed25519.Verify([]byte(t.From), t.Encode(), sig)
	t.Signature = sig
	return ret
}
func (t *Transaction) SetAmount(f float64) {
	t.Amount = int64(f * 1000)
}
func (t *Transaction) GetAmount() float64 {
	return float64(t.Amount)/1000
}
func NewTransaction() (t *Transaction) {
	t = new(Transaction)
	t.From = make(Address, 32)
	t.To = make(Address, 32)
	t.Signature = make([]byte, 64)
	return
}
type Block struct {
	LastHash []byte
	Hash []byte
	Nonce int64
	RewardTo Address
	Signature Address
	TX []Transaction
	Time int64
}
var BirthdayBlock = &Block{
	make([]byte, 64),
	make([]byte, 64),
	0,
	DecodeAddress("IVABM6GKTVHYGK4ZTZ2IJITTK5JV4JG4HIZQNIRQQKYFZT2AXZDA"),
	make([]byte, 64),
	nil,
	0,
}
func (b *Block) Verify() bool {
	osig := b.Signature
	b.Signature = make([]byte, 64)
	ret := ed25519.Verify([]byte(b.RewardTo), b.Encode(), osig)
	b.Signature = osig
	return ret
}
func (b *Block) Sign(key ed25519.PrivateKey) {
	b.RewardTo = Address(key[32:])
	b.Signature = make([]byte, 64)
	b.Signature = ed25519.Sign(key, b.Encode())
}
func (b *Block) SetHash() {
	osig := b.Signature
	b.Hash = make([]byte, 64)
	b.Signature = make([]byte, 64)
	ret := HashBytes(b.Encode())
	b.Signature = osig
	b.Hash = ret
}
func (b *Block) ProofOfWork(difficulty int, threads int) {
	null64 := make([]byte, 64)
	b.Signature = make([]byte, 64)
	b.Nonce = 0
	b.Time = 0
	hashes := 0
	finished := false
	done := make(chan int64)
	for i := 0; i < threads; i++ {
	go func() {
	b2, _ := DecodeBlock(b.Encode())
	for {
		b2.Hash = null64
		b2.Hash = HashBytes(b2.Encode())
		hashes++
		if b2.VerifyPoW(difficulty) { break }
		b2.Nonce++
		if finished { return }
	}
	done <- b2.Nonce
	} ()
	}
	for {
	select {
	case b.Nonce = <- done:
		b.SetHash()
		return
	default:
		time.Sleep(1 * time.Second)
		fmt.Printf("\r %d hashes/second    \r", hashes)
		hashes = 0
	}
	}

}
var MAXUINT512 = new(big.Int).Exp(big.NewInt(2), big.NewInt(512), big.NewInt(0))
func (b *Block) VerifyPoW(difficulty int) bool {
	t := b.Time
	defer func() { b.Time = t } ()
	b.Time = 0
	threshold := new(big.Int).Div(MAXUINT512, big.NewInt(int64(difficulty)))
	if new(big.Int).SetBytes(b.Hash).Cmp(threshold) <= 0 { return true }
	return false
}
func (b *Block) Encode() []byte {
	ret, _ := msgpack.Marshal(b)
	return ret
}
func (b *Block) AddTransaction(t *Transaction) {
	b.TX = append(b.TX, *t)
}
func DecodeBlock(d []byte) (*Block, error) {
	b := new(Block)
	e := msgpack.Unmarshal(d, b)
	if e != nil { return nil, e }
	return b, nil
}
func NewBlock() *Block {
	b := new(Block)
	b.RewardTo = make(Address, 32)
	b.Signature = make(Address, 64)
	b.LastHash = make([]byte, 64)
	b.Hash = make([]byte, 64)
	b.TX = make([]Transaction, 0, 32)
	return b
}

