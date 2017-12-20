package main
import (
	"encoding/binary"
	"strings"
	"fmt"
)
func ProofOfWork(diff int, data []byte) (nonce int, hash []byte) {
	s := strings.Repeat("0", diff)
	data2 := append(data,0,0,0,0,0,0,0,0)
	for {
		binary.LittleEndian.PutUint64(data2[len(data2)-8:], uint64(nonce))
		hash = HashBytes(data2)
		if strings.HasPrefix(fmt.Sprintf("%x", hash), s) {
			break
		}
		nonce++
	}
	return
}
