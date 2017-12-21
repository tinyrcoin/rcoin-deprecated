package main

import (
	"bufio"
	"fmt"
	"os"
	"net/http"
	"io/ioutil"
	"encoding/json"
)

func cliCall(path string) Reply {
	c, e := http.Get("http://" + *rpcport + path)
	if e != nil {
		fmt.Println("Can't connect to node:", e)
		os.Exit(1)
	}
	buf, _ := ioutil.ReadAll(c.Body)
	c.Body.Close()
	var r Reply
	e = json.Unmarshal(buf, &r)
	if e != nil {
		fmt.Println("Bad reply from server.")
		os.Exit(1)
	}
	return r
}
func cliInfo(arg1, arg2, arg3 string) {
	ret := cliCall("/wallet/stat?name=" + *cli)
	if ret["error"] != nil {
		fmt.Println("No such wallet.")
		os.Exit(1)
	}
	fmt.Printf("Wallet %s\n---\nWallet Address: %v\nWallet Balance: %f\n", *cli, ret["address"], ret["balance"])
}
func cliSend(to, amtstr, arg3 string) {
	if to == "" || amtstr == "" {
		fmt.Println("Usage: send <to> <amount>")
		return
	}
	cliCall("/wallet/send?name=" + *cli + "&to=" + to + "&amount=" + amtstr)
	fmt.Println("Sent coins.")
}
func cliStatus(arg1, arg2, arg3 string) {
	for k, v := range cliCall("/stat") {
		fmt.Printf("%s: %v\n", k, v)
	}
}
func cliConsole() {
	if *cli == "create" {
		fmt.Printf("New wallet name: ")
		fmt.Scan(cli)
		ret := cliCall("/wallet/create?name=" + *cli)
		if ret["error"] != nil {
			fmt.Println("Error: wallet exists")
			os.Exit(1)
		}
	}
	r := bufio.NewReader(os.Stdin)
	help := map[string]string {
		"info": "Show wallet information",
		"send": "Usage: send <to> <amount>\nSend <amount> coins to <to>.\nI caution you: sending to a bad address will cause you to lose <amount> coins forever!",
		"exit": "Exit the console",
		"status": "Show node status",
	}
	fns := map[string]func(a, b, c string) {
		"info": cliInfo,
		"exit": func (a, b, c string) { os.Exit(0) },
		"send": cliSend,
		"status": cliStatus,
		"help": func (a, b, c string) {
			if a == "" {
				fmt.Println("Usage: help <string>")
				return
			}
			if _, ok := help[a]; !ok {
				fmt.Printf("No such topic: %s\n", a)
				return
			}
			fmt.Println(help[a])
		},
	}
	cliInfo("","","")
	for {
	fmt.Printf("> ")
	ln, err := r.ReadString('\n')
	if err != nil {
		println(err.Error())
		return
	}
	var cmd, arg1, arg2, arg3 string
	fmt.Sscan(ln, &cmd, &arg1, &arg2, &arg3)
	if fns[cmd] == nil {
		fmt.Printf("Commands:")
		for k, _ := range fns { fmt.Printf(" %s", k) }
		fmt.Println("")
		continue
	}
	fns[cmd](arg1,arg2,arg3)
	}
}
