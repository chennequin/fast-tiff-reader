package tiffio

type BinaryReader interface {
	open(name string) error
	close() error
	read(offset uint64, p []byte) (n int, err error)
}

type CachedBinaryReader interface {
	BinaryReader
	openMetaData()
	closeMetaData()
	readBlock(offset, size uint64) error
}
