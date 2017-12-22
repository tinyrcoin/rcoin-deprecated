
package main
import "time"
import "log"
func GetTransactions() []Transaction {
	if unconfirmed.Length() == 0 { return nil }
	ret := []Transaction(nil)
	del := []string(nil)
	unconfirmed.Range(func(kk, vv interface{}) bool {
		k := kk.(string)
		v := vv.(*Transaction)
		if len(ret) > 85 { return false }
		ret = append(ret, *v)
		del = append(del, k)
		return true
	})
	for _, v := range del {
		unconfirmed.Delete(v)
	}
	return ret
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
		b.ProofOfWork(chain.GetDifficulty(), threads)
		b.Time = time.Now().Unix()
		b.Sign(payout)
		if !b.Verify() { panic("Block verification error") }
		if !chain.Verify(b) {
			log.Println("Someone beat me to this block.")
			continue
		}
		log.Printf("Solved block in %d seconds.", b.Time - st)
		Broadcast(Command{Type:CMD_BLOCK,Block:*b}, "")
		chain.AddBlock(b)
	}
}
