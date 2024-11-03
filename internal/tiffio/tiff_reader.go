package tiffio

import (
	"TiffReader/internal/tiffio/model"
	"TiffReader/internal/tiffio/tags"
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"strings"
)

const (
	TiffOffset    = 4
	BigTiffOffset = 8
)

var LittleEndianSignature = []byte{0x49, 0x49}
var BigEndianSignature = []byte{0x4d, 0x4d}
var TiffMarker = []byte{0x2a, 0x00}
var BigTiffMarker = []byte{0x2b, 0x00}

type TiffReader struct {
	binary    BinaryReader
	isBigTiff bool
	byteOrder binary.ByteOrder
}

func NewTiffReader(binary BinaryReader) *TiffReader {
	return &TiffReader{
		binary: binary,
	}
}

func (r *TiffReader) Open(name string) error {
	return r.binary.open(name)
}

func (r *TiffReader) Close() {
	err := r.binary.close()
	if err != nil {
		log.Fatalf("unable to close reader")
	}
}

func (r *TiffReader) ReadMetaData() (model.TIFF, error) {
	nextIFD, err := r.ReadHeader()
	if err != nil {
		return model.TIFF{}, fmt.Errorf("unable to read header: %s", err)
	}

	imageFileDirectories := make([]model.IFD, 0)
	for nextIFD != 0 {
		if r.isBigTiff {
			ifd, err := r.ReadBigIFD(nextIFD)
			if err != nil {
				return model.TIFF{}, fmt.Errorf("unable to read IDF: %s", err)
			}
			fmt.Printf("%s\n", ifd)
			imageFileDirectories = append(imageFileDirectories, ifd)
			nextIFD = ifd.NextIFD
		} else {
			ifd, err := r.ReadIFD(nextIFD)
			if err != nil {
				return model.TIFF{}, fmt.Errorf("unable to read IDF: %s", err)
			}
			fmt.Printf("%s\n", ifd)
			imageFileDirectories = append(imageFileDirectories, ifd)
			nextIFD = ifd.NextIFD
		}
	}

	tiffImg := model.TIFF{
		IFDs: imageFileDirectories,
	}
	return tiffImg, nil
}

func (r *TiffReader) ReadHeader() (uint64, error) {
	_, err := r.binary.seek(0)
	if err != nil {
		return 0, fmt.Errorf("cannot seek to header: %w", err)
	}

	buffer := make([]byte, 16)
	_, err = r.binary.read(buffer)
	if err != nil {
		return 0, fmt.Errorf("cannot read header: %w", err)
	}

	slog.Debug("header", "hex", hex.EncodeToString(buffer[:16]))

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
		if r.byteOrder.Uint32(buffer[4:8]) != 8 {
			return 0, errors.New(fmt.Sprintf("BigTiff Bytesize of offsets not supported"))
		}
		nextIFD := r.byteOrder.Uint64(buffer[8:16])
		return nextIFD, nil
	}

	return 0, errors.New(fmt.Sprintf("Not a TIFF header: %s", hex.EncodeToString(buffer[:4])))
}

func (r *TiffReader) ReadIFD(offset uint64) (model.IFD, error) {
	buffer, err := r.read2BytesAt(offset)
	if err != nil {
		return model.IFD{}, fmt.Errorf("ReadIFD: cannot read: %w", err)
	}

	// read number of tags
	nbTags := uint64(r.byteOrder.Uint16(buffer[:2]))
	offset += 2

	// read tags
	tagsInIFD := make(map[tags.TagID]model.Tag, nbTags)
	for range nbTags {
		tag, err := r.ReadTag(offset)
		if err != nil {
			return model.IFD{}, fmt.Errorf("ReadIFD: cannot read TAG: %w", err)
		}
		tagsInIFD[tag.GetTagID()] = tag
		offset += 12
	}

	// offset to next IDF
	nextIFD, err := r.read4BytesOffsetAt(offset)
	if err != nil {
		return model.IFD{}, fmt.Errorf("ReadIFD: cannot read offset: %w", err)
	}
	ifd := model.IFD{
		NextIFD: nextIFD,
		Tags:    tagsInIFD,
	}
	return ifd, nil
}

func (r *TiffReader) ReadBigIFD(offset uint64) (model.IFD, error) {
	buffer, err := r.read8BytesAt(offset)
	if err != nil {
		return model.IFD{}, fmt.Errorf("ReadIFD: cannot read: %w", err)
	}

	// read number of tags
	nbTags := r.byteOrder.Uint64(buffer[:8])
	offset += 8

	// read tags
	tagsInIFD := make(map[tags.TagID]model.Tag, nbTags)
	for range nbTags {
		tag, err := r.ReadBigTag(offset)
		if err != nil {
			return model.IFD{}, fmt.Errorf("ReadIFD: cannot read TAG: %w", err)
		}
		tagsInIFD[tag.GetTagID()] = tag
		offset += 20
	}

	// offset to next IDF
	nextIFD, err := r.read8BytesOffsetAt(offset)
	if err != nil {
		return model.IFD{}, fmt.Errorf("ReadIFD: cannot read offset: %w", err)
	}
	ifd := model.IFD{
		NextIFD: nextIFD,
		Tags:    tagsInIFD,
	}
	return ifd, nil
}

func (r *TiffReader) ReadTag(offset uint64) (model.Tag, error) {
	buffer, err := r.readBytesAt(offset, 12)
	if err != nil {
		return nil, fmt.Errorf("ReadTag: cannot read: %w", err)
	}

	slog.Debug("header", "hex", hex.EncodeToString(buffer[:12]))

	tagID := r.byteOrder.Uint16(buffer[:2])
	tagType := r.byteOrder.Uint16(buffer[2:4])
	numValues := uint64(r.byteOrder.Uint32(buffer[4:8]))

	buffer = buffer[8:]

	switch tagType {

	// byte
	case 0x1:
	// undefined
	case 0x7:
		if numValues <= 4 {
			tag := model.DataTag[byte]{
				TagID:  tagID,
				Values: buffer[0:int(numValues)],
			}
			return tag, nil
		}

		rOffset := uint64(r.byteOrder.Uint32(buffer[0:4]))
		values, err := r.readBytesAt(rOffset, numValues)
		if err != nil {
			return nil, err
		}
		tag := model.DataTag[byte]{
			TagID:  tagID,
			Values: values,
		}
		return tag, nil

	// ASCII (nul terminated \0) string
	case 0x2:
		rOffset := uint64(r.byteOrder.Uint32(buffer[0:4]))
		values, err := r.readStringValuesAt(rOffset, numValues)
		if err != nil {
			return nil, err
		}
		tag := model.DataTag[string]{
			TagID:  tagID,
			Values: values,
		}
		return tag, nil

	// short
	case 0x3:
		var values []uint16
		if numValues == 1 {
			values = []uint16{r.byteOrder.Uint16(buffer[0:2])}
		}
		if numValues == 2 {
			values = []uint16{
				r.byteOrder.Uint16(buffer[0:2]),
				r.byteOrder.Uint16(buffer[2:4]),
			}
		}
		if numValues > 2 {
			rOffset := uint64(r.byteOrder.Uint32(buffer[0:4]))
			values, err = readValuesAt(r, rOffset, numValues, 2, r.bytesToUint16)
			if err != nil {
				return nil, err
			}
		}
		tag := model.DataTag[uint16]{
			TagID:  tagID,
			Values: values,
		}
		return tag, nil

	// long
	case 0x4:
		var values []uint32
		if numValues == 1 {
			values = []uint32{r.byteOrder.Uint32(buffer[0:4])}
		}
		if numValues > 1 {
			rOffset := uint64(r.byteOrder.Uint32(buffer[0:4]))
			values, err = readValuesAt(r, rOffset, numValues, 4, r.bytesToUint32)
			if err != nil {
				return nil, err
			}
		}
		tag := model.DataTag[uint32]{
			TagID:  tagID,
			Values: values,
		}
		return tag, nil

	// rational
	case 0x5:
		var values []model.Rational
		rOffset := uint64(r.byteOrder.Uint32(buffer[0:4]))
		values, err = readValuesAt(r, rOffset, numValues, 8, r.bytesToRational)
		if err != nil {
			return nil, err
		}
		tag := model.DataTag[model.Rational]{
			TagID:  tagID,
			Values: values,
		}
		return tag, nil

	// signed byte
	case 0x6:
		var values []int8
		if numValues <= 4 {
			values = make([]int8, numValues)
			for i := range numValues {
				values[i] = int8(buffer[i : i+1][0])
			}
		}
		if numValues > 4 {
			rOffset := uint64(r.byteOrder.Uint32(buffer[0:4]))
			values, err = readValuesAt(r, rOffset, numValues, 1, r.bytesToInt8)
			if err != nil {
				return nil, err
			}
		}
		tag := model.DataTag[int8]{
			TagID:  tagID,
			Values: values,
		}
		return tag, nil

	// signed short
	case 0x8:
		var values []int16
		if numValues <= 2 {
			values = make([]int16, numValues)
			for i := range numValues {
				values[i] = r.bytesToInt16(buffer[i : i+2])
			}
		} else {
			rOffset := uint64(r.byteOrder.Uint32(buffer[0:4]))
			values, err = readValuesAt(r, rOffset, numValues, 2, r.bytesToInt16)
			if err != nil {
				return nil, err
			}
		}
		tag := model.DataTag[int16]{
			TagID:  tagID,
			Values: values,
		}
		return tag, nil

	// signed long
	case 0x9:
		var values []int32
		if numValues <= 1 {
			values = []int32{r.bytesToInt32(buffer[0:4])}
		} else {
			rOffset := uint64(r.byteOrder.Uint32(buffer[0:4]))
			values, err = readValuesAt(r, rOffset, numValues, 4, r.bytesToInt32)
			if err != nil {
				return nil, err
			}
		}
		tag := model.DataTag[int32]{
			TagID:  tagID,
			Values: values,
		}
		return tag, nil

	// signed rational
	case 0xa:
		var values []model.SignedRational
		rOffset := uint64(r.byteOrder.Uint32(buffer[0:4]))
		values, err = readValuesAt(r, rOffset, numValues, 8, r.bytesToSignedRational)
		if err != nil {
			return nil, err
		}
		tag := model.DataTag[model.SignedRational]{
			TagID:  tagID,
			Values: values,
		}
		return tag, nil

	// float
	case 0xb:
		var values []float32
		if numValues <= 1 {
			values = []float32{r.bytesToFloat32(buffer[0:4])}
		} else {
			rOffset := uint64(r.byteOrder.Uint32(buffer[0:4]))
			values, err = readValuesAt(r, rOffset, numValues, 4, r.bytesToFloat32)
			if err != nil {
				return nil, err
			}
		}
		tag := model.DataTag[float32]{
			TagID:  tagID,
			Values: values,
		}
		return tag, nil

	// double
	case 0xc:
		var values []float64
		rOffset := uint64(r.byteOrder.Uint32(buffer[0:4]))
		values, err = readValuesAt(r, rOffset, numValues, 8, r.bytesToFloat64)
		if err != nil {
			return nil, err
		}
		tag := model.DataTag[float64]{
			TagID:  tagID,
			Values: values,
		}
		return tag, nil

	default:
		return nil, errors.New(fmt.Sprintf("Unknown tag type: %d", tagType))
	}

	return nil, nil
}

func (r *TiffReader) ReadBigTag(offset uint64) (model.Tag, error) {
	buffer, err := r.readBytesAt(offset, 20)
	if err != nil {
		return nil, fmt.Errorf("ReadTag: cannot read: %w", err)
	}

	slog.Debug("header", "hex", hex.EncodeToString(buffer[:20]))

	tagID := r.byteOrder.Uint16(buffer[:2])
	tagType := r.byteOrder.Uint16(buffer[2:4])
	numValues := r.byteOrder.Uint64(buffer[4:12])

	buffer = buffer[12:]

	switch tagType {

	// BYTE
	case 0x1:
	// UNDEFINED
	case 0x7:
		if numValues <= 8 {
			tag := model.DataTag[byte]{
				TagID:  tagID,
				Values: buffer[0:int(numValues)],
			}
			return tag, nil
		}

		rOffset := r.byteOrder.Uint64(buffer[0:8])
		values, err := r.readBytesAt(rOffset, numValues)
		if err != nil {
			return nil, err
		}
		tag := model.DataTag[byte]{
			TagID:  tagID,
			Values: values,
		}
		return tag, nil

	// ASCII (nul terminated \0) string
	case 0x2:
		rOffset := r.byteOrder.Uint64(buffer[0:8])
		values, err := r.readStringValuesAt(rOffset, numValues)
		if err != nil {
			return nil, err
		}
		tag := model.DataTag[string]{
			TagID:  tagID,
			Values: values,
		}
		return tag, nil

	// SHORT
	case 0x3:
		var values []uint16
		if numValues <= 4 {
			values = make([]uint16, numValues)
			for i := 0; i < int(numValues); i++ {
				ofs := i * 2
				values[i] = r.byteOrder.Uint16(buffer[ofs : ofs+2])
			}
		}
		if numValues > 4 {
			rOffset := r.byteOrder.Uint64(buffer[0:8])
			values, err = readValuesAt(r, rOffset, numValues, 2, r.bytesToUint16)
			if err != nil {
				return nil, err
			}
		}
		tag := model.DataTag[uint16]{
			TagID:  tagID,
			Values: values,
		}
		return tag, nil

	// LONG
	case 0x4:
		var values []uint32
		if numValues == 1 {
			values = []uint32{r.byteOrder.Uint32(buffer[0:4])}
		}
		if numValues == 2 {
			values = []uint32{
				r.byteOrder.Uint32(buffer[0:4]),
				r.byteOrder.Uint32(buffer[4:8]),
			}
		}
		if numValues > 2 {
			rOffset := r.byteOrder.Uint64(buffer[0:8])
			values, err = readValuesAt(r, rOffset, numValues, 4, r.bytesToUint32)
			if err != nil {
				return nil, err
			}
		}
		tag := model.DataTag[uint32]{
			TagID:  tagID,
			Values: values,
		}
		return tag, nil

	// RATIONAL
	case 0x5:
		var values []model.Rational
		rOffset := r.byteOrder.Uint64(buffer[0:8])
		values, err = readValuesAt(r, rOffset, numValues, 8, r.bytesToRational)
		if err != nil {
			return nil, err
		}
		tag := model.DataTag[model.Rational]{
			TagID:  tagID,
			Values: values,
		}
		return tag, nil

	// SIGNED BYTE
	case 0x6:
		var values []int8
		if numValues <= 8 {
			values = make([]int8, numValues)
			for i := range numValues {
				values[i] = int8(buffer[i : i+1][0])
			}
		}
		if numValues > 8 {
			rOffset := r.byteOrder.Uint64(buffer[0:8])
			values, err = readValuesAt(r, rOffset, numValues, 1, r.bytesToInt8)
			if err != nil {
				return nil, err
			}
		}
		tag := model.DataTag[int8]{
			TagID:  tagID,
			Values: values,
		}
		return tag, nil

	// SIGNED SHORT
	case 0x8:
		var values []int16
		if numValues <= 4 {
			values = make([]int16, numValues)
			for i := range numValues {
				values[i] = r.bytesToInt16(buffer[i : i+2])
			}
		} else {
			rOffset := uint64(r.byteOrder.Uint32(buffer[0:4]))
			values, err = readValuesAt(r, rOffset, numValues, 2, r.bytesToInt16)
			if err != nil {
				return nil, err
			}
		}
		tag := model.DataTag[int16]{
			TagID:  tagID,
			Values: values,
		}
		return tag, nil

	// SIGNED LONG
	case 0x9:
		var values []int32
		if numValues == 1 {
			values = []int32{r.bytesToInt32(buffer[0:4])}
		}
		if numValues == 2 {
			values = []int32{
				r.bytesToInt32(buffer[0:4]),
				r.bytesToInt32(buffer[4:8]),
			}
		}
		if numValues > 2 {
			rOffset := r.byteOrder.Uint64(buffer[0:8])
			values, err = readValuesAt(r, rOffset, numValues, 4, r.bytesToInt32)
			if err != nil {
				return nil, err
			}
		}
		tag := model.DataTag[int32]{
			TagID:  tagID,
			Values: values,
		}
		return tag, nil

	// SIGNED RATIONAL
	case 0xa:
		var values []model.SignedRational
		rOffset := r.byteOrder.Uint64(buffer[0:8])
		values, err = readValuesAt(r, rOffset, numValues, 8, r.bytesToSignedRational)
		if err != nil {
			return nil, err
		}
		tag := model.DataTag[model.SignedRational]{
			TagID:  tagID,
			Values: values,
		}
		return tag, nil

	// FLOAT
	case 0xb:
		var values []float32
		if numValues <= 1 {
			values = []float32{r.bytesToFloat32(buffer[0:4])}
		}
		if numValues == 2 {
			values = []float32{
				r.bytesToFloat32(buffer[0:4]),
				r.bytesToFloat32(buffer[4:8]),
			}
		}
		if numValues > 2 {
			rOffset := r.byteOrder.Uint64(buffer[0:8])
			values, err = readValuesAt(r, rOffset, numValues, 4, r.bytesToFloat32)
			if err != nil {
				return nil, err
			}
		}
		tag := model.DataTag[float32]{
			TagID:  tagID,
			Values: values,
		}
		return tag, nil

	// DOUBLE
	case 0xc:
		var values []float64
		if numValues == 1 {
			values = []float64{r.bytesToFloat64(buffer[0:8])}
		}
		if numValues > 1 {
			rOffset := r.byteOrder.Uint64(buffer[0:8])
			values, err = readValuesAt(r, rOffset, numValues, 8, r.bytesToFloat64)
			if err != nil {
				return nil, err
			}
		}
		tag := model.DataTag[float64]{
			TagID:  tagID,
			Values: values,
		}
		return tag, nil

	// LONG8
	case 0x10:
		var values []uint64
		if numValues == 1 {
			values = []uint64{r.bytesToUint64(buffer[0:8])}
		}
		if numValues > 1 {
			rOffset := r.byteOrder.Uint64(buffer[0:8])
			values, err = readValuesAt(r, rOffset, numValues, 8, r.bytesToUint64)
			if err != nil {
				return nil, err
			}
		}
		tag := model.DataTag[uint64]{
			TagID:  tagID,
			Values: values,
		}
		return tag, nil

	// SIGNED LONG8
	case 0x11:
		var values []int64
		if numValues == 1 {
			values = []int64{int64(r.bytesToUint64(buffer[0:8]))}
		}
		if numValues > 1 {
			rOffset := r.byteOrder.Uint64(buffer[0:8])
			values, err = readValuesAt(r, rOffset, numValues, 8, r.bytesToInt64)
			if err != nil {
				return nil, err
			}
		}
		tag := model.DataTag[int64]{
			TagID:  tagID,
			Values: values,
		}
		return tag, nil

	default:
		return nil, errors.New(fmt.Sprintf("Unknown tag type: %d", tagType))
	}

	return nil, nil
}

func (r *TiffReader) readStringValuesAt(offset uint64, numValues uint64) ([]string, error) {
	_, seekError := r.binary.seek(offset)
	if seekError != nil {
		return nil, fmt.Errorf("readStringsAt: unable to seek at %d: %w", offset, seekError)
	}

	var sb strings.Builder
	for {
		// read from underlying binary source
		buffer := make([]byte, 256)
		bytesRead, readError := r.binary.read(buffer)
		if readError != nil {
			return nil, fmt.Errorf("readStringsAt: unable to read: %w", readError)
		}

		// extract zero terminated C-style string
		str, _, readError := readCString(buffer[:bytesRead])
		if readError != nil && errors.Is(readError, ZeroNotFound) {
			sb.WriteString(str)
			continue
		}

		sb.WriteString(str)
		break
	}

	return []string{sb.String()}, nil
}

func readValuesAt[T model.TagType](r *TiffReader, offset uint64, numValues uint64, elementSize uint64, fromBytesFn func(data []byte) T) ([]T, error) {
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

func (r *TiffReader) read4BytesOffsetAt(offset uint64) (uint64, error) {
	buffer, err := r.read4BytesAt(offset)
	if err != nil {
		return 0, err
	}
	nextOffset := uint64(r.byteOrder.Uint32(buffer))
	return nextOffset, nil
}

func (r *TiffReader) read8BytesOffsetAt(offset uint64) (uint64, error) {
	buffer, err := r.read8BytesAt(offset)
	if err != nil {
		return 0, err
	}
	nextOffset := r.byteOrder.Uint64(buffer)
	return nextOffset, nil
}

func (r *TiffReader) read2BytesAt(offset uint64) ([]byte, error) {
	return r.readBytesAt(offset, 2)
}

func (r *TiffReader) read4BytesAt(offset uint64) ([]byte, error) {
	return r.readBytesAt(offset, 4)
}

func (r *TiffReader) read8BytesAt(offset uint64) ([]byte, error) {
	return r.readBytesAt(offset, 8)
}

func (r *TiffReader) readBytesAt(offset uint64, n uint64) ([]byte, error) {
	_, err := r.binary.seek(offset)
	if err != nil {
		return nil, fmt.Errorf("cannot seek to %d: %w", offset, err)
	}

	buffer := make([]byte, n)
	bytesRead, err := r.binary.read(buffer)
	if err != nil {
		return nil, fmt.Errorf("cannot read %d bytes at %d: %w", n, offset, err)
	}
	if bytesRead != int(n) {
		return nil, fmt.Errorf("unexpected end when reading: expected %d, got %d", n, bytesRead)
	}

	return buffer, nil
}

func (r *TiffReader) bytesToUint16(data []byte) uint16 {
	return r.byteOrder.Uint16(data)
}

func (r *TiffReader) bytesToUint32(data []byte) uint32 {
	return r.byteOrder.Uint32(data)
}

func (r *TiffReader) bytesToUint64(data []byte) uint64 {
	return r.byteOrder.Uint64(data)
}

func (r *TiffReader) bytesToRational(data []byte) model.Rational {
	return model.Rational{
		Numerator:   r.byteOrder.Uint32(data[0:4]),
		Denominator: r.byteOrder.Uint32(data[4:8]),
	}
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

func (r *TiffReader) bytesToSignedRational(data []byte) model.SignedRational {
	return model.SignedRational{
		Numerator:   int32(r.byteOrder.Uint32(data[0:4])),
		Denominator: int32(r.byteOrder.Uint32(data[4:8])),
	}
}

func (r *TiffReader) bytesToFloat32(data []byte) float32 {
	return float32(r.byteOrder.Uint64(data))
}

func (r *TiffReader) bytesToFloat64(data []byte) float64 {
	return float64(r.byteOrder.Uint64(data))
}

func (r *TiffReader) GetTile(img model.TIFF, level, tile int) ([]byte, error) {
	var tileOffset uint64
	var tileBytesCount uint64

	//if r.isBigTiff {
	//	tileOffset = img.Level(level).Tag(tags.TileOffsets).AsUint64s()[tile]
	//	tileBytesCount = img.Level(level).Tag(tags.TileByteCounts).AsUint64s()[tile]
	//} else {
	//	tileOffset = uint64(img.Level(level).Tag(tags.TileOffsets).AsUint32s()[tile])
	//	tileBytesCount = uint64(img.Level(level).Tag(tags.TileByteCounts).AsUint32s()[tile])
	//}

	tileOffset = img.Level(level).Tag(tags.TileOffsets).GetUintVal(tile)
	tileBytesCount = img.Level(level).Tag(tags.TileByteCounts).GetUintVal(tile)

	data, err := r.readBytesAt(tileOffset, tileBytesCount)
	if err != nil {
		return nil, fmt.Errorf("GetTile: cannot read tile at level %d, tile %d: %w", level, tile, err)
	}

	return data, nil
}
