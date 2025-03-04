package tiffio

import (
	"fmt"
	"log/slog"
	"math"
	"os"
	"sync"
)

type FileBinaryReader struct {
	file *os.File
	lock sync.Mutex
}

func NewFileBinaryReader() BinaryReader {
	return &FileBinaryReader{}
}

func (f *FileBinaryReader) open(name string) error {
	f.lock.Lock()
	defer f.lock.Unlock()
	if f.file != nil {
		if err := f.close(); err != nil {
			slog.Warn("error closing file", err)
		}
	}
	file, err := os.Open(name)
	if err != nil {
		return fmt.Errorf("unable to open file %w", err)
	}
	f.file = file
	return nil
}

func (f *FileBinaryReader) close() error {
	slog.Info("closing file", "file", f.file.Name())
	return f.file.Close()
}

func (f *FileBinaryReader) read(offset uint64, p []byte) (int, error) {
	f.lock.Lock()
	defer f.lock.Unlock()
	if offset > math.MaxInt64 {
		return 0, fmt.Errorf("value %d exceeds int64 maximum limit", offset)
	}
	_, err := f.file.Seek(int64(offset), 0)
	n, err := f.file.Read(p)
	return n, err
}
