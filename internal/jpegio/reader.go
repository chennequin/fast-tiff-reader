package jpegio

import (
	"TiffReader/internal/jpegio/model"
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"log/slog"
)

/*
	Decode the structure of a JPEG image.
	  SOI (Start of Image) : FFD8
		Application Segments (APPn) : FFE0 à FFEF
		DQT (Define Quantization Table) : FFDB
		 DHT (Define Huffman Table) : FFC4 (souvent situé juste après le segment DQT).
		 SOF (Start of Frame) : FFC0, FFC1, etc. (décrit les dimensions de l'image et les composants).
		 SOS (Start of Scan) : FFDA (indique le début des données d'image).
		 Données d'Image
		 EOI (End of Image) : FFD9
*/

var SOI = []byte{0xFF, 0xD8}
var APPn = []byte{0xFF, 0xE0}
var DQT = []byte{0xFF, 0xDB}
var DHT = []byte{0xFF, 0xC4}
var SOS = []byte{0xFF, 0xDA}
var EOI = []byte{0xFF, 0xD9}

func MergeSegments(img1, img2 []byte) ([]byte, error) {
	jpegTile, err := parseJPEG(img1)
	if err != nil {
		return nil, fmt.Errorf("unable to parse JPEG: %w", err)
	}
	jpegTables, err := parseJPEG(img2)
	if err != nil {
		return nil, fmt.Errorf("unable to parse JPEG: %w", err)
	}
	merged, err := mergeJPEG(jpegTile, jpegTables)
	if err != nil {
		return nil, fmt.Errorf("unable to merge JPEG: %w", err)
	}
	encoded := encodeJPEG(merged)
	return encoded, nil
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

func encodeJPEG(j model.Jpeg) []byte {
	buffer := make([]byte, 0, j.TotalSize())
	buffer = append(buffer, SOI...)
	for _, app := range j.APPn {
		buffer = append(buffer, app...)
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

	if !isSOI(data) {
		return img, fmt.Errorf("invalid JPEG format: missing SOI marker")
	}

	offset := 2

	for offset < len(data) {
		slog.Debug("parse", "bytes", hex.EncodeToString(data[offset:]))

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

		return img, fmt.Errorf("invalid JPEG format: unknown block 0x%s", hex.EncodeToString(data[offset:offset+2]))
	}

	return img, nil
}

func isSOI(data []byte) bool {
	return bytes.Equal(data[0:2], SOI)
}

func isDQT(data []byte) bool {
	return bytes.Equal(data[:2], DQT)
}

func isDHT(data []byte) bool {
	return bytes.Equal(data[:2], DHT)
}

func isSOF(data []byte) bool {
	return data[0] == 0xFF && (data[1] == 0xC0 || data[1] == 0xC1 || data[1] == 0xC2 || data[1] == 0xC3)
}

func isSOS(data []byte) bool {
	return bytes.Equal(data[:2], SOS)
}

func isEOI(data []byte) bool {
	return bytes.Equal(data[:2], EOI)
}

func jpegSegment(data []byte) (int, []byte) {
	size := int(binary.BigEndian.Uint16(data[2:4]))
	return size, data[:size+2]
}
