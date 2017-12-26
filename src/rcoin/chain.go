package main
import "github.com/syndtr/goleveldb/leveldb"
//import "github.com/syndtr/goleveldb/leveldb/util"
import "time"
import "github.com/satori/go.uuid"
import "fmt"
import "math/rand"
import "math/big"
import "encoding/base32"
import "strings"
import "encoding/json"
import "github.com/vmihailenco/msgpack"
import "golang.org/x/crypto/ed25519"
type Chain struct {
	DB *leveldb.DB
}
type Address []byte
func (a Address) String() string {
	return strings.ToLower(base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(a))
}
type Transaction struct {
	To Address
	From Address
	Amount int64
	UUID string
	Nonce int64
	Time int64
	Balance int64
	Signature []byte
}
func (t *Transaction) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"to": t.To.String(),
		"from": t.From.String(),
		"uuid": t.UUID,
		"time": t.Time,
		"signature": Address(t.Signature).String(),
		"amount": AmountToFloat(t.Amount),
		"raw": t.Encode(),
	})
}
func NewTransaction() *Transaction { n := new(Transaction); n.Time = time.Now().Unix(); n.UUID = uuid.NewV4().String(); return n }
func DecodeTransaction(b []byte) *Transaction {
	var t Transaction
	err := msgpack.Unmarshal(b, &t)
	if err != nil { return nil }
	return &t
}
var Genesis = &Transaction{
	To: StringToAddress("IVABM6GKTVHYGK4ZTZ2IJITTK5JV4JG4HIZQNIRQQKYFZT2AXZDA"),
	From: Address("Original"),
	UUID: "RCoin's first transaction!",
	Signature: []byte{1,2,3,4,5,6,7,8},
	Amount: FloatToAmount(1000.0),
}
func AmountToFloat(i int64) float64 {
	return float64(i)/10000
}
func FloatToAmount(f float64) int64 {
	return int64(f*10000)
}
func GetDifficulty() int {
	return 10+int(time.Now().Unix() - 1514245745)/8640*20
}
func GetOldDifficulty(tm int64) int {
	return 10+int(tm - 1514245745)/8640*20
}
func (t *Transaction) Encode() []byte {
	b, _ := msgpack.Marshal(t)
	return b
}
func (t *Transaction) Sign(w *Wallet) {
	t.Signature = []byte{}
	t.Signature = w.Sign(t.Encode())
}
func CalcReward(diff int) int64 {
	return ((25550)-int64(diff))*100
}
func (t *Transaction) Verify() bool {
	if (t.Time/8640) < ((chain.Latest()-4320)/8640) {
		return false
	}
	if t.From.String() == t.To.String() || t.Amount < chain.GetBalance(t.From) || t.Amount == 0 {
		return false
	}	
	if t.Nonce > 0 {
		if (t.Time - chain.LatestMinedOf(t.To)) < 290 {
			return false
		}
		if t.Amount > CalcReward(GetOldDifficulty(t.Time)) { return false }
		if !t.VerifyPoW(GetOldDifficulty(t.Time)) { return false }
		return true
	}
	osig := t.Signature
	t.Signature = []byte{}
	ret := ed25519.Verify([]byte(t.From), t.Encode(), osig)
	t.Signature = osig
	return ret
}

func StringToAddress(s string) Address {
	data, _ := base32.StdEncoding.WithPadding(base32.NoPadding).DecodeString(strings.ToUpper(s))
	if data == nil { return Address("") }
	return Address(data)
}
func OpenChain(path string) (*Chain, error) {
	c := new(Chain)
	db, e := leveldb.OpenFile(path, nil)
	if e != nil { return nil, e }
	c.DB = db
	//if _, e = db.Get([]byte("block0"), nil); e != nil {
	//	db.Put([]byte("block0"),BirthdayBlock.Encode(), nil)
	//}
	c.AddTransaction(Genesis)
	return c, nil
}
func (c *Chain) AddTransaction(t *Transaction) {
	c.DB.Put(t.Signature, t.Encode(), nil)
}
func (b *Transaction) ProofOfWork(difficulty int, threads int, cancel *bool) {
	null64 := make([]byte, 64)
	b.Signature = make([]byte, 64)
	b.Nonce = 0
	b.Time = 0
	hashes := 0
	finished := false
	done := make(chan int64)
	for i := 0; i < threads; i++ {
	go func() {
	b2 := DecodeTransaction(b.Encode())
	b2.Nonce = rand.Int63()
	for {
		b2.From = null64
		b2.From = HashBytes(b2.Encode())
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
		b.From = null64
		b.From = HashBytes(b.Encode())
		if cancel != nil {
		*cancel = true
		}
		finished = true
		b.Time = time.Now().Unix()
		fmt.Println("")
		return
	default:
		if cancel != nil && *cancel {
			finished = true
			fmt.Printf("\r                                         \n")
			return
		}
		time.Sleep(1 * time.Second)
		fmt.Printf("\r[%d hashes/second]", hashes)
		hashes = 0
	}
	}
	fmt.Println("")

}
var MAXUINT512 = new(big.Int).Exp(big.NewInt(2), big.NewInt(512), big.NewInt(0))
func (ob *Transaction) VerifyPoW(difficulty int) bool {
	b := DecodeTransaction(ob.Encode())
	b.From = make([]byte, 64)
	b.Signature = make([]byte, 64)
	b.Time = 0
	b.From = HashBytes(b.Encode())
	threshold := new(big.Int).Div(MAXUINT512, big.NewInt(int64(difficulty)))
	if new(big.Int).SetBytes(b.From).Cmp(threshold) <= 0 { return true }
	return false
}
func (c *Chain) GetBalance(a Address) int64 {
	bal := int64(0)
	iter := c.DB.NewIterator(nil, nil)
	for iter.Next() {
		//fmt.Printf("%x", iter.Key())
		t := DecodeTransaction(iter.Value())
		if t.From.String() == a.String() {
			bal -= t.Amount
		}
		if t.To.String() == a.String() {
			bal += t.Amount
		}
	}
	iter.Release()
	return bal
}
func (c *Chain) History(a Address, num int) []Transaction {
	bal := []Transaction{}
	iter := c.DB.NewIterator(nil, nil)
	for iter.Next() {
		t := DecodeTransaction(iter.Value())
		bal = append(bal, *t)
	}
	iter.Release()
	return bal
}
func (c *Chain) LatestMinedOf(a Address) int64 {
	bal := int64(0)
	iter := c.DB.NewIterator(nil, nil)
	for iter.Next() {
		t := DecodeTransaction(iter.Value())
		if t.To.String() == a.String() && t.Nonce > 0 && t.Time > bal {
			bal = t.Time - 1
		}
	}
	iter.Release()
	return bal
}
func (c *Chain) Latest() int64 {
	bal := int64(0)
	iter := c.DB.NewIterator(nil, nil)
	for iter.Next() {
		t := DecodeTransaction(iter.Value())
		if t.Time >= bal {
			bal = t.Time - 1
		}
	}
	iter.Release()
	return bal
}
func (c *Chain) GetTransaction(sig []byte) *Transaction {
	x, _ := c.DB.Get(sig, nil)
	if x == nil { return nil }
	return DecodeTransaction(x)
}
