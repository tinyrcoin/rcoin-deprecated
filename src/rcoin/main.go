package main
import (
	"fmt"
	"time"
	"strings"
	"os"
	"flag"
	"log"
)
var chain *Chain
var tests = map[string]func(){}
var cli = flag.String("cli", "", "Use the basic built in command line interface to access a `wallet`. Or enter 'create' as the wallet name to create a new one.")
var datadir = flag.String("data", userhome() + "/.rcoinplus", "Data `path`")
var dotest = flag.String("test", "", "Run a test `module`")
var mining = flag.Bool("miner", true, "Enable mining")
var threads = flag.Int("minerthreads", 2, "Number of threads for mining")
var rpcport = flag.String("rpc", "127.0.0.1:3009", "RPC listen `port`")
var pra = ""
var peeraddr = &pra
var pausemining = false
var opts = flag.String("o", "", "Specify misc `options` separated by commas. Options: nonat, forceupnp, noirc.")
func main() {
	flag.Parse()
	if *cli != "" {
		cliConsole()
		return
	}
	if *dotest != "" {
		tests[*dotest]()
		return
	}
	if len(strings.Split(*peeraddr, ":")) == 1 {
		*peeraddr = "0.0.0.0" + *peeraddr
	}
	os.Mkdir(*datadir, 0755)
	log.Println("Starting rcoind")

	log.Printf("RPC: %s", *rpcport)
	chain, _ = OpenChain(*datadir + "/rcoin.db")
	if chain == nil { log.Fatal("Blockchain corrupt") }
	go InitPeerFramework()
	if *mining {
		w := GetWallet("default")
		if w == nil {
		w = GenerateWallet()
		log.Println("Auto-generating default wallet.")
		PutWallet("default", w)
		}
		if w != nil {
			go func() {
				lt := int64(10)
				for {
					fmt.Printf("Last mined block at %s\n", time.Unix(chain.LatestMinedOf(w.Public), 0).String())
					if (300-(time.Now().Unix() - chain.LatestMinedOf(w.Public))) > 0 {
					fmt.Printf("Waiting for mandatory cooldown... (%d seconds)\n", 300 - (time.Now().Unix() - chain.LatestMinedOf(w.Public)))
					fmt.Printf("\u2590                        \u258C\r")
					for i := int64(0); i < ((300 - (time.Now().Unix() - chain.LatestMinedOf(w.Public))) / 12); i++ {
						fmt.Printf("\u258C\b")
						time.Sleep(6 * time.Second) // Mandatory network cooldown
						fmt.Printf("\u2588")
						time.Sleep(6 * time.Second) // Mandatory network cooldown
					}
					}
					st := time.Now().Unix()	
					for pausemining { time.Sleep(time.Second) }
					t := NewTransaction()
					t.To = w.Public
					t.Amount = CalcReward(GetDifficulty())
					c := false
					t.ProofOfWork(GetDifficulty(), 2, &c)
					t.Time = time.Now().Unix()
					t.Sign(w)
					if !t.Verify() { continue }
					chain.AddTransaction(t)					
					Broadcast(Command{TX:*t,Type:CMD_TX})
					lt = (lt + (time.Now().Unix()-st)) / 2
				}
			} ()
		}
	}
	RPCServer(*rpcport)
}

func userhome() string {
	s := os.Getenv("HOME")
	if s == "" { s = os.Getenv("USERPROFILE") }
	return s
}
