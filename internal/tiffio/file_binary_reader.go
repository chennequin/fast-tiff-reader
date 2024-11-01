package tiffio

import (
	"fmt"
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

func (f *FileBinaryReader) seek(offset int64) (int64, error) {
	newOffset, err := f.file.Seek(offset, 0)
	return newOffset, err
}

func (f *FileBinaryReader) read(p []byte) (int, error) {
	n, err := f.file.Read(p)
	return n, err
}
