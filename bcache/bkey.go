package bcache

type BKey struct {
	High uint64
	Low  uint64
	Ptr  [6]uint64
}

func (b *BKey) Ptrs() uint64 {
	return bitfield(b.High, 60, 3)
}
func (b *BKey) HeaderSize() uint64 {
	return bitfield(b.High, 58, 2)
}
func (b *BKey) Csum() uint64 {
	return bitfield(b.High, 56, 2)
}
func (b *BKey) Pinned() uint64 {
	return bitfield(b.High, 55, 1)
}
func (b *BKey) Dirty() uint64 {
	return bitfield(b.High, 36, 1)
}
func (b *BKey) Size() uint64 {
	return bitfield(b.High, 20, 16)
}
func (b *BKey) Inode() uint64 {
	return bitfield(b.High, 0, 20)
}
func (b *BKey) Offset() uint64 {
	return b.Low
}
func (b *BKey) Start() uint64 {
	return b.Offset() - b.Size()
}

type BucketDisk struct {
	Prio uint16
	Gen  uint8
}
