package tiffio

type BinaryReader interface {
	open(name string) error
	close() error
	seek(offset int64) (int64, error)
	read(p []byte) (n int, err error)
}
