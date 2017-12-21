package main

import "net"
import "time"
import "fmt"
import "strings"
import mrand "math/rand"
import "bufio"
import "log"
const IRC_SERVER = "irc.lfnet.org:6667"
func PeerDiscover() {
	if !strings.Contains(*opts, "noirc") {
		log.Println("Looking for peers using IRC")
		go IRCPeerDiscover()
	}
}
func IRCPeerDiscover() {
	mrand.Seed(time.Now().UnixNano())
	me := fmt.Sprintf("rcoin%d", mrand.Intn(65536))
	for {
		time.Sleep(1*time.Second)
		irc, err := net.Dial("tcp4", IRC_SERVER) // not everyone has ipv6
		if err != nil { continue }
		fmt.Fprintf(irc, "USER rcoin * 8 :Rcoin-discovery\r\n")
		fmt.Fprintf(irc, "NICK %s\r\n", me)
		bf := bufio.NewReader(irc)
		for {
			var src, cmd, arg1, arg2, arg3 string
			var ircln string
			ircln, err = bf.ReadString('\n')
			fmt.Sscan(ircln + "\n", &src, &cmd, &arg1, &arg2, &arg3)
			if err != nil {
				break
			}
			switch cmd {
				case "001":
					fmt.Fprintf(irc, "JOIN #rcoin\r\n")
					fmt.Fprintf(irc, "PRIVMSG #rcoin :.req\r\n")
					break
				case "PRIVMSG":
					srchost := strings.Split(src, "@")[1]
					if strings.Contains(srchost, ":") { srchost = "[" + srchost + "]" }
					_, ok := peers.Load(srchost + ":" + arg3)
					if arg2 == ":.conn" && peers.Length() < 30 && (!ok) {
						ConnectPeer(srchost + ":" + arg3, true)
					}
					if arg2 == ":.req" && !IsNatted() {
						fmt.Fprintf(irc, "PRIVMSG %s :.conn %s\r\n", arg1, strings.Split(*peeraddr, ":")[1])
					}
					break
				case "PING":
					fmt.Fprintf(irc, "PONG %s\r\n", arg1)
					break
			}
		}
	}
}

