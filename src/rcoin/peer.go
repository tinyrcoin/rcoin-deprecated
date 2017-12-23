package main

import (
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
)
type ConcurrentMap struct { sync.Map }
var unconfirmed = new(ConcurrentMap)
func (c *ConcurrentMap) Length() int {
	o := 0
	c.Map.Range(func(k,v interface{}) bool { o++; return true })
	return o
}
var br *bufio.Reader
var myid = uuid.NewV4().String()
const ROOM = "rcoin"
var ipfsapi = flag.String("ipfs", "http://127.0.0.1:5001/api/v0/pubsub/", "IPFS API Pubsub Endpoint")
const (
	CMD_BLOCK = 1
	CMD_TX = 2
	CMD_SYNC = 3
)
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
	http.Get(*ipfsapi + "pub?arg=" + ROOM + "&arg=" + url.QueryEscape(msg))
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
	tries := false
	retr:
	resp, err := http.Get(*ipfsapi + "sub?arg=" + ROOM + "&discover=true")
	if err != nil {
		if !tries {
			tries = true
			exec.Command("ipfs", "--init", "--enable-pubsub-experiment").Start()
			goto retr
		}
		log.Println(err)
		log.Fatal("Couldn't initialize peer framework")
	}
	Broadcast(Command{Type:CMD_SYNC,RangeStart:chain.Height()})

	br = bufio.NewReader(resp.Body)
	for {
		data, err := getMessage()
		if err != nil {
			log.Println(err)
			log.Fatal("Lost connection with peers")
		}
		go func() {
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
				for i := cmd.RangeStart; i != cmd.RangeEnd && i < chain.Height(); i++ {
					Broadcast(Command{To:cmd.From,Type:CMD_BLOCK,Block:*(chain.GetBlock(i))})
				}
				if cmd.A == 0 { Broadcast(Command{Type:CMD_SYNC,To:cmd.From,RangeStart:chain.Height()}) }
			break
			case CMD_BLOCK:
				if !chain.Verify(&cmd.Block) {
					log.Printf("Bad block from %s", cmd.From)
					break
				}
				for _, t := range cmd.Block.TX {
					unconfirmed.Delete(string(t.Signature))
				}
				chain.AddBlock(&cmd.Block)
			break
			case CMD_TX:
				if !cmd.TX.Verify() {
					log.Printf("Bad transaction from %s", cmd.From)
				}
				unconfirmed.Store(string(cmd.TX.Signature), &cmd.TX)
			break
		}
		} ()
	}
}
