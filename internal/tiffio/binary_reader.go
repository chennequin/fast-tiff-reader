package tiffio

type BinaryReader interface {
	open(name string) error
	close() error
	seek(offset uint64) (uint64, error)
	read(p []byte) (n int, err error)
}
