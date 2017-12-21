package main
import "github.com/syndtr/goleveldb/leveldb"
import "github.com/syndtr/goleveldb/leveldb/util"
import "fmt"
type ChainCache struct {
	balances map[string]int64
	hashes map[string]int64
	height int64
}
type Chain struct {
	DB *leveldb.DB
	Cache ChainCache
	LastDifficulty int
	GaveDiff int64
}
func OpenChain(path string) (*Chain, error) {
	c := new(Chain)
	db, e := leveldb.OpenFile(path, nil)
	if e != nil { return nil, e }
	c.DB = db
	if _, e = db.Get([]byte("block0"), nil); e != nil {
		db.Put([]byte("block0"),BirthdayBlock.Encode(), nil)
	}
	c.Cache.balances = map[string]int64{}
	c.Cache.hashes = map[string]int64{}
	return c, nil
}
func (c *Chain) AddRawBlock(data []byte) {
	id := fmt.Sprintf("block%d", c.Height())
	c.DB.Put([]byte(id), data, nil)
	c.Cache.height++
}
func (c *Chain) AddBlock(b *Block) {
	for _, t := range b.TX {
	if _, ok := c.Cache.balances[t.From.String()]; ok {
	c.Cache.balances[t.From.String()] -= t.Amount - t.CalcFee()
	}
	if _, ok := c.Cache.balances[t.To.String()]; ok {
	c.Cache.balances[t.To.String()] += t.Amount
	}
	}
	if _, ok := c.Cache.balances[b.RewardTo.String()]; ok {
	c.Cache.balances[b.RewardTo.String()] += b.CalcReward()
	}
	c.AddRawBlock(b.Encode())
	c.GetDifficulty()
}
func (c *Chain) GetRawBlock(id int64) []byte {
	d, _ := c.DB.Get([]byte(fmt.Sprintf("block%d", id)), nil)
	return d
}
func (c *Chain) GetBlock(id int64) *Block {
	r, _ := DecodeBlock(c.GetRawBlock(id))
	return r
}
func (c *Chain) GetDifficulty() (r int) {
	return c.getDifficulty(c.Height())
}
func (c *Chain) getDifficulty(height int64) (r int) {
	or := c.LastDifficulty
	defer func() {
		r = c.LastDifficulty
		if recover() != nil { r = or }
		if r < 1 { r = 10 } // ?!
	} ()
	if c.Height() < 4 {
		c.LastDifficulty = 10
		return
	}
	blk := c.GetBlock(height - 1)
	blk2 := c.GetBlock(height - 2)
	c.LastDifficulty = int((height*100)/((blk.Time-blk2.Time)+1))
	return
}
func (c *Chain) HashToBlockNum(hash []byte) int64 {
	if _, ok := c.Cache.hashes[string(hash)]; !ok {
		iter := c.DB.NewIterator(util.BytesPrefix([]byte("block")), nil)
		id := int64(0)
		for iter.Next() {
			b, _ := DecodeBlock(iter.Value())
			if string(b.Hash) == string(hash) {
				c.Cache.hashes[string(hash)] = id
				goto done
			}
			id++
		}
		iter.Release()
		return -1
		done:
		iter.Release()
	}
	return c.Cache.hashes[string(hash)]
}
func (c *Chain) Verify(b *Block) bool {
	if !b.Verify() {
		return false
	}
	if !b.VerifyPoW(c.GetDifficulty()) {
		return false
	}
	if len(b.TX) > 90 {
		return false
	}
	if string(c.GetBlock(c.Height()-1).Hash) != string(b.LastHash) {
		return false
	}
	if c.HashToBlockNum(b.LastHash) == -1 {
		return false
	}
	for _, v := range b.TX {
	if c.GetBalanceRaw(v.From) < v.Amount {
		return false
	}
	}
	return true
}
func (c *Chain) Height() int64 {
	if c.Cache.height == 0 {
		iter := c.DB.NewIterator(util.BytesPrefix([]byte("block")), nil)
		for iter.Next() { c.Cache.height++ }
	}
	return c.Cache.height
}
func (c *Chain) GetBalance(a Address) float64 {
	return float64(c.GetBalanceRaw(a)) / 1000
}
func (c *Chain) GetBalanceRaw(a Address) int64 {
	if _, ok := c.Cache.balances[a.String()]; !ok {
	ret := int64(0)
	iter := c.DB.NewIterator(util.BytesPrefix([]byte("block")), nil)
	for iter.Next() {
		blk, _ := DecodeBlock(iter.Value())
		fees := blk.CalcReward()
		for _, v := range blk.TX {
			if a.Equals(v.From) {
				ret -= v.Amount - v.CalcFee()
			}
			if a.Equals(v.To) {
				ret += v.Amount
			}
		}
		if a.Equals(blk.RewardTo) {
			ret += fees
		}
	}
	iter.Release()
	c.Cache.balances[a.String()] = ret
	}
	return c.Cache.balances[a.String()]
}
