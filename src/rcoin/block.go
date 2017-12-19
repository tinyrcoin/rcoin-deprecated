package main
import "golang.org/x/crypto/ed25519"
import "encoding/binary"
import "fmt"
import "time"
type Block struct {
	Signature []byte
	Hash []byte
	LastHash []byte
	To []byte
	From []byte
	Amount int64
	Time int64
}
func NewBlock() *Block {
	ret := new(Block)
	ret.Signature = make([]byte, 64)
	ret.Hash = make([]byte, 64)
	ret.LastHash = make([]byte, 64)
	ret.To = make([]byte, 64)
	ret.From = make([]byte, 64)
	ret.Amount = 0
	ret.Time = 0
	return ret
}
func (b *Block) Sign(key ed25519.PrivateKey) {
	copy(b.From, key[32:])
	copy(b.Signature, ed25519.Sign(key, b.Encode(false)))
}
const BLOCK_FMT = `Block
  Signature data: %x
  Hash: %x
  Hash of previous block: %x
  To: %x
  From: %x
  Amount: RCN %d
  Time: %s
`
func (b *Block) String() string {
	return fmt.Sprintf(BLOCK_FMT, b.Signature, b.Hash, b.LastHash, b.To, b.From, b.Amount, time.Unix(b.Time, 0).String())
}
func (b *Block) SetHash() {
	empty := make([]byte, 64)
	copy(b.Signature, empty)
	copy(b.Hash, empty)
	hash := HashBytes(b.Encode(false))
	copy(b.Hash, hash)
}
func (b *Block) HashSign(key ed25519.PrivateKey) {
	b.SetHash()
	b.Sign(key)
}
func (b *Block) Verify() bool {
	ok := ed25519.Verify(b.From[:32], b.Encode(false), b.Signature)
	return ok
}
func (b *Block) Encode(full bool) []byte {
	out := make([]byte, 64+64+64+64+64+8+8)
	if full {
	copy(out, b.Signature)
	}
	copy(out[64:], b.Hash)
	copy(out[128:], b.LastHash)
	copy(out[192:], b.To)
	copy(out[256:], b.From)
	binary.LittleEndian.PutUint64(out[320:], uint64(b.Amount))
	binary.LittleEndian.PutUint64(out[328:], uint64(b.Time))
	return out
}
func BlockDecode(raw []byte) *Block {
	out := NewBlock()
	copy(out.Signature, raw)
	copy(out.Hash, raw[64:])
	copy(out.LastHash, raw[128:])
	copy(out.To, raw[192:])
	copy(out.From, raw[256:])
	out.Amount = int64(binary.LittleEndian.Uint64(raw[320:]))
	out.Time = int64(binary.LittleEndian.Uint64(raw[328:]))
	return out
}
