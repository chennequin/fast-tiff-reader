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

var LittleEndianSignature = [2]byte{0x49, 0x49}
var BigEndianSignature = [2]byte{0x4d, 0x4d}
var TiffVersion = [2]byte{0x2a, 0x00}

type TiffReader struct {
	name      string
	byteOrder binary.ByteOrder
	binary    BinaryReader
}

func NewTiffReader(binary BinaryReader) *TiffReader {
	return &TiffReader{
		binary: binary,
	}
}

func (r *TiffReader) Open(name string) error {
	r.name = name
	return r.binary.open(name)
}

func (r *TiffReader) Close() {
	err := r.binary.close()
	if err != nil {
		log.Fatalf("unable to close %s", r.name)
	}
}

func (r *TiffReader) ReadHeader() (int64, error) {
	_, err := r.binary.seek(0)
	if err != nil {
		return -1, fmt.Errorf("cannot seek to header: %w", err)
	}

	buffer := make([]byte, 8)
	_, err = r.binary.read(buffer)
	if err != nil {
		return -1, fmt.Errorf("cannot read header: %w", err)
	}

	if bytes.Equal(buffer[:2], BigEndianSignature[:]) {
		r.byteOrder = binary.BigEndian
		slog.Debug("Byte-order is big-endian.")
	} else if bytes.Equal(buffer[:2], LittleEndianSignature[:]) {
		r.byteOrder = binary.LittleEndian
		slog.Debug("Byte-order is little-endian.")
	} else {
		return -1, errors.New(fmt.Sprintf("Unknown TIFF header: %s", hex.EncodeToString(buffer[:4])))
	}

	nextIFD := int64(r.byteOrder.Uint32(buffer[4:8]))
	return nextIFD, nil
}

func (r *TiffReader) ReadIFD(offset int64) (model.IFD, error) {
	// read number of tags
	buffer, err := r.read2BytesAt(offset)
	if err != nil {
		return model.IFD{}, fmt.Errorf("ReadIFD: cannot read: %w", err)
	}

	nbTags := r.byteOrder.Uint16(buffer[:2])
	offset += 2

	// read tags
	tags := make(map[tags.TagID]model.Tag, nbTags)
	for range nbTags {
		tag, err := r.ReadTag(offset)
		if err != nil {
			return model.IFD{}, fmt.Errorf("ReadIFD: cannot read TAG: %w", err)
		}
		tags[tag.GetTagID()] = tag
		offset += 12
	}

	// offset to next IDF
	buffer, err = r.read4BytesAt(offset)
	if err != nil {
		return model.IFD{}, fmt.Errorf("ReadIFD: cannot read: %w", err)
	}

	nextIFD := int64(r.byteOrder.Uint32(buffer[:4]))
	ifd := model.IFD{
		NextIFD: nextIFD,
		Tags:    tags,
	}
	return ifd, nil
}

//func edd() {
//	switch t := tag.(type) {
//	case model.Tag[uint16]:
//		// Traiter le Tag[uint16]
//	case model.Tag[uint32]:
//		// Traiter le Tag[uint32]
//	default:
//		// Type inconnu
//	}
//}

func (r *TiffReader) ReadTag(offset int64) (model.Tag, error) {
	buffer, err := r.readBytesAt(offset, 12)
	if err != nil {
		return nil, fmt.Errorf("ReadTag: cannot read: %w", err)
	}

	tagID := r.byteOrder.Uint16(buffer[:2])
	tagType := r.byteOrder.Uint16(buffer[2:4])
	numValues := r.byteOrder.Uint32(buffer[4:8])

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

		rOffset := int64(r.byteOrder.Uint32(buffer[0:4]))
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
		rOffset := int64(r.byteOrder.Uint32(buffer[0:4]))
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
			rOffset := int64(r.byteOrder.Uint32(buffer[0:4]))
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
			rOffset := int64(r.byteOrder.Uint32(buffer[0:4]))
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
		rOffset := int64(r.byteOrder.Uint32(buffer[0:4]))
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
			rOffset := int64(r.byteOrder.Uint32(buffer[0:4]))
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
			rOffset := int64(r.byteOrder.Uint32(buffer[0:4]))
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
			rOffset := int64(r.byteOrder.Uint32(buffer[0:4]))
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
		rOffset := int64(r.byteOrder.Uint32(buffer[0:4]))
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
			rOffset := int64(r.byteOrder.Uint32(buffer[0:4]))
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
		rOffset := int64(r.byteOrder.Uint32(buffer[0:4]))
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

func (r *TiffReader) readStringValuesAt(offset int64, numValues uint32) ([]string, error) {
	values := make([]string, numValues)

	for i := range numValues {
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
		values[i] = sb.String()
	}

	return values, nil
}

func readValuesAt[T model.TagType](r *TiffReader, offset int64, numValues, elementSize uint32, fromBytesFn func(data []byte) T) ([]T, error) {
	buffer, err := r.readBytesAt(offset, numValues*elementSize)
	if err != nil {
		return nil, fmt.Errorf("readValuesAt: cannot read: %w", err)
	}

	vOffset := uint32(0)
	values := make([]T, numValues)
	for i := range numValues {
		values[i] = fromBytesFn(buffer[vOffset : vOffset+elementSize])
		vOffset += elementSize
	}

	return values, nil
}

func (r *TiffReader) read2BytesAt(offset int64) ([]byte, error) {
	return r.readBytesAt(offset, 2)
}

func (r *TiffReader) read4BytesAt(offset int64) ([]byte, error) {
	return r.readBytesAt(offset, 4)
}

func (r *TiffReader) readBytesAt(offset int64, n uint32) ([]byte, error) {
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
	return int32(r.byteOrder.Uint32(data[0:2]))
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
	tileOffset := int64(img.Level(level).Tag(tags.TileOffsets).AsUint32s()[tile])
	tileBytesCount := img.Level(level).Tag(tags.TileByteCounts).AsUint32s()[tile]

	data, err := r.readBytesAt(tileOffset, tileBytesCount)
	if err != nil {
		return nil, fmt.Errorf("GetTile: cannot read tile at level %d, tile %d: %w", level, tile, err)
	}

	return data, nil
}
