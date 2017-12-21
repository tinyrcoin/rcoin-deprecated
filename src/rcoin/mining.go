
package main
import "time"
import "log"
func GetTransactions() []Transaction {
	if len(unconfirmed) == 0 { return nil }
	ret := []Transaction(nil)
	del := []string(nil)
	for k, v := range unconfirmed {
		if len(ret) > 85 { break }
		ret = append(ret, *v)
		del = append(del, k)
	}
	for _, v := range del {
		delete(unconfirmed, v)
	}
	return ret
}

func Miner(threads int, payout []byte) {
	log.Println("Starting miner")
	log.Printf("Rewards go to %s", Address(payout[32:]).String())
	for {
		txs := GetTransactions()
		if txs == nil { time.Sleep(30*time.Second); continue }
		log.Printf("Working on block with %d transactions.", len(txs))
		b := NewBlock()
		b.TX = txs
		b.LastHash = chain.GetBlock(chain.Height()-1).Hash
		b.ProofOfWork(chain.GetDifficulty(), threads)
		b.Time = time.Now().Unix()
		b.Sign(payout)
		if !b.Verify() { panic("Block verification error") }
		if !chain.Verify(b) {
			log.Println("Someone beat me to this block.")
			continue
		}
		Broadcast(Command{Type:CMD_BLOCK,Block:*b}, "")
		chain.AddBlock(b)
	}
}
