package tiffio

import (
	"fmt"
	"math"
	"os"
)

type FileBinaryReader struct {
	file *os.File
}

func NewFileBinaryReader() BinaryReader {
	return &FileBinaryReader{}
}

func (f *FileBinaryReader) open(name string) error {
	file, err := os.Open(name)
	if err != nil {
		return fmt.Errorf("unable to open file %w", err)
	}
	f.file = file
	return nil
}

func (f *FileBinaryReader) close() error {
	return f.file.Close()
}

func (f *FileBinaryReader) seek(offset uint64) (uint64, error) {
	if offset > math.MaxInt64 {
		return 0, fmt.Errorf("value %d exceeds int64 maximum limit", offset)
	}
	newOffset, err := f.file.Seek(int64(offset), 0)
	return uint64(newOffset), err
}

func (f *FileBinaryReader) read(p []byte) (int, error) {
	n, err := f.file.Read(p)
	return n, err
}
