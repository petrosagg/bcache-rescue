package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/petrosagg/bcache-rescue/bcache"
	"github.com/petrosagg/bcache-rescue/crc64"
	"github.com/satori/go.uuid"
	"log"
	"os"
)

func main() {
	file, err := os.Open("/dev/sdb2") // For read access.
	if err != nil {
		log.Fatal(err)
	}

	sb := bcache.CacheSuperBlock{}

	file.Seek(bcache.SB_START, 0)
	err = binary.Read(file, binary.LittleEndian, &sb)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("sb.magic\t\t")
	if sb.Magic == bcache.Magic {
		fmt.Printf("ok\n")
	} else {
		fmt.Printf("bad magic\n")
		fmt.Println("Invalid superblock: bad magic")
	}

	fmt.Printf("sb.first_sector\t\t%d", sb.Offset)
	if sb.Offset == bcache.SB_SECTOR {
		fmt.Printf(" [match]\n")
	} else {
		fmt.Printf(" [expected %ds]\n", bcache.SB_SECTOR)
		fmt.Println("Invalid superblock: bad magic")
	}

	buf := new(bytes.Buffer)
	binary.Write(buf, binary.LittleEndian, sb)
	expected_sum := crc64.Checksum(buf.Bytes()[8:], crc64.ECMA)
	fmt.Printf("sb.csum\t\t\t%x", sb.Csum)
	if sb.Csum == expected_sum {
		fmt.Printf(" [match]\n")
	} else {
		fmt.Printf(" [expected %x]\n", expected_sum)
	}

	fmt.Printf("sb.version\t\t%x", sb.Version)
	switch sb.Version {
	// These are handled the same by the kernel
	case bcache.BCACHE_SB_VERSION_CDEV:
	case bcache.BCACHE_SB_VERSION_CDEV_WITH_UUID:
		fmt.Printf(" [cache device]\n")

	// The second adds data offset support
	case bcache.BCACHE_SB_VERSION_BDEV:
	case bcache.BCACHE_SB_VERSION_BDEV_WITH_OFFSET:
		fmt.Printf(" [backing device]\n")

	default:
		fmt.Printf(" [unknown]\n")
	}

	fmt.Printf("\n")

	fmt.Printf("dev.label\t\t")
	label := string(sb.Label[:])
	if label[0] != 0x00 {
		fmt.Printf(label)
	} else {
		fmt.Printf("(empty)")
	}

	fmt.Printf("\n")

	u, _ := uuid.FromBytes(sb.UUID[:])
	fmt.Printf("dev.uuid\t\t%s\n", u.String())

	fmt.Printf("dev.sectors_per_block\t%d\ndev.sectors_per_bucket\t%d\n", sb.BlockSize, sb.BucketSize)

	if !sb.IsBackingDevice() {
		// total_sectors includes the superblock;
		fmt.Printf("dev.cache.first_sector\t%d\n", sb.BucketSize*sb.FirstBucket)
		fmt.Printf("dev.cache.cache_sectors\t%d\n", uint64(sb.BucketSize)*(sb.Nbuckets-uint64(sb.FirstBucket)))
		fmt.Printf("dev.cache.total_sectors\t%d\n", uint64(sb.BucketSize)*sb.Nbuckets)
		// fmt.Printf("dev.cache.ordered\t%s\n", CACHE_SYNC(&sb) ? "yes" : "no")
		// fmt.Printf("dev.cache.discard\t%s\n", CACHE_DISCARD(&sb) ? "yes" : "no")
		fmt.Printf("dev.cache.pos\t\t%d\n", sb.Nr_this_dev)
		// fmt.Printf("dev.cache.replacement\t%d", CACHE_REPLACEMENT(&sb))

		// switch CACHE_REPLACEMENT(&sb) {
		// 	case CACHE_REPLACEMENT_LRU:
		// 		printf(" [lru]\n");
		// 		break;
		// 	case CACHE_REPLACEMENT_FIFO:
		// 		printf(" [fifo]\n");
		// 		break;
		// 	case CACHE_REPLACEMENT_RANDOM:
		// 		printf(" [random]\n");
		// 		break;
		// 	default:
		// 		putchar('\n');
		// }
	}

	fmt.Printf("\n")

	u, _ = uuid.FromBytes(sb.SetUUID[:])
	fmt.Printf("cset.uuid\t\t%s\n", u.String())
}
