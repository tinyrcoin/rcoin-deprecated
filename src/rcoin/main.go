package main
import (
	"os"
	"flag"
	"log"
)
var chain *Chain
var tests = map[string]func(){}
var datadir = flag.String("data", userhome() + "/.rcoin", "Data path")
var dotest = flag.String("test", "", "Run a test module")
var mining = flag.Bool("miner", true, "Enable mining")
var rpcport = flag.String("rpc", ":3009", "RPC listen port")
func main() {
	flag.Parse()
	if *dotest != "" {
		tests[*dotest]()
		return
	}
	log.Println("Starting rcoind")
	chain, _ = OpenChain(*datadir + "/rcoin.db")
	if chain == nil { log.Fatal("Blockchain corrupt") }
	RPCServer(*rpcport)
}

func userhome() string {
	s := os.Getenv("HOME")
	if s == "" { s = os.Getenv("USERPROFILE") }
	return s
}
