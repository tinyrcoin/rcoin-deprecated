package main
import "github.com/satori/go.uuid"
import "log"
import "os"
import "time"
import "net"
import "encoding/binary"
import "fmt"
import "github.com/vmihailenco/msgpack"
import "sync"
import "strings"
import "io/ioutil"
import "io"

type ConcurrentMap struct {
	sync.Map
}
func (c *ConcurrentMap) Length() int {
	i := 0
	c.Map.Range(func (k, v interface{}) bool { i++; return true })
	return i
}
//var unconfirmed = map[string]*Transaction{}
var nodeuuid = uuid.NewV4()
const /*\ COMMAND_TYPES \*/ (
	CMD_BLOCK = 1
	CMD_TX = 2
	CMD_PEER = 3
	CMD_GETBLOCK = 4
	CMD_SYNC = 5
	CMD_UUID = 6
)     /*\ COMMAND_TYPES \*/
type Command struct {
	Type uint8 // command type
	Block Block // a block, that belongs in the blockchain
	Text string // could be anything
	TX Transaction // for an unverified transaction
	RangeStart int64
	RangeEnd int64
	A, B, C int64
	// if this struct changed the protocol will break horribly
	// just a nice little warning
}
var unconfirmed = new(ConcurrentMap)
type Peer struct {
	Conn net.Conn
	Inbound bool
	Out bool
}
var peers = new(ConcurrentMap)
type PartialNodeBlockchain struct {
	// TODO: Partial node blockchain
}
func (p *Peer) GetCommand() (Command, error) {
	b := make([]byte, 4)
	_, err := io.ReadFull(p.Conn, b)
	if err != nil { return Command{}, err }
	i := int(binary.LittleEndian.Uint32(b))
	if i > (1024*1024*8) {
		return Command{}, fmt.Errorf("Someone is trying to DoS Me!")
	}
	o := make([]byte, i)
	_, err = io.ReadFull(p.Conn, o)
	if err != nil { return Command{}, err }
	var oc Command
	err = msgpack.Unmarshal(o, &oc)
	if err != nil { return Command{}, err }
	return oc, nil
}
func (p *Peer) PutCommand(c Command) error {
	if p.Out {
		return fmt.Errorf("Command buffer full")
	}
	p.Out = true
	p.putCommand(c)
	p.Out = false
	return nil
}
func (p *Peer) putCommand(c Command) {
	b := make([]byte, 4)
	o, _ := msgpack.Marshal(&c)
	binary.LittleEndian.PutUint32(b, uint32(len(o)))
	p.Conn.Write(b)
	p.Conn.Write(o)
}
func Broadcast(c Command, not string) {
	peers.Range(func(k, v interface{}) bool {
		if k.(string) != not {
			v.(*Peer).PutCommand(c)
		}
		return true
	})
}
func (p *Peer) Main(addr string) {
	p.PutCommand(Command{Type:CMD_SYNC,RangeStart:chain.Height()})
	Broadcast(Command{Type:CMD_PEER,Text:p.Conn.RemoteAddr().String()}, p.Conn.RemoteAddr().String())
	p.PutCommand(Command{Type:CMD_UUID,Text:nodeuuid.String()})
	if !IsNatted() && !p.Inbound {
		p.PutCommand(Command{Type:CMD_PEER,Text:":"+strings.Split(*peeraddr,":")[1],A:1})
	}
	done := false
	p.Out = false
	go func() {
		for !done {
			time.Sleep(60*time.Second)
			p.PutCommand(Command{Type:CMD_SYNC,RangeStart:chain.Height()})
		}
	} ()
	for {
		cmd, err := p.GetCommand()
		if err != nil {
			log.Printf("Lost peer: %s\n", err.Error())
			break
		}
		switch cmd.Type {
			case CMD_BLOCK:
				if !chain.Verify(&cmd.Block) {
					log.Println("Almost silently dropping bad block.")
					break
				}
				for _, v := range cmd.Block.TX {
					unconfirmed.Delete(string(v.Signature))
				}
				chain.AddBlock(&cmd.Block)
				Broadcast(cmd, p.Conn.RemoteAddr().String())
			break
			case CMD_UUID:
				if cmd.Text == nodeuuid.String() { goto end }
				log.Println("Connected to " + addr)
			break
			case CMD_SYNC:
				for i := cmd.RangeStart; i != cmd.RangeEnd && i < chain.Height(); i++ {
					if chain.GetBlock(i) == nil { continue }
					if i < chain.Height() { p.PutCommand(Command{Type:CMD_BLOCK,Block:*(chain.GetBlock(i))}) }
				}
				limit := 0
				unconfirmed.Range(func (k, v interface{}) bool { p.PutCommand(Command{Type:CMD_TX,TX:*(v.(*Transaction))}); return true })
				peers.Range(func (k, v interface{}) bool {
					limit++
					vp := v.(*Peer)
					if !vp.Inbound && v.(*Peer) != p {
						p.PutCommand(Command{Type:CMD_PEER,Text:k.(string)})
					}
					return limit < 30
				})
			break
			case CMD_TX:
				if !cmd.TX.Verify() {
					log.Printf("I got a bad transaction"); break
				}
				if chain.GetBalanceRaw(cmd.TX.From) < cmd.TX.Amount {
					log.Printf("(2) I got a bad transaction"); break
				}
				if *mining {
					unconfirmed.Store(string(cmd.TX.Signature), &cmd.TX)
				}
				Broadcast(cmd, p.Conn.RemoteAddr().String())
			break
			case CMD_PEER:
				if cmd.A == 1 {
					DbAddPeer(strings.Split(p.Conn.RemoteAddr().String(), ":")[0] + cmd.Text)
					break
				}
				if _, ok := peers.Load(cmd.Text); ok { break }
				if peers.Length() < 50 {
					go ConnectPeer(cmd.Text, true)
					break
				}
				DbAddPeer(cmd.Text)
			break
		}
	}
	end:
	done = true
}
func AddPeer(n net.Conn, inbound bool) {
	AddPeerEx(n.RemoteAddr().String(), n, inbound)
}
func AddPeerEx(s string, n net.Conn, inbound bool) {
	p := &Peer{n,inbound,false}
	peers.Store(n.RemoteAddr().String(), p)
	p.Main(s)
	peers.Delete(n.RemoteAddr().String())
}

func ConnectPeer(addr string, save bool) {
	if strings.HasPrefix(addr, "10.") || strings.HasPrefix(addr, "172.16.") || strings.HasPrefix(addr, "192.168.") {
		return
	}
	if _, ok := peers.Load(addr); ok {
		return
	}
	for {
	if _, ok := peers.Load(addr); ok {
		time.Sleep(time.Second*5)
		continue
	}
	n, e := net.Dial("tcp", addr)
	if e != nil {
		time.Sleep(time.Second*60)
		continue
	}
	if save { DbAddPeer(addr) }
	AddPeerEx(addr, n, false)
	time.Sleep(time.Second*60)
	}
}

func ListenPeer(addr string) {
	tryagain:
	srv, err := net.Listen("tcp", addr)
	if err != nil { return }
	for {
		peer, err := srv.Accept()
		if err != nil {
			log.Printf("We had an issue with the server: %s", err)
			srv.Close()
			goto tryagain
		}
		go AddPeer(peer, true)
	}
}
func DbAddPeer(addr string) {
	if DbHasPeer(addr) { return }
	peerDb, err := os.OpenFile(*datadir + "/peers.txt", os.O_RDWR|os.O_CREATE, 0755)
	if err != nil {
		log.Printf("Warning: can't add peer to database: %v", err)
	}
	peerDb.Seek(0,2)
	fmt.Fprintln(peerDb, addr)
	peerDb.Close()	
}
func DbHasPeer(addr string) bool {
	d, _ := ioutil.ReadFile(*datadir + "/peers.txt")
	if d == nil {
		return false
	}
	return strings.Contains(string(d), addr + "\n")
}
func DbConnectPeers() {
	f, err := os.Open(*datadir + "/peers.txt")
	if err != nil { return }
	for {
		var p string
		_, err := fmt.Fscanln(f, &p)
		if err != nil {
			break
		}
		go ConnectPeer(p, false)
	}
	f.Close()
}
