package main

import "net"
import "time"
import "fmt"
import "strings"
import mrand "math/rand"
import "bufio"
const IRC_SERVER = "irc.lfnet.org:6667"
func PeerDiscover() {
	if !strings.Contains(*opts, "noirc") {
	go IRCPeerDiscover()
	}
}
func IRCPeerDiscover() {
	me := fmt.Sprintf("rcoin%d", mrand.Intn(65536))
	for {
		time.Sleep(1*time.Second)
		irc, err := net.Dial("tcp", IRC_SERVER)
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
				println(err.Error())
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
					if arg2 == ":.conn" && len(peers) < 30 && peers[srchost + ":" + arg3] == nil {
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

