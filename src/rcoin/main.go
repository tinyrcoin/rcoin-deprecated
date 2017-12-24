package main
import (
	"strings"
	"os"
	"flag"
	"log"
)
var chain *Chain
var tests = map[string]func(){}
var cli = flag.String("cli", "", "Use the basic built in command line interface to access a `wallet`. Or enter 'create' as the wallet name to create a new one.")
var datadir = flag.String("data", userhome() + "/.rcoin2", "Data `path`")
var dotest = flag.String("test", "", "Run a test `module`")
var mining = flag.Bool("miner", true, "Enable mining")
var threads = flag.Int("minerthreads", 2, "Number of threads for mining")
var rpcport = flag.String("rpc", "127.0.0.1:3009", "RPC listen `port`")
var pra = ""
var peeraddr = &pra
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
			go Miner(*threads, []byte(w.Private))
		}
	}
	RPCServer(*rpcport)
}

func userhome() string {
	s := os.Getenv("HOME")
	if s == "" { s = os.Getenv("USERPROFILE") }
	return s
}
