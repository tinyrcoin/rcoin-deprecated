package main
import "github.com/syndtr/goleveldb/leveldb"
import "github.com/syndtr/goleveldb/leveldb/util"

type Chain struct {
	DB *leveldb.DB
}

func OpenChain(path string) (*Chain, error) {
	c := new(Chain)
	db, e := leveldb.OpenFile(path, nil)
	if e != nil { return nil, e }
	c.DB = db
	if _, e = db.Get([]byte("block0"), nil); e != nil {
		db.Put([]byte("block0"),BirthdayBlock.Encode(), nil)
	}
	return c, nil
}
func (c *Chain) GetBalance(a Address) int64 {
	ret := int64(0)
	iter := c.DB.NewIterator(util.BytesPrefix([]byte("block")), nil)
	for iter.Next() {
		blk, _ := DecodeBlock(iter.Value())
		if a.Equals(blk.RewardTo) {
			ret += int64(len(blk.TX)) * 20
		}
		for _, v := range blk.TX {
			if a.Equals(v.From) {
				ret -= v.Amount
			}
			if a.Equals(v.To) {
				ret += v.Amount
			}
		}
	}
	iter.Release()
	return ret
}
