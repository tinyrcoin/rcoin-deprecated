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
var br2 *bufio.Reader
var myid = uuid.NewV4().String()
const ROOM = "rcoinplus-v1"
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
func BroadcastTo(m string, c Command) {
	c.From = myid
	b, _ := msgpack.Marshal(c)
	sendMessageTo(m,string(b))
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
var inmsg = make(chan string, 8)
var inerr = make(chan error, 8)
func rdrIn(r *bufio.Reader) {
	for {
	a, b := r.ReadString('\n')
	inmsg <- a
	inerr <- b
	}
}
func getMessage() ([]byte, error) {
	loop:
		k := <- inmsg
		e := <- inerr
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
	os.Setenv("IPFS_PATH", *datadir + "/ipfs.db")
	retr:
	resp, err := http.Get(*ipfsapi + "sub?arg=" + ROOM + "&discover=true")
	resp2, err := http.Get(*ipfsapi + "sub?arg=" + myid + "&discover=true")
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
	Broadcast(Command{Type:CMD_SYNC,A:chain.Latest()})
	br = bufio.NewReader(resp.Body)
	br2 = bufio.NewReader(resp2.Body)
	go rdrIn(br)
	go rdrIn(br2)
	for {
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
		if cmd.From == myid {return}
		if cmd.To != "" && cmd.To != myid {return}
		switch cmd.Type {
			case CMD_SYNC:
				//
				go func() {
				id := cmd.From
				subby, _ := http.Get(*ipfsapi + "sub?arg=" + id + "&discover=true")
				iter := chain.DB.NewIterator(nil, nil)
				for iter.Next() {
					t := DecodeTransaction(iter.Value())
					if t.Time > cmd.A {
						BroadcastTo(id, Command{Type:CMD_TX,TX:*t})
					}
				}
				subby.Body.Close()
				} ()
			break
			case CMD_TX:
				if cmd.TX.Time == 0 { return }
				if !cmd.TX.Verify() || cmd.TX.From.String() == cmd.TX.To.String() {
					log.Printf("Bad transaction from %s", cmd.From)
				}
				chain.AddTransaction(&cmd.TX)
			break
		}
		} ()
	}
}
