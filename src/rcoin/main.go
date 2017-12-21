package main
import (
	"os"
	"flag"
	"log"
)
var chain *Chain
var tests = map[string]func(){}
var datadir = flag.String("data", userhome() + "/.rcoin", "Data `path`")
var dotest = flag.String("test", "", "Run a test `module`")
var mining = flag.Bool("miner", true, "Enable mining")
var rpcport = flag.String("rpc", "127.0.0.1:3009", "RPC listen `port`")
var peeraddr = flag.String("p2p", ":30009", "P2P `port`")
var bootstrap = flag.String("boot", "", "Bootstrap `peer`")
var opts = flag.String("o", "", "Specify misc. `options` separated by commas. Options: nonat, forceupnp, noirc.")
func main() {
	flag.Parse()
	if *dotest != "" {
		tests[*dotest]()
		return
	}
	os.Mkdir(*datadir, 0755)
	log.Println("Starting rcoind")
	chain, _ = OpenChain(*datadir + "/rcoin.db")
	if chain == nil { log.Fatal("Blockchain corrupt") }
	if *bootstrap != "" { go ConnectPeer(*bootstrap, true) }
	if *mining {
		w := GetWallet("default")
		if w != nil {
			go Miner(2, []byte(w.Private))
		}
	}
	go ListenPeer(*peeraddr)
	go PortForward()
	go PeerDiscover()
	RPCServer(*rpcport)
}

func userhome() string {
	s := os.Getenv("HOME")
	if s == "" { s = os.Getenv("USERPROFILE") }
	return s
}
