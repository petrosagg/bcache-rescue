package bcache

var Magic = [16]byte{0xc6, 0x85, 0x73, 0xf6, 0x4e, 0x1a, 0x45, 0xca, 0x82, 0x65, 0xf5, 0x7f, 0x48, 0xba, 0x6d, 0x81}

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

type CacheSuperBlock struct {
	Csum    uint64
	Offset  uint64 // sector where this sb was written
	Version uint64

	Magic [16]byte
	UUID  [16]byte

	SetUUID [16]byte
	Label   [SB_LABEL_SIZE]byte
	Flags   uint64
	Seq     uint64
	Pad     [8]uint64
	CacheDevice

	LastMount   uint32 // time_t
	FirstBucket uint16
	Keys        uint16
	D           [SB_JOURNAL_BUCKETS]uint64 // journal buckets
}

func (sb *CacheSuperBlock) IsBackingDevice() bool {
	return sb.Version == BCACHE_SB_VERSION_BDEV || sb.Version == BCACHE_SB_VERSION_BDEV_WITH_OFFSET
}
