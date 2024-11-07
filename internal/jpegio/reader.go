package jpegio

import (
	"TiffReader/internal/jpegio/model"
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"
)

// Decode the structure of a JPEG image.
// SOI (Start of Image): 0xFFD8
// Application Segments (APPn): 0xFFE0 to 0xFFEF
// DRI (Define Restart Interval - 0xFFDD)
// DQT (Define Quantization Table): 0xFFDB
// DHT (Define Huffman Table): 0xFFC4 (often located just after the DQT segment).
// SOF (Start of Frame): 0xFFC0, 0xFFC1, etc. (describes the image dimensions and components).
// SOS (Start of Scan): 0xFFDA (indicates the start of image data).
// Image Data
// EOI (End of Image): 0xFFD9

var SOI = []byte{0xFF, 0xD8}
var APPn = []byte{0xFF, 0xE0}
var DRI = []byte{0xFF, 0xDD}
var DQT = []byte{0xFF, 0xDB}
var DHT = []byte{0xFF, 0xC4}
var SOS = []byte{0xFF, 0xDA}
var EOI = []byte{0xFF, 0xD9}
var CMT = []byte{0xFF, 0xFE}

// MergeSegments combines two JPEG images by appending the Huffman and Quantization tables from the second image to the first.
func MergeSegments(img, imgJpegTables, iccProfile []byte) (int, int, []byte, error) {
	jpegTile, err := parseJPEG(img)
	if err != nil {
		return 0, 0, nil, fmt.Errorf("unable to parse JPEG: %w", err)
	}
	jpegTables, err := parseJPEG(imgJpegTables)
	if err != nil {
		return 0, 0, nil, fmt.Errorf("unable to parse JPEG: %w", err)
	}

	merged, err := mergeJPEG(jpegTile, jpegTables)
	if err != nil {
		return 0, 0, nil, fmt.Errorf("unable to merge JPEG: %w", err)
	}

	encoded := encodeJPEG(merged, iccProfile)
	width, height, err := decodeSOF(merged.SOF)

	return width, height, encoded, err
}

func DecodeSOF(img []byte) (int, int, error) {
	jpegTile, err := parseJPEG(img)
	if err != nil {
		return 0, 0, fmt.Errorf("unable to parse JPEG: %w", err)
	}
	return decodeSOF(jpegTile.SOF)
}

func mergeJPEG(img1, img2 model.Jpeg) (model.Jpeg, error) {
	if img1.DQT == nil && img2.DQT != nil {
		img1.DQT = img2.DQT
	}
	if img1.DHT == nil && img2.DHT != nil {
		img1.DHT = img2.DHT
	}
	return img1, nil
}

func encodeJPEG(j model.Jpeg, iccProfile []byte) []byte {
	buffer := make([]byte, 0, j.TotalSize())
	buffer = append(buffer, SOI...)
	for _, app := range j.APPn {
		buffer = append(buffer, app...)
	}
	if len(iccProfile) < 65535-4-4 {
		buffer = append(buffer, createICCSegment(iccProfile)...)
	}
	for _, dqt := range j.DQT {
		buffer = append(buffer, dqt...)
	}
	for _, dht := range j.DHT {
		buffer = append(buffer, dht...)
	}
	buffer = append(buffer, j.SOF...)
	buffer = append(buffer, j.SOS...)
	buffer = append(buffer, EOI...)
	return buffer
}

func parseJPEG(data []byte) (model.Jpeg, error) {
	var img model.Jpeg

	if len(data) == 0 {
		return model.Jpeg{}, nil
	}

	if !isSOI(data) {
		return img, fmt.Errorf("invalid JPEG format: missing SOI marker")
	}

	offset := 2

	for offset < len(data) {
		if isAPPn(data[offset:]) {
			size, block := jpegSegment(data[offset:])
			img.APPn = append(img.APPn, block)
			offset += size + 2
			continue
		}

		if isDRI(data[offset:]) {
			size, block := jpegSegment(data[offset:])
			img.DRI = append(img.DRI, block)
			offset += size + 2
			continue
		}

		if isDQT(data[offset:]) {
			size, block := jpegSegment(data[offset:])
			img.DQT = append(img.DQT, block)
			offset += size + 2
			continue
		}

		if isDHT(data[offset:]) {
			size, block := jpegSegment(data[offset:])
			img.DHT = append(img.DHT, block)
			offset += size + 2
			continue
		}

		if isSOF(data[offset:]) {
			size, block := jpegSegment(data[offset:])
			img.SOF = block
			offset += size + 2
			continue
		}

		if isSOS(data[offset:]) {
			img.SOS = data[offset : len(data)-2]
			offset = len(data) - 2
			continue
		}

		if isEOI(data[offset:]) {
			offset += 2
			continue
		}

		if isCMT(data[offset:]) {
			// ignore JPEG Comment Block (0xFFFE)
			size, _ := jpegSegment(data[offset:])
			offset += size + 2
			continue
		}

		return img, fmt.Errorf("invalid JPEG format: unknown block 0x%s", hex.EncodeToString(data[offset:offset+2]))
	}

	return img, nil
}

func isSOI(data []byte) bool {
	return len(data) >= 2 && bytes.Equal(data[0:2], SOI)
}

func isAPPn(data []byte) bool {
	return len(data) >= 2 && data[0] == APPn[0] && (data[1]&0xf0) == APPn[1]
}

func isDRI(data []byte) bool {
	return len(data) >= 2 && bytes.Equal(data[:2], DRI)
}

func isDQT(data []byte) bool {
	return len(data) >= 2 && bytes.Equal(data[:2], DQT)
}

func isDHT(data []byte) bool {
	return len(data) >= 2 && bytes.Equal(data[:2], DHT)
}

func isSOF(data []byte) bool {
	return len(data) >= 2 && data[0] == 0xFF && (data[1] == 0xC0 || data[1] == 0xC1 || data[1] == 0xC2 || data[1] == 0xC3)
}

func isSOS(data []byte) bool {
	return len(data) >= 2 && bytes.Equal(data[:2], SOS)
}

func isEOI(data []byte) bool {
	return len(data) >= 2 && bytes.Equal(data[:2], EOI)
}

func isCMT(data []byte) bool {
	return len(data) >= 2 && bytes.Equal(data[:2], CMT)
}

func jpegSegment(data []byte) (int, []byte) {
	size := int(binary.BigEndian.Uint16(data[2:4]))
	return size, data[:size+2]
}

func decodeSOF(sofSegment []byte) (width, height int, err error) {
	if len(sofSegment) < 9 {
		return 0, 0, fmt.Errorf("SOF segment too short")
	}

	height = int(binary.BigEndian.Uint16(sofSegment[5:7]))
	width = int(binary.BigEndian.Uint16(sofSegment[7:9]))

	return width, height, nil
}

func createICCSegment(iccProfile []byte) []byte {
	segment := []byte{0xFF, 0xE2}
	length := uint16(len(iccProfile) + 4 + 2)
	segment = append(segment, byte(length>>8), byte(length&0xFF))
	segment = append(segment, []byte("ICC_")...)
	segment = append(segment, iccProfile...)
	return segment
}
