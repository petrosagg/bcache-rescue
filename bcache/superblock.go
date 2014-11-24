package bcache

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"os"

	"github.com/petrosagg/bcache-rescue/crc64"
	"github.com/satori/go.uuid"
)

// Version 0: Cache device
// Version 1: Backing device
// Version 2: Seed pointer into btree node checksum
// Version 3: Cache device with new UUID format
// Version 4: Backing device with data offset

const BCACHE_SB_VERSION_CDEV = 0
const BCACHE_SB_VERSION_BDEV = 1
const BCACHE_SB_VERSION_CDEV_WITH_UUID = 3
const BCACHE_SB_VERSION_BDEV_WITH_OFFSET = 4
const BCACHE_SB_MAX_VERSION = 4
const SB_SECTOR = 8
const SB_LABEL_SIZE = 32
const SB_JOURNAL_BUCKETS = 256
const BDEV_DATA_START_DEFAULT = 16 // sectors
const SB_START = (SB_SECTOR * 512)

var Magic = [16]byte{0xc6, 0x85, 0x73, 0xf6, 0x4e, 0x1a, 0x45, 0xca, 0x82, 0x65, 0xf5, 0x7f, 0x48, 0xba, 0x6d, 0x81}

type CacheDevice struct {
	Nbuckets    uint64 // device size
	BlockSize   uint16 // sectors
	BucketSize  uint16 // sectors
	Nr_in_set   uint16
	Nr_this_dev uint16
}

type BackingDevice struct {
	DataOffset uint64

	// block_size from the cache device section is still used by
	// backing devices, so don't add anything here until we fix
	// things to not need it for backing devices anymore
}

type CacheSuperblock struct {
	Csum    uint64
	Offset  uint64 // sector where this sb was written
	Version uint64

	Magic [16]byte
	UUID  [16]byte

	SetMagic uint64
	Padding  uint64
	Label    [SB_LABEL_SIZE]byte
	Flags    uint64
	Seq      uint64
	Pad      [8]uint64
	CacheDevice

	LastMount        uint32 // time_t
	FirstBucket      uint16
	NrJournalBuckets uint16
	D                [SB_JOURNAL_BUCKETS]uint64 // journal buckets
}

//type Cache struct {
//    Set *CacheSet
//	Superblock CacheSuperblock
//	SuperblockBio Bio
//	SuperblockBioVec [1]BioVec
//
//	KObj KObject
//    struct kobject      kobj;
//    struct block_device *bdev;
//
//    unsigned        watermark[WATERMARK_MAX];
//
//    struct task_struct  *alloc_thread;
//
//    struct closure      prio;
//    struct prio_set     *disk_buckets;

func MatchCsum(expected uint64, s interface{}) {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.LittleEndian, s)

	b := buf.Bytes()

	fmt.Println("Normal length", len(b))

	for i := 8; i <= len(b); i++ {
		if crc64.Checksum(b[8:i], crc64.ECMA) == expected {
			fmt.Printf("Correct len: %d Buffer: %x\n", len(b[8:i])+8, b[0:i])
		}
	}
}

func CsumSet(s interface{}) uint64 {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.LittleEndian, s)
	return crc64.Checksum(buf.Bytes()[8:], crc64.ECMA)
}

func ReadSuperblock(sb *CacheSuperblock, dev string) error {
	file, err := os.Open(dev) // For read access.
	if err != nil {
		return err
	}

	file.Seek(SB_START, 0)
	err = binary.Read(file, binary.LittleEndian, sb)

	if err != nil {
		return err
	}

	if sb.Offset != SB_SECTOR || sb.Magic != Magic {
		return errors.New("Not a bcache superblock")
	}

	if sb.NrJournalBuckets > SB_JOURNAL_BUCKETS {
		return errors.New("Too many journal buckets")
	}

	if sb.Csum != CsumSet(sb) {
		return errors.New("Bad checksum")
	}

	if sb.UUID == [16]byte{} {
		return errors.New("Bad UUID")
	}

	switch sb.Version {
	case BCACHE_SB_VERSION_BDEV:
		return errors.New("Baching device not implemented yet")
		// sb.DataOffset = BDEV_DATA_START_DEFAULT
	case BCACHE_SB_VERSION_BDEV_WITH_OFFSET:
		return errors.New("Baching device not implemented yet")
		// if sb.DataOffset < BDEV_DATA_START_DEFAULT {
		// 	return errors.New("Bad data offset")
		// }
	case BCACHE_SB_VERSION_CDEV:
	case BCACHE_SB_VERSION_CDEV_WITH_UUID:
		if sb.Nbuckets > math.MaxUint64 {
			return errors.New("Too many buckets")
		}

		if sb.Nbuckets < (1 << 7) {
			return errors.New("Not enough buckets")
		}
	}

	return nil
}

func (sb *CacheSuperblock) IsBackingDevice() bool {
	return sb.Version == BCACHE_SB_VERSION_BDEV || sb.Version == BCACHE_SB_VERSION_BDEV_WITH_OFFSET
}

func (sb *CacheSuperblock) PrintInfo() {
	fmt.Printf("sb.magic\t\t")
	if sb.Magic == Magic {
		fmt.Printf("ok\n")
	} else {
		fmt.Printf("bad magic\n")
		fmt.Println("Invalid superblock: bad magic")
	}

	fmt.Printf("sb.first_sector\t\t%d", sb.Offset)
	if sb.Offset == SB_SECTOR {
		fmt.Printf(" [match]\n")
	} else {
		fmt.Printf(" [expected %ds]\n", SB_SECTOR)
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
	case BCACHE_SB_VERSION_CDEV:
	case BCACHE_SB_VERSION_CDEV_WITH_UUID:
		fmt.Printf(" [cache device]\n")

	// The second adds data offset support
	case BCACHE_SB_VERSION_BDEV:
	case BCACHE_SB_VERSION_BDEV_WITH_OFFSET:
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

	// u, _ = uuid.FromBytes(sb.SetUUID[:])
	// fmt.Printf("cset.uuid\t\t%s\n", u.String())
}
