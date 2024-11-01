package main

import (
	"TiffReader/internal/tiffio"
	"TiffReader/internal/tiffio/model"
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"os"
	"slices"
	"strings"
)

var ExifLittleEndianSignature = [2]byte{0x49, 0x49}
var ExifBigEndianSignature = [2]byte{0x4d, 0x4d}
var TiffVersion = [2]byte{0x2a, 0x00}

// var globalOffset = 0
var byteOrder binary.ByteOrder

func main() {
	slog.SetLogLoggerLevel(slog.LevelDebug)

	name := "assets/CMU-1.tiff"

	binaryReader := tiffio.NewFileBinaryReader()
	tiffReader := tiffio.NewReader(binaryReader)
	err := tiffReader.Open(name)
	if err != nil {
		log.Fatalf("unable to open %s", name)
	}
	defer func(t *tiffio.Reader) {
		err := t.Close()
		if err != nil {
			log.Fatalf("unable to close %s", name)
		}
	}(tiffReader)

	nextIFD, err := tiffReader.ReadHeader()
	if err != nil {
		log.Fatalf("unable to read header: %s", err)
	}

	imageFileDirectories := make([]model.IFD, 0)
	for nextIFD != 0 {
		ifd, err := tiffReader.ReadIFD(nextIFD)
		if err != nil {
			log.Fatalf("unable to read IDF: %s", err)
		}
		fmt.Printf("%s\n", ifd)
		imageFileDirectories = append(imageFileDirectories, ifd)
		nextIFD = ifd.NextIFD
	}

	file, err := os.Open(name)
	if err != nil {
		log.Fatalf("unable to open file %s", name)
	}
	defer file.Close()

	offset := int64(0)
	_, err = file.Seek(offset, 0)
	if err != nil {
		log.Fatalf("unable to seek at %d", offset)
	}

	// read TIFF header
	buffer := make([]byte, 20)
	n, err := file.Read(buffer)
	if err != nil {
		log.Fatalf("unable to read file %s", file.Name())
	}
	slog.Debug("read", "bytes", hex.EncodeToString(buffer[:n]))

	if bytes.Equal(buffer[:2], ExifBigEndianSignature[:]) {
		slog.Debug("Byte-order is big-endian.")
		byteOrder = binary.BigEndian
	} else if bytes.Equal(buffer[:2], ExifLittleEndianSignature[:]) {
		slog.Debug("Byte-order is little-endian.")
		byteOrder = binary.LittleEndian
	} else {
		log.Fatalf("Unknown TIFF header: %s", hex.EncodeToString(buffer[:4]))
	}

	if !bytes.Equal(buffer[2:4], TiffVersion[:]) {
		log.Fatalf("TIFF version does not match: %s", hex.EncodeToString(buffer[:4]))
	}

	//nextIFD := int64(byteOrder.Uint32(buffer[4:8]))
	for nextIFD != 0 {
		nextIFD = readIFD(file, nextIFD)
	}
}

func readHeader(file *os.File, offset int64) {

}

func readIFD(file *os.File, offset int64) int64 {
	_, err := file.Seek(offset, 0)
	if err != nil {
		log.Fatalf("unable to seek at %d", offset)
	}

	buffer := make([]byte, 20)
	n, err := file.Read(buffer)
	if err != nil {
		log.Fatalf("unable to read file %s", file.Name())
	}

	slog.Debug("readIFD: read", "bytes", hex.EncodeToString(buffer[:n]))

	// read number of tags
	nbTags := byteOrder.Uint16(buffer[:2])
	slog.Info("readIFD: found", "nbTags", nbTags)
	offset += 2

	// read tags
	for range nbTags {
		rOffset := readTag(file, offset)
		offset += int64(rOffset)
	}

	_, err = file.Seek(offset, 0)
	if err != nil {
		log.Fatalf("unable to seek at %d", offset)
	}

	buffer = make([]byte, 4)
	_, err = file.Read(buffer)
	if err != nil {
		log.Fatalf("unable to read file %s", file.Name())
	}

	// offset to next IDF
	nextIFD := int64(byteOrder.Uint32(buffer[:4]))
	slog.Debug("readIFD: found", "nextIFD", nextIFD)

	return nextIFD
}

func readTag(file *os.File, offset int64) int {
	slog.Debug("readTag", "seek at", offset)
	_, err := file.Seek(offset, 0)
	if err != nil {
		log.Fatalf("unable to seek at %d", offset)
	}

	buffer := make([]byte, 12)
	n, err := file.Read(buffer)
	if err != nil {
		log.Fatalf("unable to read file %s", file.Name())
	}
	slog.Debug("readTag", "bytes", hex.EncodeToString(buffer[:n]))

	idTag := byteOrder.Uint16(buffer[:2])
	tagType := byteOrder.Uint16(buffer[2:4])
	numVals := byteOrder.Uint32(buffer[4:8])
	slog.Info("readTag: found", "ID", idTag, "Label", model.TagsIDsLabels[idTag], "Type", tagType, "NumVals", numVals)

	vOffset := 8
	switch tagType {

	// byte
	case 0x1:
		if numVals <= 4 {
			for range numVals {
				val := buffer[vOffset : vOffset+1]
				slog.Info("readTag: found", "byte", val)
			}
		}
		if numVals > 4 {
			rOffset := int64(byteOrder.Uint32(buffer[vOffset : vOffset+4]))
			readValuesAt(file, rOffset, tagType, numVals)
		}
		vOffset += 4

	// ASCII (nul terminated \0) string
	case 0x2:
		rOffset := int64(byteOrder.Uint32(buffer[vOffset : vOffset+4]))
		readValuesAt(file, rOffset, tagType, numVals)
		vOffset += 4

	// short
	case 0x3:
		if numVals == 1 {
			val := byteOrder.Uint16(buffer[vOffset : vOffset+2])
			slog.Info("readTag: found", "short", val)
		}
		if numVals == 2 {
			val := byteOrder.Uint16(buffer[vOffset : vOffset+2])
			slog.Info("readTag: found", "short", val)
			val = byteOrder.Uint16(buffer[vOffset+2 : vOffset+4])
			slog.Info("readTag: found", "short", val)
		}
		if numVals > 2 {
			rOffset := int64(byteOrder.Uint32(buffer[vOffset : vOffset+4]))
			readValuesAt(file, rOffset, tagType, numVals)
		}
		vOffset += 4

	// long
	case 0x4:
		if numVals == 1 {
			val := byteOrder.Uint32(buffer[vOffset : vOffset+4])
			slog.Info("readTag: found", "long", val)
		}
		if numVals > 1 {
			rOffset := int64(byteOrder.Uint32(buffer[vOffset : vOffset+4]))
			readValuesAt(file, rOffset, tagType, numVals)
		}
		vOffset += 4

	// rational
	case 0x5:
		rOffset := int64(byteOrder.Uint32(buffer[vOffset : vOffset+4]))
		readValuesAt(file, rOffset, tagType, numVals)
		vOffset += 4

	// signed byte
	case 6:
		if numVals <= 4 {
			for range numVals {
				val := int8(buffer[vOffset : vOffset+1][0])
				slog.Info("readTag: found", "signed byte", val)
			}
		}
		if numVals > 4 {
			rOffset := int64(byteOrder.Uint32(buffer[vOffset : vOffset+4]))
			readValuesAt(file, rOffset, tagType, numVals)
		}
		vOffset += 4

	// undefined
	case 0x7:
		if numVals <= 4 {
			slog.Info("readTag: found", "undefined", slices.Clone(buffer[vOffset:vOffset+int(numVals)]))
		}
		if numVals > 4 {
			rOffset := int64(byteOrder.Uint32(buffer[vOffset : vOffset+4]))
			readValuesAt(file, rOffset, tagType, numVals)
		}
		vOffset += 4

	// signed short
	case 0x8:
		if numVals <= 2 {
			for range numVals {
				val := int16(byteOrder.Uint16(buffer[vOffset : vOffset+2]))
				slog.Info("readTag: found", "signed short", val)
			}
		} else {
			rOffset := int64(byteOrder.Uint32(buffer[vOffset : vOffset+4]))
			readValuesAt(file, rOffset, tagType, numVals)
		}
		vOffset += 4

	// signed long
	case 0x9:
		if numVals <= 1 {
			for range numVals {
				val := int32(byteOrder.Uint32(buffer[vOffset : vOffset+4]))
				slog.Info("readTag: found", "signed long", val)
			}
		} else {
			rOffset := int64(byteOrder.Uint32(buffer[vOffset : vOffset+4]))
			readValuesAt(file, rOffset, tagType, numVals)
		}
		vOffset += 4

	// signed rational
	case 0xa:
		rOffset := int64(byteOrder.Uint32(buffer[vOffset : vOffset+4]))
		readValuesAt(file, rOffset, tagType, numVals)
		vOffset += 4

	// float
	case 0xb:
		if numVals <= 1 {
			val := float32(byteOrder.Uint32(buffer[vOffset : vOffset+4]))
			slog.Info("readTag: found", "float", val)
		} else {
			rOffset := int64(byteOrder.Uint32(buffer[vOffset : vOffset+4]))
			readValuesAt(file, rOffset, tagType, numVals)
		}
		vOffset += 4

	// double
	case 0xc:
		rOffset := int64(byteOrder.Uint32(buffer[vOffset : vOffset+4]))
		readValuesAt(file, rOffset, tagType, numVals)
		vOffset += 4

	default:
		log.Fatalf("Unknown tag type: %d", tagType)
	}

	slog.Debug("readTag: return", "jump", vOffset)
	return vOffset
}

func readValuesAt(file *os.File, offset int64, tagType uint16, numVals uint32) int {
	slog.Debug("readValuesAt: reading values", "offset", offset, "tagType", tagType, "numVals", numVals)

	_, seekError := file.Seek(offset, 0)
	if seekError != nil {
		log.Fatalf("unable to seek at %d", offset)
	}

	switch tagType {

	// byte
	case 0x1:
		buffer := make([]byte, numVals*1)
		bytesRead, err := file.Read(buffer)
		if bytesRead != int(numVals*1) {
			log.Fatalf("unable to read buffer from %s", file.Name())
		}
		if err != nil {
			log.Fatalf("unable to read file %s", file.Name())
		}

		vOffset := 0
		for range numVals {
			val := buffer[vOffset : vOffset+1]
			slog.Info("readValuesAt: found", "byte", val)
			vOffset += 1
		}
		return vOffset

	// ASCII (nul terminated \0) string
	case 0x2:
		vOffset := 0
		for range numVals {
			_, seekError := file.Seek(offset, 0)
			if seekError != nil {
				log.Fatalf("unable to seek at %d", offset)
			}

			var sb strings.Builder
			for {
				buffer := make([]byte, 256)
				bytesRead, readError := file.Read(buffer)
				if readError != nil {
					log.Fatalf("unable to read file %s", file.Name())
				}

				str, bytesProcessed, readError := readCString(buffer[:bytesRead])

				if readError != nil && errors.Is(readError, tiffio.ZeroNotFound) {
					vOffset += bytesRead
					sb.WriteString(str)
					continue
				}

				vOffset += bytesProcessed
				sb.WriteString(str)
				break
			}

			slog.Info("readValuesAt: found", "string", sb.String())
			return vOffset
		}

	// short
	case 0x3:
		buffer := make([]byte, numVals*2)
		bytesRead, err := file.Read(buffer)
		if bytesRead != int(numVals*2) {
			log.Fatalf("unable to read buffer from %s", file.Name())
		}
		if err != nil {
			log.Fatalf("unable to read file %s", file.Name())
		}

		vOffset := 0
		for range numVals {
			val := byteOrder.Uint16(buffer[vOffset : vOffset+2])
			slog.Info("readValuesAt: found", "short", val)
			vOffset += 2
		}
		return vOffset

	// long
	case 0x4:
		buffer := make([]byte, numVals*4)
		bytesRead, err := file.Read(buffer)
		if bytesRead != int(numVals*4) {
			log.Fatalf("unable to read buffer from %s", file.Name())
		}
		if err != nil {
			log.Fatalf("unable to read file %s", file.Name())
		}

		vOffset := 0
		for range numVals {
			val := byteOrder.Uint32(buffer[vOffset : vOffset+4])
			if numVals < 10 {
				slog.Info("readValuesAt: found", "long", val)
			}
			vOffset += 4
		}
		return vOffset

	// rational
	case 0x5:
		buffer := make([]byte, numVals*8)
		bytesRead, err := file.Read(buffer)
		if bytesRead != int(numVals*8) {
			log.Fatalf("unable to read buffer from %s", file.Name())
		}
		if err != nil {
			log.Fatalf("unable to read file %s", file.Name())
		}

		vOffset := 0
		for range numVals {
			numerator := byteOrder.Uint32(buffer[vOffset : vOffset+4])
			denominator := byteOrder.Uint32(buffer[vOffset+4 : vOffset+8])
			slog.Info("readValuesAt: found", "numerator", numerator, "denominator", denominator)
			vOffset += 8
		}
		return vOffset

	// signed byte
	case 0x6:
		buffer := make([]byte, numVals*1)
		bytesRead, err := file.Read(buffer)
		if bytesRead != int(numVals*1) {
			log.Fatalf("unable to read buffer from %s", file.Name())
		}
		if err != nil {
			log.Fatalf("unable to read file %s", file.Name())
		}

		vOffset := 0
		for range numVals {
			val := int8(buffer[vOffset : vOffset+1][0])
			slog.Info("readValuesAt: found", "byte", val)
			vOffset += 1
		}
		return vOffset

	// undefined
	case 0x7:
		buffer := make([]byte, numVals*1)
		bytesRead, err := file.Read(buffer)
		if bytesRead != int(numVals*1) {
			log.Fatalf("unable to read buffer from %s", file.Name())
		}
		if err != nil {
			log.Fatalf("unable to read file %s", file.Name())
		}

		slog.Info("readValuesAt: found", "undefined", slices.Clone(buffer))
		return bytesRead

	// signed short
	case 0x8:
		buffer := make([]byte, numVals*2)
		bytesRead, err := file.Read(buffer)
		if bytesRead != int(numVals*2) {
			log.Fatalf("unable to read buffer from %s", file.Name())
		}
		if err != nil {
			log.Fatalf("unable to read file %s", file.Name())
		}

		vOffset := 0
		for range numVals {
			val := int16(byteOrder.Uint16(buffer[vOffset : vOffset+2]))
			slog.Info("readValuesAt: found", "signed short", val)
			vOffset += 2
		}
		return vOffset

	// signed long
	case 0x9:
		buffer := make([]byte, numVals*4)
		bytesRead, err := file.Read(buffer)
		if bytesRead != int(numVals*4) {
			log.Fatalf("unable to read buffer from %s", file.Name())
		}
		if err != nil {
			log.Fatalf("unable to read file %s", file.Name())
		}

		vOffset := 0
		for range numVals {
			val := int32(byteOrder.Uint32(buffer[vOffset : vOffset+2]))
			slog.Info("readValuesAt: found", "signed long", val)
			vOffset += 4
		}
		return vOffset

	// signed rational
	case 0xa:
		buffer := make([]byte, numVals*8)
		bytesRead, err := file.Read(buffer)
		if bytesRead != int(numVals*8) {
			log.Fatalf("unable to read buffer from %s", file.Name())
		}
		if err != nil {
			log.Fatalf("unable to read file %s", file.Name())
		}

		vOffset := 0
		for range numVals {
			numerator := int32(byteOrder.Uint32(buffer[vOffset : vOffset+4]))
			denominator := int32(byteOrder.Uint32(buffer[vOffset+4 : vOffset+8]))
			slog.Info("readValuesAt: found", "numerator ", numerator, "denominator ", denominator)
			vOffset += 8
		}
		return vOffset

	// float
	case 0xb:
		buffer := make([]byte, numVals*4)
		bytesRead, err := file.Read(buffer)
		if bytesRead != int(numVals*4) {
			log.Fatalf("unable to read buffer from %s", file.Name())
		}
		if err != nil {
			log.Fatalf("unable to read file %s", file.Name())
		}

		vOffset := 0
		for range numVals {
			val := float32(byteOrder.Uint32(buffer[vOffset : vOffset+4]))
			slog.Info("readTag: found", "float", val)
			vOffset += 4
		}
		return vOffset

	// double
	case 0xc:
		buffer := make([]byte, numVals*8)
		bytesRead, err := file.Read(buffer)
		if bytesRead != int(numVals*8) {
			log.Fatalf("unable to read buffer from %s", file.Name())
		}
		if err != nil {
			log.Fatalf("unable to read file %s", file.Name())
		}

		vOffset := 0
		for range numVals {
			val := float64(byteOrder.Uint64(buffer[vOffset : vOffset+8]))
			slog.Info("readTag: found", "double", val)
			vOffset += 8
		}
		return vOffset

	default:
		log.Fatalf("Unknown tag type: %d", tagType)
	}

	return 0
}

func readCString(buf []byte) (string, int, error) {
	n := bytes.IndexByte(buf, 0)
	if n == -1 {
		return string(buf), len(buf), tiffio.ZeroNotFound
	}
	return string(buf[:n]), n, nil
}
