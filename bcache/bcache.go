package bcache

func bitfield(i uint64, offset, size uint) uint64 {
	return (i >> offset) & ^(^uint64(0) << size)
}
