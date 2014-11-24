package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/petrosagg/bcache-rescue/bcache"
	"io"
	"os"
)

const CHUNK_SIZE = 2048 // 10MB
const FILE = "/dev/sdb2"

func KEY_FIELD(name string, i uint64, offset, size int) {
	d := (i >> uint(offset)) & ^(^uint64(0) << uint(size))
	fmt.Println(name, d)
}

func scan(src io.ReadSeeker, pattern []byte, limit int) {
	l := len(pattern)

	buf := make([]byte, CHUNK_SIZE)

	offset := -int64(CHUNK_SIZE - l)

	for limit != 0 {
		offset += int64(CHUNK_SIZE - l)
		src.Seek(offset, 0)
		// fmt.Printf("Offset: %dMB\n", offset/1024/1024)
		_, err := io.ReadFull(src, buf)
		if err != nil {
			fmt.Println("ERROR:", err)
			return
		}

		if idx := bytes.Index(buf, pattern); idx != -1 {
			src.Seek(int64(idx-8)+offset, 0)

			bset := bcache.JSet{}
			err := bset.Load(src)

			if err != nil {
				fmt.Println("ERROR:", err)
			}

			if bset.Keys != 15 {
				continue
			}

			fmt.Println()
			fmt.Printf("%d %d %d\n", bset.Seq, bset.LastSeq, bset.Keys)

			fmt.Printf("%x\n", bset.Start[0].High)
			KEY_FIELD("KEY_PTRS", bset.Start[0].High, 60, 3)
			KEY_FIELD("HEADER_SIZE", bset.Start[0].High, 58, 2)
			KEY_FIELD("KEY_CSUM", bset.Start[0].High, 56, 2)
			KEY_FIELD("KEY_PINNED", bset.Start[0].High, 55, 1)
			KEY_FIELD("KEY_DIRTY", bset.Start[0].High, 36, 1)

			KEY_FIELD("KEY_SIZE", bset.Start[0].High, 20, 16)
			KEY_FIELD("KEY_INODE", bset.Start[0].High, 0, 20)

			fmt.Println("Ptrs:", bset.Start[0].Ptrs())
			fmt.Println("HeaderSize:", bset.Start[0].HeaderSize())
			fmt.Println("Csum:", bset.Start[0].Csum())
			fmt.Println("Pinned:", bset.Start[0].Pinned())
			fmt.Println("Dirty:", bset.Start[0].Dirty())

			fmt.Println("Size:", bset.Start[0].Size())
			fmt.Println("Inode:", bset.Start[0].Inode())

			// fmt.Println("Magic:", bset.Magic)
			// fmt.Println("Seq:", bset.Seq)
			// fmt.Println("Version:", bset.Version)
			// fmt.Println("Keys:", bset.Keys)
			// fmt.Println("LastSeq:", bset.LastSeq)
			// fmt.Println("UUIDBucket:", bset.UUIDBucket)
			// fmt.Println("BTreeRoot:", bset.BTreeRoot)
			// fmt.Println("BtreeLevel:", bset.BtreeLevel)
			// fmt.Println("PrioBucket:", bset.PrioBucket)
			// fmt.Println("Data Length:", len(bset.Data))
			// // fmt.Println("Data:", bset.Data)
			// fmt.Println()

			// test := make([]bcache.BKey, 0)
			// bset.Start = test

			//if
			// bcache.MatchCsum(bset.Csum, bset) // == bset.Csum {
			//fmt.Println("Great success")
			//}

			limit--
		}

	}
}

func main() {
	sb := bcache.CacheSuperblock{}
	bcache.ReadSuperblock(&sb, FILE)

	fmt.Printf("Calculating BSET magic %x XOR %x\n", bcache.JSET_MAGIC, sb.SetMagic)

	bset_magic := bcache.JSET_MAGIC ^ sb.SetMagic

	fmt.Printf("JSET Magic: %d\n", bset_magic)

	pattern := new(bytes.Buffer)
	binary.Write(pattern, binary.LittleEndian, &bset_magic)

	fmt.Println("Pattern size", len(pattern.Bytes()))

	file, err := os.Open(FILE) // For read access.
	if err != nil {
		fmt.Println(err)
	}

	scan(file, pattern.Bytes(), 3)
}
