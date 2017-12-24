package main

import (
	"io/ioutil"
	"time"
	"log"
	"net/http"
	"sync"
	"encoding/json"
	"encoding/base64"
	"bufio"
	"github.com/satori/go.uuid"
	"net/url"
	"flag"
	"github.com/vmihailenco/msgpack"
	"os/exec"
	"os"
)
type ConcurrentMap struct { sync.Map }
var unconfirmed = new(ConcurrentMap)
func (c *ConcurrentMap) Length() int {
	o := 0
	c.Map.Range(func(k,v interface{}) bool { o++; return true })
	return o
}
var br *bufio.Reader
var brp *bufio.Reader
var myid = uuid.NewV4().String()
const ROOM = "rcoin-v5"
var ipfsapi = flag.String("ipfs", "http://127.0.0.1:5001/api/v0/pubsub/", "IPFS API Pubsub Endpoint")
const (
	CMD_BLOCK = 1
	CMD_TX = 2
	CMD_SYNC = 3
)
var haltmine = false
type Command struct {
	From string
	To string
	Block Block
	TX Transaction
	RangeStart, RangeEnd int64
	A int64	
	Type int
}
func Broadcast(c Command) {
	c.From = myid
	b, _ := msgpack.Marshal(c)
	sendMessage(string(b))
}
func decodeMessage(in string) []byte {
	var m map[string]interface{}
	_ = json.Unmarshal([]byte(in), &m)
	if len(m) == 0 { return nil }
	d, _ := base64.StdEncoding.DecodeString(m["data"].(string))
	return d
}
func sendMessage(msg string) {
	k, e := http.Get(*ipfsapi + "pub?arg=" + ROOM + "&arg=" + url.QueryEscape(msg))
	if e != nil { return }
	k.Body.Close()
}
func sendMessageTo(to string, msg string) {
	k, e := http.Get(*ipfsapi + "pub?arg=" + to + "&arg=" + url.QueryEscape(msg))
	if e != nil { return }
	k.Body.Close()
}
func getMessage() ([]byte, error) {
	loop:
		k, e := br.ReadString('\n')
		if e != nil {
			return nil, e
		}
		m := decodeMessage(k)
		if m == nil {
			goto loop
		}
		return m, nil
}
func InitPeerFramework() {
	tries := 0
	var heights = map[string]int64{}
	var ignore = map[string]bool{}
	var hashes = map[string][]byte{}
	tophash := []byte("")
	topheight := chain.Height()
	os.Setenv("IPFS_PATH", *datadir + "/ipfs.db")
	retr:
	resp, err := http.Get(*ipfsapi + "sub?arg=" + ROOM + "&discover=true")
	if err != nil {
		if tries < 15 {
			tries++
			if tries == 1 {
			c := exec.Command("ipfs", "daemon", "--init", "--enable-pubsub-experiment")			
			c.Stdout = ioutil.Discard
			c.Stderr = c.Stdout
			c.Start()
			}
			log.Println("Waiting for ipfs to start...")
			time.Sleep(1500*time.Millisecond)
			goto retr
		}
		log.Println(err)
		log.Fatal("Couldn't initialize peer framework")
	}
	go func() {
		for {
			Broadcast(Command{Type:CMD_SYNC,RangeStart:chain.Height()})
			time.Sleep(10*time.Second)
		}
	} ()
	br = bufio.NewReader(resp.Body)
	for {
		topheight = 0
		for _, v := range heights {
			if v > topheight { topheight = v }
		}
		haltmine = chain.Height() < topheight
		tv := map[string]int{}
		for _, v := range hashes {
			tv[string(v)]++
		}
		th := 0
		for k, v := range tv {
			if v > th { th = v; tophash = []byte(k) }
		}
		for k, _ := range tv {
			if string(k) != string(tophash) {
				th--
			}
		}
		if th < 0 { tophash = []byte("") }
		ignore = map[string]bool{}
		data, err := getMessage()
		if err != nil {
			log.Println(err)
			log.Fatal("Lost connection with peers")
		}
		func() {
		var cmd Command
		err = msgpack.Unmarshal(data, &cmd)
		if err != nil {
			log.Println("Got corrupt message")
			return
		}
		if cmd.From == myid {return }
		if cmd.To != "" && cmd.To != myid {return }
		switch cmd.Type {
			case CMD_SYNC:
				go func() {
				for i := cmd.RangeStart; i != cmd.RangeEnd && i < chain.Height(); i++ {
					Broadcast(Command{To:cmd.From,Type:CMD_BLOCK,Block:*(chain.GetBlock(i))})
				}
				} ()
				if cmd.A == 0 {
					Broadcast(Command{Type:CMD_SYNC,To:cmd.From,RangeStart:chain.Height(),A:1}) 
				}
				if cmd.RangeStart != chain.Height() && !ignore[cmd.From] {
					log.Printf("peer: Syncing with %s (their blockchain height: %d, my height: %d)\n", cmd.From, cmd.RangeStart, chain.Height())
					heights[cmd.From] = cmd.RangeStart
				}
			break
			case CMD_BLOCK:
				if chain.HashToBlockNum(cmd.Block.Hash) != -1 {
					break
				}
				hashes[cmd.From] = cmd.Block.Hash
				if !chain.Verify(&cmd.Block) && string(tophash) == string(cmd.Block.Hash) {
					heights[cmd.From] = 0
					ignore[cmd.From] = true
					break
				}
				log.Printf("New block height: %d", chain.Height()+1)
				for _, t := range cmd.Block.TX {
					unconfirmed.Delete(string(t.Signature))
				}
				chain.AddBlock(&cmd.Block)
			break
			case CMD_TX:
				if !cmd.TX.Verify() || cmd.TX.From.String() == cmd.TX.To.String() {
					log.Printf("Bad transaction from %s", cmd.From)
				}
				unconfirmed.Store(string(cmd.TX.Signature), &cmd.TX)
			break
		}
		} ()
	}
}
