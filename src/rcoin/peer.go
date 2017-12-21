package main
import "log"
import "os"
import "net"
import "encoding/binary"
import "fmt"
import "github.com/vmihailenco/msgpack"
import "sync"
import "strings"
import "io/ioutil"
type ConcurrentMap struct {
	sync.Map
}
func (c *ConcurrentMap) Length() int {
	i := 0
	c.Map.Range(func (k, v interface{}) bool { i++; return true })
	return i
}
//var unconfirmed = map[string]*Transaction{}
var unconfirmed = new(ConcurrentMap)
type Peer struct {
	Conn net.Conn
	Inbound bool
}
var peers = new(ConcurrentMap)
const /*\ COMMAND_TYPES \*/ (
	CMD_BLOCK = 1
	CMD_TX = 2
	CMD_PEER = 3
	CMD_GETBLOCK = 4
	CMD_SYNC = 5
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
type PartialNodeBlockchain struct {
	// TODO: Partial node blockchain
}
func (p *Peer) GetCommand() (Command, error) {
	b := make([]byte, 4)
	_, err := p.Conn.Read(b)
	if err != nil { return Command{}, err }
	i := int(binary.LittleEndian.Uint32(b))
	if i > (1024*1024*8) {
		return Command{}, fmt.Errorf("Someone is trying to DoS Me!")
	}
	o := make([]byte, i)
	_, err = p.Conn.Read(o)
	if err != nil { return Command{}, err }
	var oc Command
	err = msgpack.Unmarshal(o, &oc)
	if err != nil { return Command{}, err }
	return oc, nil
}
func (p *Peer) PutCommand(c Command) {
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
func (p *Peer) Main() {
	p.PutCommand(Command{Type:CMD_SYNC,RangeStart:chain.Height()})
	Broadcast(Command{Type:CMD_PEER,Text:p.Conn.RemoteAddr().String()}, p.Conn.RemoteAddr().String())
	for {
		cmd, err := p.GetCommand()
		if err != nil {
			log.Printf("Lost peer: %s\n", err.Error())
			break
		}
		switch cmd.Type {
			case CMD_BLOCK:
				if !chain.Verify(&cmd.Block) {
					log.Printf("I got a bad block: dropping (we may have a bad peer or a hard fork occurred)")
					break
				}
				for _, v := range cmd.Block.TX {
					unconfirmed.Delete(string(v.Signature))
				}
				chain.AddBlock(&cmd.Block)
				Broadcast(cmd, p.Conn.RemoteAddr().String())
			break
			case CMD_SYNC:
				for i := cmd.RangeStart; i != cmd.RangeEnd && i < chain.Height(); i++ {
					if chain.GetBlock(i) == nil { continue }
					if i < chain.Height() { p.PutCommand(Command{Type:CMD_BLOCK,Block:*(chain.GetBlock(i))}) }
				}
				unconfirmed.Range(func (k, v interface{}) bool { p.PutCommand(Command{Type:CMD_TX,TX:*(v.(*Transaction))}); return true })
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
		}
	}
}

func AddPeer(n net.Conn, inbound bool) {
	p := &Peer{n,inbound}
	peers.Store(n.RemoteAddr().String(), p)
	p.Main()
	peers.Delete(n.RemoteAddr().String())
}

func ConnectPeer(addr string, save bool) {
	log.Printf("Connecting to peer %s", addr)
	n, e := net.Dial("tcp", addr)
	if e != nil {
		log.Printf("[peer %s] Failed to connect: %s", addr, e.Error())
		return
	}
	if save { DbAddPeer(addr) }
	AddPeer(n, false)
}

func ListenPeer(addr string) {
	srv, err := net.Listen("tcp", addr)
	if err != nil { return }
	for {
		peer, _ := srv.Accept()
		go AddPeer(peer, true)
	}
}
func DbAddPeer(addr string) {
	peerDb, err := os.OpenFile(*datadir + "/peers.txt", os.O_APPEND|os.O_CREATE, 0755)
	if err != nil {
		log.Printf("Warning: can't add peer to database: %v", err)
	}
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
