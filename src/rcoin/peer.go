package main
import "log"
import "net"
import "encoding/binary"
import "fmt"
import "github.com/vmihailenco/msgpack"
type Peer struct {
	Conn net.Conn
	Inbound bool
}
const /* COMMAND_TYPES */ (
	CMD_BLOCK = 1
	CMD_TX = 2
	CMD_PEER = 3
)
type Command struct {
	Type uint8 // command type
	Block Block // a block, that belongs in the blockchain
	Text string // could be anything
	TX Transaction // for an unverified transaction
	// if this is changed the protocol will break horribly
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
	if i > (1024*1024*32) {
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
func (p *Peer) Main() {
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
				chain.AddBlock(&cmd.Block)
			break
		}
	}
}
