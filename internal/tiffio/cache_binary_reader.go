package tiffio

import "sync"

const IFDCommonSize = 20 * 20

type CacheBinaryReader struct {
	binary   BinaryReader
	offset   uint64
	segments map[uint64][]byte
	lock     sync.RWMutex
}

func NewCacheBinaryReader(binary BinaryReader) CachedBinaryReader {
	return &CacheBinaryReader{
		binary:   binary,
		segments: make(map[uint64][]byte),
	}
}

func (f *CacheBinaryReader) open(name string) error {
	return f.binary.open(name)
}

func (f *CacheBinaryReader) close() error {
	return f.binary.close()
}

func (f *CacheBinaryReader) openMetaData() {
}

func (f *CacheBinaryReader) closeMetaData() {
	clear(f.segments)
}

func (f *CacheBinaryReader) readBlock(offset, size uint64) error {
	buffer := make([]byte, size)
	_, err := f.binary.read(offset, buffer)
	f.lock.Lock()
	defer f.lock.Unlock()
	f.segments[offset] = buffer
	return err
}

func (f *CacheBinaryReader) read(offset uint64, p []byte) (int, error) {
	//if f.within(offset, p) {
	//	return len(p), nil
	//}
	return f.binary.read(offset, p)
}

func (f *CacheBinaryReader) within(offset uint64, p []byte) bool {
	f.lock.RLock()
	defer f.lock.RUnlock()
	for segmentStart, buffer := range f.segments {
		endOffset := segmentStart + uint64(len(buffer))
		if offset >= segmentStart && offset+uint64(len(p)) <= endOffset {
			b := int(offset - segmentStart)
			copy(p, buffer[b:b+len(p)])
			return true
		}
	}
	return false
}

//func (f *CacheBinaryReader) readIFD(size uint64, p []byte) (int, error) {
//	buffer := make([]byte, size)
//	n, err := f.binary.read(buffer)
//	copy(p, buffer[:len(p)])
//	bytesRead := min(n, len(p))
//	return bytesRead, err
//}
