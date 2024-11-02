package main

import (
	"TiffReader/internal/jpegio"
	"TiffReader/internal/tiffio"
	"TiffReader/internal/tiffio/model"
	"TiffReader/internal/tiffio/tags"
	"bytes"
	"encoding/binary"
	"fmt"
	"image"
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
	tiffReader := tiffio.NewTiffReader(binaryReader)

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

	level := 7
	tileNum := 1
	output := fmt.Sprintf("tile_%d_%d.jpeg", level, tileNum)

	tile, err := tiffReader.GetTile(tiffImg, level, tileNum)
	if err != nil {
		log.Fatalf("unable to read image: %s", err)
	}

	imageWidth := int(tiffImg.Level(level).Tag(tags.ImageWidth).AsUint16s()[0])
	imageLength := int(tiffImg.Level(level).Tag(tags.ImageLength).AsUint16s()[0])
	tileWidth := int(tiffImg.Level(level).Tag(tags.TileWidth).AsUint16s()[0])
	tileLength := int(tiffImg.Level(level).Tag(tags.TileLength).AsUint16s()[0])

	numTilesHorizontal := imageWidth / tileWidth
	numTilesVertical := imageLength / tileLength
	lastTileWidth := imageWidth % tileWidth
	lastTileLength := imageLength % tileLength

	if lastTileWidth > 0 {
		numTilesHorizontal += 1
	}
	if lastTileLength > 0 {
		numTilesVertical += 1
	}

	tilePosX := tileNum % numTilesHorizontal
	tilePosY := tileNum / numTilesVertical

	width := tileWidth
	length := tileLength

	_ = numTilesHorizontal
	_ = numTilesVertical
	_ = lastTileWidth
	_ = lastTileLength
	_ = tilePosX
	_ = tilePosY
	_ = width
	_ = length

	if tilePosX == numTilesHorizontal-1 {
		width = lastTileWidth
	}

	if tilePosY == numTilesVertical-1 {
		length = lastTileLength
	}

	jpegTables := tiffImg.Level(level).Tag(tags.JPEGTables)
	encoded, err := jpegio.MergeSegments(tile, jpegTables.AsBytes())

	err = os.WriteFile(output, encoded, os.ModePerm)
	if err != nil {
		fmt.Println("Error writing file:", err)
		return
	}

	img, err := jpeg.Decode(bytes.NewReader(encoded))
	if err != nil {
		fmt.Println("Erreur de décompression de l'image:", err)
		return
	}

	cropRect := image.Rect(0, 0, width, length)
	croppedImg := img.(interface {
		SubImage(r image.Rectangle) image.Image
	}).SubImage(cropRect)

	outFile, err := os.Create(fmt.Sprintf("clone_%s", output))
	if err != nil {
		fmt.Println("Erreur de création du fichier:", err)
		return
	}

	if err := jpeg.Encode(outFile, croppedImg, nil); err != nil {
		fmt.Println("Erreur d'encodage de l'image:", err)
	}

	println("OK")
}
