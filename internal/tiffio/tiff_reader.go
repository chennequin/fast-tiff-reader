package tiffio

import (
	"TiffReader/internal/tiffio/model"
	"TiffReader/internal/tiffio/tags"
	"bytes"
	"context"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"log/slog"
)

const LogLevelTrace = -5

const (
	averageNumberOfTags = 20
	TiffHeaderSize      = 16
	TiffTagSize         = 12
	TiffOffsetSize      = 4
	BigTiffTagSize      = 20
	BigTiffOffsetSize   = 8
)

var LittleEndianSignature = []byte{0x49, 0x49}
var BigEndianSignature = []byte{0x4d, 0x4d}
var TiffMarker = []byte{0x2a, 0x00}
var BigTiffMarker = []byte{0x2b, 0x00}

// TiffReader is a structure that provides methods to read TIFF files.
// It contains a BinaryReader for reading binary data.
type TiffReader struct {
	binary    CachedBinaryReader
	isBigTiff bool
	byteOrder binary.ByteOrder
}

// NewTiffReader creates and returns a new instance of TiffReader,
// initialized with the provided BinaryReader.
func NewTiffReader(binary CachedBinaryReader) *TiffReader {
	return &TiffReader{
		binary: binary,
	}
}

// Open opens the TIFF file specified by the name using the underlying BinaryReader.
func (r *TiffReader) Open(name string) error {
	return r.binary.open(name)
}

// Close closes the TIFF file and releases any resources used by the BinaryReader.
// If an error occurs during closing, it logs a fatal error.
func (r *TiffReader) Close() {
	err := r.binary.close()
	if err != nil {
		log.Fatalf("unable to close reader")
	}
}

// ReadMetaData reads the TIFF metadata from the image file.
// It returns a TIFFMetadata structure containing the entries found.
// In case of errors during reading, it returns an error with context.
func (r *TiffReader) ReadMetaData() (model.TIFFMetadata, error) {
	nextOffset, err := r.readHeader()
	if err != nil {
		return model.TIFFMetadata{}, fmt.Errorf("unable to read header: %s", err)
	}

	entries := make([]model.TIFFDirectory, 0)
	for nextOffset != 0 {
		var ifd model.TIFFDirectory
		if r.isBigTiff {
			ifd, nextOffset, err = r.readBigIFD(nextOffset)
		} else {
			ifd, nextOffset, err = r.readIFD(nextOffset)
		}
		if err != nil {
			return model.TIFFMetadata{}, fmt.Errorf("unable to read IDF: %s", err)
		}
		entries = append(entries, ifd)
		slog.Debug("Metadata", "IFD", ifd)
	}

	return model.NewTIFFMetadata(entries), nil
}

// GetTileData retrieves the tile data for a specific level and tile index from the TIFF image.
func (r *TiffReader) GetTileData(img model.TIFFMetadata, levelIdx, tileIdx int) ([]byte, error) {
	level, err := img.Level(levelIdx)
	if err != nil {
		return nil, err
	}
	tileOffsetTag, err := level.Tag(tags.TileOffsets)
	if err != nil {
		return nil, err
	}
	tileBytesCountTag, err := level.Tag(tags.TileByteCounts)
	if err != nil {
		return nil, err
	}

	if tileIdx >= tileOffsetTag.ValuesCount() {
		return nil, errors.New(fmt.Sprintf("invalid tileIdx: %d", tileIdx))
	}

	tileOffset := tileOffsetTag.GetUintVal(tileIdx)
	tileBytesCount := tileBytesCountTag.GetUintVal(tileIdx)

	data, err := r.readBytesAt(tileOffset, tileBytesCount)
	if err != nil {
		return nil, fmt.Errorf("GetTile: cannot read tile at level %d, tile %d: %w", level, tileIdx, err)
	}

	return data, nil
}

// GetStripData retrieves the strip data for a specific level and strip index from the TIFF image.
func (r *TiffReader) GetStripData(img model.TIFFMetadata, levelIdx, stripIdx int) ([]byte, error) {
	level, err := img.Level(levelIdx)
	if err != nil {
		return nil, err
	}
	stripOffsetTag, err := level.Tag(tags.StripOffsets)
	if err != nil {
		return nil, err
	}
	stripBytesCountTag, err := level.Tag(tags.StripByteCounts)
	if err != nil {
		return nil, err
	}

	if stripIdx >= stripOffsetTag.ValuesCount() {
		return nil, errors.New(fmt.Sprintf("invalid stripIdx: %d", stripIdx))
	}

	stripOffset := stripOffsetTag.GetUintVal(stripIdx)
	stripBytesCount := stripBytesCountTag.GetUintVal(stripIdx)

	data, err := r.readBytesAt(stripOffset, stripBytesCount)
	if err != nil {
		return nil, fmt.Errorf("GetTile: cannot read strip at level %d, strip %d: %w", level, stripIdx, err)
	}

	return data, nil
}

// --------------------------
// decode TIFF binary blocks
// --------------------------

func (r *TiffReader) readHeader() (uint64, error) {
	buffer, err := r.readBytesAt(0, TiffHeaderSize)
	if err != nil {
		return 0, fmt.Errorf("cannot seek to header: %w", err)
	}

	slog.Debug("readHeader", "hex", hex.EncodeToString(buffer[:TiffHeaderSize]))

	if bytes.Equal(buffer[:2], BigEndianSignature) {
		r.byteOrder = binary.BigEndian
		slog.Debug("Byte-order is big-endian.")
	} else if bytes.Equal(buffer[:2], LittleEndianSignature) {
		r.byteOrder = binary.LittleEndian
		slog.Debug("Byte-order is little-endian.")
	} else {
		return 0, errors.New(fmt.Sprintf("Unknown TIFF header: %s", hex.EncodeToString(buffer[:4])))
	}

	if bytes.Equal(buffer[2:4], TiffMarker) {
		slog.Debug("Tiff format")
		r.isBigTiff = false
		nextIFD := uint64(r.byteOrder.Uint32(buffer[4:8]))
		return nextIFD, nil
	}

	if bytes.Equal(buffer[2:4], BigTiffMarker) {
		slog.Debug("BigTiff format")
		r.isBigTiff = true
		offsetSize := r.byteOrder.Uint32(buffer[4:8])
		if offsetSize != 8 {
			return 0, errors.New(fmt.Sprintf("BigTiff size of offsets not supported: %d", offsetSize))
		}
		nextIFD := r.byteOrder.Uint64(buffer[8:16])
		return nextIFD, nil
	}

	return 0, errors.New(fmt.Sprintf("Not a TIFF header: %s", hex.EncodeToString(buffer[:8])))
}

func (r *TiffReader) readIFD(offset uint64) (model.TIFFDirectory, uint64, error) {
	predictedSize := uint64(2 + averageNumberOfTags*TiffTagSize)
	if err := r.binary.readBlock(offset, predictedSize); err != nil {
		return model.TIFFDirectory{}, 0, fmt.Errorf("readBigIFD: cannot read block: %w", err)
	}

	buffer, err := r.readBytesAt(offset, 2)
	if err != nil {
		return model.TIFFDirectory{}, 0, fmt.Errorf("readIFD: cannot read: %w", err)
	}

	// read number of tags
	nbTags := uint64(r.byteOrder.Uint16(buffer[:2]))
	offset += 2

	// complete an un-complete pre-read
	if realSize := 8 + nbTags*TiffTagSize; realSize > predictedSize {
		if err = r.binary.readBlock(offset+predictedSize, realSize-predictedSize); err != nil {
			return model.TIFFDirectory{}, 0, fmt.Errorf("readBigIFD: cannot read block: %w", err)
		}
	}

	// read all tags
	tagMap := make(map[tags.TagID]model.TIFFTag, nbTags)
	for range nbTags {
		tag, err := r.readTag(offset)
		if err != nil {
			return model.TIFFDirectory{}, 0, fmt.Errorf("readIFD: cannot read Tag: %w", err)
		}
		tagMap[tag.GetTagID()] = tag
		offset += TiffTagSize
	}

	// offset to next IDF
	nextOffset, err := r.read4BytesOffsetAt(offset)
	if err != nil {
		return model.TIFFDirectory{}, 0, fmt.Errorf("readIFD: cannot read offset: %w", err)
	}
	return model.NewTIFFDirectory(tagMap), nextOffset, nil
}

func (r *TiffReader) readBigIFD(offset uint64) (model.TIFFDirectory, uint64, error) {
	predictedSize := uint64(8 + averageNumberOfTags*BigTiffTagSize)
	if err := r.binary.readBlock(offset, predictedSize); err != nil {
		return model.TIFFDirectory{}, 0, fmt.Errorf("readBigIFD: cannot read block: %w", err)
	}

	buffer, err := r.readBytesAt(offset, 8)
	if err != nil {
		return model.TIFFDirectory{}, 0, fmt.Errorf("readBigIFD: cannot read: %w", err)
	}

	// read number of tags
	nbTags := r.byteOrder.Uint64(buffer[:8])
	offset += 8

	// complete an un-complete pre-read
	if realSize := 8 + nbTags*BigTiffTagSize; realSize > predictedSize {
		if err = r.binary.readBlock(offset+predictedSize, realSize-predictedSize); err != nil {
			return model.TIFFDirectory{}, 0, fmt.Errorf("readBigIFD: cannot read block: %w", err)
		}
	}

	// read tags
	tagMap := make(map[tags.TagID]model.TIFFTag, nbTags)
	for range nbTags {
		tag, err := r.readBigTag(offset)
		if err != nil {
			return model.TIFFDirectory{}, 0, fmt.Errorf("readBigIFD: cannot read Tag: %w", err)
		}
		tagMap[tag.GetTagID()] = tag
		offset += BigTiffTagSize
	}

	// offset to next IDF
	nextOffset, err := r.read8BytesOffsetAt(offset)
	if err != nil {
		return model.TIFFDirectory{}, 0, fmt.Errorf("readBigIFD: cannot read offset: %w", err)
	}
	return model.NewTIFFDirectory(tagMap), nextOffset, nil
}

func (r *TiffReader) isValidSize(numValues, size uint64) bool {
	if r.isBigTiff {
		return numValues*size <= BigTiffOffsetSize
	}
	return numValues*size <= TiffOffsetSize
}

func (r *TiffReader) offsetFrom(buffer []byte) uint64 {
	if r.isBigTiff {
		return r.byteOrder.Uint64(buffer[0:BigTiffOffsetSize])
	}
	return uint64(r.byteOrder.Uint32(buffer[0:TiffOffsetSize]))
}

func (r *TiffReader) readTagValues(buffer []byte, tagID, tagType uint16, numValues uint64) (model.TIFFTag, error) {
	switch tagType {
	// byte, undefined
	case 0x1, 0x7:
		values, err := r.readValuesBytes(buffer, numValues)
		tag := model.DataTag[byte]{TagID: tags.TagID(tagID), Values: values}
		return tag, err

	// ASCII (nul terminated \0) string
	case 0x2:
		values, err := r.readValuesString(buffer, numValues)
		tag := model.DataTag[string]{TagID: tags.TagID(tagID), Values: values}
		return tag, err

	// short
	case 0x3:
		values, err := r.readValuesShorts(buffer, numValues)
		tag := model.DataTag[uint16]{TagID: tags.TagID(tagID), Values: values}
		return tag, err

	// long
	case 0x4:
		values, err := r.readValuesLongs(buffer, numValues)
		tag := model.DataTag[uint32]{TagID: tags.TagID(tagID), Values: values}
		return tag, err

	// rational
	case 0x5:
		values, err := r.readValuesRationals(buffer, numValues)
		tag := model.DataTag[model.Rational]{TagID: tags.TagID(tagID), Values: values}
		return tag, err

	// signed byte
	case 0x6:
		values, err := r.readValuesSignedBytes(buffer, numValues)
		tag := model.DataTag[int8]{TagID: tags.TagID(tagID), Values: values}
		return tag, err

	// signed short
	case 0x8:
		values, err := r.readValuesSignedShorts(buffer, numValues)
		tag := model.DataTag[int16]{TagID: tags.TagID(tagID), Values: values}
		return tag, err

	// signed long
	case 0x9:
		values, err := r.readValuesSignedLongs(buffer, numValues)
		tag := model.DataTag[int32]{TagID: tags.TagID(tagID), Values: values}
		return tag, err

	// signed rational
	case 0xa:
		values, err := r.readValuesSignedRationals(buffer, numValues)
		tag := model.DataTag[model.SignedRational]{TagID: tags.TagID(tagID), Values: values}
		return tag, err

	// float
	case 0xb:
		values, err := r.readValuesFloats(buffer, numValues)
		tag := model.DataTag[float32]{TagID: tags.TagID(tagID), Values: values}
		return tag, err

	// double
	case 0xc:
		values, err := r.readValuesDoubles(buffer, numValues)
		tag := model.DataTag[float64]{TagID: tags.TagID(tagID), Values: values}
		return tag, err

	// LONG8
	case 0x10:
		values, err := r.readValuesLong8s(buffer, numValues)
		tag := model.DataTag[uint64]{TagID: tags.TagID(tagID), Values: values}
		return tag, err

	// SIGNED LONG8
	case 0x11:
		values, err := r.readValuesSignedLong8s(buffer, numValues)
		tag := model.DataTag[int64]{TagID: tags.TagID(tagID), Values: values}
		return tag, err
	}

	return nil, errors.New(fmt.Sprintf("unknown tag type: %d", tagType))
}

func (r *TiffReader) readTag(offset uint64) (model.TIFFTag, error) {
	buffer, err := r.readBytesAt(offset, TiffTagSize)
	if err != nil {
		return nil, fmt.Errorf("readTag: cannot read: %w", err)
	}

	//slog.Debug("readTag", "hex", hex.EncodeToString(buffer[:TiffTagSize]))
	slog.Log(context.Background(), LogLevelTrace, "readTag", "hex", hex.EncodeToString(buffer[:TiffTagSize]))

	tagID := r.byteOrder.Uint16(buffer[:2])
	tagType := r.byteOrder.Uint16(buffer[2:4])
	numValues := uint64(r.byteOrder.Uint32(buffer[4:8]))

	tag, err := r.readTagValues(buffer[8:], tagID, tagType, numValues)
	if err != nil {
		return nil, fmt.Errorf("readTag: cannot read values: %w", err)
	}
	return tag, nil
}

func (r *TiffReader) readBigTag(offset uint64) (model.TIFFTag, error) {
	buffer, err := r.readBytesAt(offset, BigTiffTagSize)
	if err != nil {
		return nil, fmt.Errorf("readBigTag: cannot read: %w", err)
	}

	slog.Debug("readBigTag", "hex", hex.EncodeToString(buffer[:BigTiffTagSize]))

	tagID := r.byteOrder.Uint16(buffer[:2])
	tagType := r.byteOrder.Uint16(buffer[2:4])
	numValues := r.byteOrder.Uint64(buffer[4:12])

	tag, err := r.readTagValues(buffer[12:], tagID, tagType, numValues)
	if err != nil {
		return nil, fmt.Errorf("readTag: cannot read values: %w", err)
	}
	return tag, nil
}

// --------------------------
// reading bytes
// --------------------------

func (r *TiffReader) read4BytesOffsetAt(offset uint64) (uint64, error) {
	buffer, err := r.readBytesAt(offset, 4)
	if err != nil {
		return 0, err
	}
	nextOffset := uint64(r.byteOrder.Uint32(buffer))
	return nextOffset, nil
}

func (r *TiffReader) read8BytesOffsetAt(offset uint64) (uint64, error) {
	buffer, err := r.readBytesAt(offset, 8)
	if err != nil {
		return 0, err
	}
	nextOffset := r.byteOrder.Uint64(buffer)
	return nextOffset, nil
}

func (r *TiffReader) readBytesAt(offset uint64, n uint64) ([]byte, error) {
	buffer := make([]byte, n)
	bytesRead, err := r.binary.read(offset, buffer)
	if err != nil {
		return nil, fmt.Errorf("cannot read %d bytes at %d: %w", n, offset, err)
	}
	if bytesRead != int(n) {
		return nil, fmt.Errorf("unexpected end when reading: expected %d, got %d", n, bytesRead)
	}

	return buffer, nil
}

// --------------------------
// bytes conversion
// --------------------------

func (r *TiffReader) bytesToUint16(data []byte) uint16 {
	return r.byteOrder.Uint16(data)
}

func (r *TiffReader) bytesToUint32(data []byte) uint32 {
	return r.byteOrder.Uint32(data)
}

func (r *TiffReader) bytesToUint64(data []byte) uint64 {
	return r.byteOrder.Uint64(data)
}

func (r *TiffReader) bytesToInt8(data []byte) int8 {
	return int8(data[0])
}

func (r *TiffReader) bytesToInt16(data []byte) int16 {
	return int16(r.byteOrder.Uint16(data[0:2]))
}

func (r *TiffReader) bytesToInt32(data []byte) int32 {
	return int32(r.byteOrder.Uint32(data[0:4]))
}

func (r *TiffReader) bytesToInt64(data []byte) int64 {
	return int64(r.byteOrder.Uint64(data[0:8]))
}

func (r *TiffReader) bytesToFloat32(data []byte) float32 {
	return float32(r.byteOrder.Uint64(data))
}

func (r *TiffReader) bytesToFloat64(data []byte) float64 {
	return float64(r.byteOrder.Uint64(data))
}

func (r *TiffReader) bytesToRational(data []byte) model.Rational {
	return model.Rational{
		Numerator:   r.byteOrder.Uint32(data[0:4]),
		Denominator: r.byteOrder.Uint32(data[4:8]),
	}
}

func (r *TiffReader) bytesToSignedRational(data []byte) model.SignedRational {
	return model.SignedRational{
		Numerator:   int32(r.byteOrder.Uint32(data[0:4])),
		Denominator: int32(r.byteOrder.Uint32(data[4:8])),
	}
}

// --------------------------
// TIFF values logic
// --------------------------

func (r *TiffReader) readValuesBytes(buffer []byte, numValues uint64) ([]byte, error) {
	if r.isValidSize(numValues, 1) {
		return buffer[0:numValues], nil
	}
	// TODO in //
	rOffset := r.offsetFrom(buffer)
	values, err := r.readBytesAt(rOffset, numValues)
	return values, err
}

func (r *TiffReader) readValuesString(buffer []byte, numValues uint64) ([]string, error) {
	if r.isValidSize(numValues, 1) {
		return []string{string(buffer[0:numValues])}, nil
	}
	rOffset := r.offsetFrom(buffer)
	values, err := r.readBytesAt(rOffset, numValues)
	return []string{string(values)}, err
}

func (r *TiffReader) readValuesShorts(buffer []byte, numValues uint64) ([]uint16, error) {
	return readValuesFn(r, buffer, numValues, 2, r.bytesToUint16)
}

func (r *TiffReader) readValuesLongs(buffer []byte, numValues uint64) ([]uint32, error) {
	return readValuesFn(r, buffer, numValues, 4, r.bytesToUint32)
}

func (r *TiffReader) readValuesLong8s(buffer []byte, numValues uint64) ([]uint64, error) {
	return readValuesFn(r, buffer, numValues, 8, r.bytesToUint64)
}

func (r *TiffReader) readValuesSignedBytes(buffer []byte, numValues uint64) ([]int8, error) {
	return readValuesFn(r, buffer, numValues, 1, r.bytesToInt8)
}

func (r *TiffReader) readValuesSignedShorts(buffer []byte, numValues uint64) ([]int16, error) {
	return readValuesFn(r, buffer, numValues, 2, r.bytesToInt16)
}

func (r *TiffReader) readValuesSignedLongs(buffer []byte, numValues uint64) ([]int32, error) {
	return readValuesFn(r, buffer, numValues, 4, r.bytesToInt32)
}

func (r *TiffReader) readValuesSignedLong8s(buffer []byte, numValues uint64) ([]int64, error) {
	return readValuesFn(r, buffer, numValues, 8, r.bytesToInt64)
}

func (r *TiffReader) readValuesFloats(buffer []byte, numValues uint64) ([]float32, error) {
	return readValuesFn(r, buffer, numValues, 4, r.bytesToFloat32)
}

func (r *TiffReader) readValuesDoubles(buffer []byte, numValues uint64) ([]float64, error) {
	return readValuesFn(r, buffer, numValues, 8, r.bytesToFloat64)
}

func (r *TiffReader) readValuesRationals(buffer []byte, numValues uint64) ([]model.Rational, error) {
	return readValuesFn(r, buffer, numValues, 8, r.bytesToRational)
}

func (r *TiffReader) readValuesSignedRationals(buffer []byte, numValues uint64) ([]model.SignedRational, error) {
	return readValuesFn(r, buffer, numValues, 8, r.bytesToSignedRational)
}

func readValuesFn[T model.TagType](r *TiffReader, buffer []byte, numValues, size uint64, fromBytesFn func([]byte) T) ([]T, error) {
	if r.isValidSize(numValues, size) {
		values := make([]T, numValues)
		for i := uint64(0); i < numValues; i++ {
			ofs := i * size
			values[i] = fromBytesFn(buffer[ofs : ofs+size])
		}
		return values, nil
	}
	rOffset := r.offsetFrom(buffer)
	values, err := readValuesFnAt(r, rOffset, numValues, size, fromBytesFn)
	return values, err
}

func readValuesFnAt[T model.TagType](r *TiffReader, offset uint64, numValues uint64, elementSize uint64, fromBytesFn func([]byte) T) ([]T, error) {
	buffer, err := r.readBytesAt(offset, numValues*elementSize)
	if err != nil {
		return nil, fmt.Errorf("readValuesAt: cannot read: %w", err)
	}

	vOffset := uint64(0)
	values := make([]T, numValues)
	for i := range numValues {
		values[i] = fromBytesFn(buffer[vOffset : vOffset+elementSize])
		vOffset += elementSize
	}

	return values, nil
}
