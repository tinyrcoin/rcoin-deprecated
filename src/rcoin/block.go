package main
import "encoding/base32"
import "github.com/vmihailenco/msgpack"
import "golang.org/x/crypto/ed25519"
type Address []byte

func (a Address) String() string {
	return base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(a)
}
type Transaction struct {
	From Address
	To Address
	Signature Address
	Amount int64
}
type Block struct {
	LastHash []byte
	Hash []byte
	Nonce int64
	RewardTo Address
	Signature Address
	TX []Transaction
}
func (b *Block) Sign(key ed25519.PrivateKey) {
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
func (b *Block) ProofOfWork(difficulty int) {
	b.Hash = make([]byte, 64)
	b.Signature = make([]byte, 64)
	b.Nonce = 0
	for {
		b.Nonce++
	}
}
func (b *Block) Encode() []byte {
	ret, _ := msgpack.Marshal(b)
	return ret
}
func DecodeBlock(d []byte) (*Block, error) {
	b := new(Block)
	e := msgpack.Unmarshal(d, b)
	if e != nil { return nil, e }
	return b, nil
}
