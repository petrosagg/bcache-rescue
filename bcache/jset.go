package bcache

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"

	"github.com/petrosagg/bcache-rescue/crc64"
)

const MAX_CACHES_PER_SET = 8
const JSET_MAGIC = uint64(0x245235c1a3625032)

type JSet struct {
	JSetFixed
	Start []BKey
}

type JSetFixed struct {
	Csum    uint64
	Magic   uint64
	Seq     uint64
	Version uint32
	Keys    uint32

	LastSeq uint64

	UUIDBucket BKey

	BTreeRoot BKey

	BtreeLevel uint16
	Pad        [3]uint16

	PrioBucket [MAX_CACHES_PER_SET]uint64
}

func (self *JSet) Load(src io.Reader) (err error) {
	err = binary.Read(src, binary.LittleEndian, &self.JSetFixed)

	if err != nil {
		return
	}

	i := uint32(0)
	for i = uint32(0); i < self.Keys; {
		key := BKey{}
		binary.Read(src, binary.LittleEndian, &key.High)
		binary.Read(src, binary.LittleEndian, &key.Low)

		// fmt.Println()
		fmt.Printf("key.High: %x key.Low: %x\n", key.High, key.Low)
		// fmt.Printf("key.Low: %x\n", key.Low)
		// fmt.Println("Ptrs:", key.Ptrs())
		// fmt.Println("HeaderSize:", key.HeaderSize())
		// fmt.Println("Csum:", key.Csum())
		// fmt.Println("Pinned:", key.Pinned())
		// fmt.Println("Dirty:", key.Dirty())

		// fmt.Println("Size:", key.Size())
		// fmt.Println("Inode:", key.Inode())

		for j := uint64(0); j < key.Ptrs(); j++ {
			if key.Low == 0x50950cbda84053a5 {
				i += 3
			}
			i++
			// fmt.Println(i, self.Keys)
			binary.Read(src, binary.LittleEndian, &key.Ptr[j])
			// fmt.Printf("Added Pointer: %x\n", key.Ptr[j])
		}

		self.Start = append(self.Start, key)
	}

	if i != self.Keys {
		fmt.Printf("WTF happened")
	}

	if !self.Verify() {
		err = errors.New("Invalid checksum")
	}

	return
}

func (j *JSet) Verify() bool {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.LittleEndian, j.JSetFixed)

	start := buf.Len()

	b := make([]byte, start+int(j.Keys*8))
	// copy(b, buf.Bytes())

	// for i := 0; i < int(j.Keys); i++ {
	// 	binary.LittleEndian.PutUint64(b[start+i*8:], j.Data[i])
	// }

	return crc64.Checksum(b[8:], crc64.ECMA) == j.Csum
}
