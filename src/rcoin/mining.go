
package main
import "time"
import "log"
import mrand "math/rand"
var del = []string{}
func GetTransactions() []Transaction {
	if unconfirmed.Length() == 0 { return nil }
	ret := []Transaction(nil)
	unconfirmed.Range(func(kk, vv interface{}) bool {
		k := kk.(string)
		v := vv.(*Transaction)
		if len(ret) > 85 { return false }
		ret = append(ret, *v)
		del = append(del, k)
		return true
	})
	return ret
}
func PurgeOld() {
	for _, v := range del {
		unconfirmed.Delete(v)
	}
	del = []string{}
}

func Miner(threads int, payout []byte) {
	log.Println("Starting miner")
	log.Printf("Rewards go to %s", Address(payout[32:]).String())
	for {
		txs := GetTransactions()
		if txs == nil { time.Sleep(time.Second); continue }
		log.Printf("Working on block with %d transactions.", len(txs))
		b := NewBlock()
		b.TX = txs
		retry:
		time.Sleep(time.Second)
		if chain.GetBlock(chain.Height()-1) == nil {
			log.Printf("What? block %d is nil!", chain.Height()-1)
			goto retry
		}
		b.LastHash = chain.GetBlock(chain.Height()-1).Hash
		st := time.Now().Unix()
		cancel := false
		height := chain.Height()
		go b.ProofOfWork(chain.GetDifficulty(), threads, &cancel)
		for {
			time.Sleep(time.Second)
			if cancel {
				break
			}
			if height != chain.Height() {
				cancel = true
				goto cont
			}
		}
		goto skip
		cont:
			continue
		skip:
		b.Time = time.Now().Unix()
		b.Sign(payout)
		if !b.Verify() { panic("Block verification error") }
		time.Sleep(time.Duration(mrand.Int63n(4000)+500)*time.Millisecond)
		if !chain.Verify(b) {
			log.Println("Someone beat me to this block.")
			continue
		}
		PurgeOld()
		log.Printf("Solved block in %d seconds.", b.Time - st)
		Broadcast(Command{Type:CMD_BLOCK,Block:*b})
		chain.AddBlock(b)
	}
}
