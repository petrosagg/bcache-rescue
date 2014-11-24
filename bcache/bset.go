package bcache

const BSET_MAGIC = uint64(0x90135c78b99e07f6)

type BSet struct {
	Csum    uint64
	Magic   uint64
	Seq     uint64
	Version uint32
	Keys    uint32
	D       uint64
}
