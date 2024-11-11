package tiffio

type FastTiffReader struct {
	binary     CachedBinaryReader
	offsets    []uint64
	bytesCount []uint64
}

func NewFastTiffReader(binary CachedBinaryReader) *FastTiffReader {
	return &FastTiffReader{binary: binary}
}
