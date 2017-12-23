package main

import (
	"net/http"
	"io"
	"encoding/json"
	"encoding/base64"
	"bufio"
	"net/url"
)

const ENDPOINT = "http://127.0.0.1:5001/api/v0/pubsub/"
func decodeMessage(in string) []byte {
	var m map[string]interface{}
	_ = json.Unmarshal([]byte(in), &m)
	if len(m) == 0 { return nil }
	d, _ := base64.StdEncoding.DecodeString(m["data"].(string))
	return d
}
func sendMessage(msg string) {
	http.Get(ENDPOINT + "pub?arg=hi&arg=" + url.QueryEscape(msg))
}
func getMessage(c io.Reader) ([]byte, error) {
	br := bufio.NewReader(c)
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
func main() {
	println("Testing ipfs pubsub sub")
	c, _ := http.Get(ENDPOINT + "sub?arg=hi&discover=true")
	sendMessage("A test by me")
	for {
		msg, _ := getMessage(c.Body)
		println(string(msg))
	}
}
