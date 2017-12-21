package main
import "strconv"
import "github.com/ccding/go-stun/stun"
import "strings"
import "log"
import "github.com/NebulousLabs/go-upnp"
var upnp_ok = false
func IsNatted() bool {
	if strings.Contains(*opts, "nonat") { return false }
	if strings.Contains(*opts, "forceupnp") { return true }
	if upnp_ok { return false }
	nattype, _, _ := stun.NewClient().Discover()
	if nattype != stun.NATNone { return true }
	return false
}

func PortForward() {
	if !IsNatted() { return }
	log.Println("You are behind a NAT. We will try to forward a port for you.")
	if !TryUPnP() {}
	if !upnp_ok {
		log.Println("You are behind a NAT and we can't forward a port for you.")
		log.Println("This is bad and limits your opportunities to connect with other user.")
		log.Println("Please manually forward the p2p port.")
	}
}

func TryUPnP() bool {
	d, err := upnp.Discover()
	if err != nil { return false }
	var port int
	str := strings.Split(*peeraddr, ":")[1]
	port, _ = strconv.Atoi(str)
	d.Clear(uint16(port))
	err = d.Forward(uint16(port), "RCoin Port Forward")
	if err != nil { return false }
	upnp_ok = true
	return true
}
