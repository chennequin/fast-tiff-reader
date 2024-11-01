package main

import (
	"TiffReader/internal/jpegio"
	"TiffReader/internal/tiffio"
	"TiffReader/internal/tiffio/model"
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"image/jpeg"
	"log"
	"log/slog"
	"os"
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
	defer tiffReader.Close()

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

	tiffImg := model.TIFF{
		IFDs: imageFileDirectories,
	}

	//lastIFD := tiffImg.IFDs[len(tiffImg.IFDs)-1]

	output := "tile_7_0.jpeg"

	tile, err := tiffReader.GetTile(tiffImg, 7, 0)
	if err != nil {
		log.Fatalf("unable to read image: %s", err)
	}

	QuantizationTable := tile[2 : 4+tile[5]]
	Left := tile[4+tile[5] : 4+4+tile[5]+tile[5]]

	JPEGTables := tiffReader.GetJPEGTables(tiffImg, 7)

	slog.Info("read", "bytes", hex.EncodeToString(QuantizationTable))
	slog.Info("read", "bytes", hex.EncodeToString(Left))
	slog.Info("read", "bytes", hex.EncodeToString(JPEGTables))

	jpegReader := jpegio.NewReader()
	jTables, err := jpegReader.Parse(JPEGTables)
	if err != nil {
		log.Fatalf("unable to parse JPEG: %s", err)
	}
	jTile, err := jpegReader.Parse(tile)
	if err != nil {
		log.Fatalf("unable to parse JPEG: %s", err)
	}

	j, err := jpegio.MergeJPEGTables(jTile, jTables)
	if err != nil {
		log.Fatalf("unable to merge JPEG: %s", err)
	}

	data := jpegio.EncodeJPEG(j)

	err = os.WriteFile(output, data, os.ModePerm)
	if err != nil {
		fmt.Println("Error writing file:", err)
		return
	}

	img, err := jpeg.Decode(bytes.NewReader(data))
	if err != nil {
		fmt.Println("Erreur de décompression de l'image:", err)
		return
	}

	outFile, err := os.Create(fmt.Sprintf("clone_%s", output))
	if err != nil {
		fmt.Println("Erreur de création du fichier:", err)
		return
	}

	if err := jpeg.Encode(outFile, img, nil); err != nil {
		fmt.Println("Erreur d'encodage de l'image:", err)
	}

	println("OK")
}
