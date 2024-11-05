package slides

import (
	"TiffReader/internal/jpegio"
	"TiffReader/internal/tiffio"
	"TiffReader/internal/tiffio/model"
	"TiffReader/internal/tiffio/tags"
	"bytes"
	"fmt"
	"image"
	"image/jpeg"
)

type SlideReader struct {
	metaData model.TIFFMetadata
	reader   *tiffio.TiffReader
}

func NewSlideReader() *SlideReader {
	return &SlideReader{}
}

func (r *SlideReader) OpenFile(name string) error {
	binaryReader := tiffio.NewFileBinaryReader()
	cacheBinaryReader := tiffio.NewCacheBinaryReader(binaryReader)
	tiffReader := tiffio.NewTiffReader(cacheBinaryReader)

	err := tiffReader.Open(name)
	if err != nil {
		return err
	}

	metaData, err := tiffReader.ReadMetaData()
	if err != nil {
		return fmt.Errorf("unable to read MetaData: %w", err)
	}

	r.reader = tiffReader
	r.metaData = metaData
	return nil
}

func (r *SlideReader) Close() {
	r.reader.Close()
	r.metaData = model.TIFFMetadata{}
}

func (r *SlideReader) LevelCount() int {
	return r.metaData.LevelCount()
}

func (r *SlideReader) GetTile(levelIdx, tileIdx int) ([]byte, error) {
	level, err := r.metaData.Level(levelIdx)
	if err != nil {
		return nil, err
	}

	jpegTables, err := level.Tag(tags.JPEGTables)
	if err != nil {
		return nil, fmt.Errorf("missing required tags: %w", err)
	}

	tileData, err := r.reader.GetTileData(r.metaData, levelIdx, tileIdx)
	if err != nil {
		return nil, fmt.Errorf("unable to obtain tile data: %w", err)
	}

	tileWidth, tileHeight, encoded, err := jpegio.MergeSegments(tileData, jpegTables.AsBytes())
	if err != nil {
		return nil, fmt.Errorf("unable to merge JPEG segments: %w", err)
	}

	expectedWidth, expectedHeight, err := r.calculateTileWidthHeight(levelIdx, tileIdx)
	if err != nil {
		return nil, fmt.Errorf("unable to calculate expected tile size: %w", err)
	}

	if tileWidth != expectedWidth || tileHeight != expectedHeight {
		encoded, err = r.cropImage(expectedWidth, expectedHeight, encoded)
	}

	return encoded, err
}

func (r *SlideReader) GetStrip(levelIdx, stripIdx int) ([]byte, error) {
	level, err := r.metaData.Level(levelIdx)
	if err != nil {
		return nil, err
	}

	jpegTables, err := level.Tag(tags.JPEGTables)
	if err != nil {
		return nil, fmt.Errorf("missing required tags: %w", err)
	}

	stripData, err := r.reader.GetStripData(r.metaData, levelIdx, stripIdx)
	if err != nil {
		return nil, fmt.Errorf("unable to obtain strip data: %w", err)
	}

	tileWidth, tileHeight, encoded, err := jpegio.MergeSegments(stripData, jpegTables.AsBytes())
	if err != nil {
		return nil, fmt.Errorf("unable to merge JPEG segments: %w", err)
	}

	expectedWidth, expectedHeight, err := r.calculateStripWidthHeight(levelIdx, stripIdx)
	if err != nil {
		return nil, fmt.Errorf("unable to calculate expected strip size: %w", err)
	}

	if tileWidth != expectedWidth || tileHeight != expectedHeight {
		encoded, err = r.cropImage(expectedWidth, expectedHeight, encoded)
	}

	return encoded, err
}

func (r *SlideReader) cropImage(expectedWidth, expectedHeight int, tileData []byte) ([]byte, error) {

	img, err := jpeg.Decode(bytes.NewReader(tileData))
	if err != nil {
		return nil, fmt.Errorf("unable to decode JPEG: %w", err)
	}

	cropRect := image.Rect(0, 0, expectedWidth, expectedHeight)
	croppedImg, ok := img.(interface {
		SubImage(r image.Rectangle) image.Image
	})
	if !ok {
		return nil, fmt.Errorf("image does not support cropping")
	}

	cropped := croppedImg.SubImage(cropRect)

	buf := bytes.NewBuffer(make([]byte, 0, len(tileData)))
	if err = jpeg.Encode(buf, cropped, nil); err != nil {
		return nil, fmt.Errorf("unable to encode JPEG: %w", err)
	}

	return buf.Bytes(), nil
}

func (r *SlideReader) calculateStripWidthHeight(levelIdx, tileIdx int) (int, int, error) {
	level, err := r.metaData.Level(levelIdx)
	if err != nil {
		return -1, -1, err
	}

	imageTags, err := level.Tags(tags.ImageWidth, tags.ImageLength)
	if err != nil {
		return -1, -1, fmt.Errorf("missing required tags: %w", err)
	}

	imageWidth := int(imageTags[0].GetUintVal(0))
	imageLength := int(imageTags[1].GetUintVal(0))

	return imageWidth, imageLength, nil
}

func (r *SlideReader) calculateTileWidthHeight(levelIdx, tileIdx int) (int, int, error) {
	level, err := r.metaData.Level(levelIdx)
	if err != nil {
		return -1, -1, err
	}

	imageTags, err := level.Tags(tags.ImageWidth, tags.ImageLength, tags.TileWidth, tags.TileLength)
	if err != nil {
		return -1, -1, fmt.Errorf("missing required tags: %w", err)
	}

	imageWidth := int(imageTags[0].GetUintVal(0))
	imageLength := int(imageTags[1].GetUintVal(0))
	tileWidth := int(imageTags[2].GetUintVal(0))
	tileLength := int(imageTags[3].GetUintVal(0))

	actualWidth := actualTileWidth(imageWidth, tileWidth, tileIdx)
	actualHeight := actualTileHeight(imageLength, tileLength, tileIdx)

	return actualWidth, actualHeight, nil
}

func actualTileWidth(imageWidth, tileWidth, tileIdx int) int {
	if lastTileWidth := imageWidth % tileWidth; lastTileWidth > 0 {
		numTilesHorizontal := imageWidth/tileWidth + 1
		tilePosX := tileIdx % numTilesHorizontal
		if tilePosX >= numTilesHorizontal-1 {
			return lastTileWidth
		}
		return tileWidth
	}
	return tileWidth
}

func actualTileHeight(imageHeight, tileHeight, tileIdx int) int {
	if lastTileHeight := imageHeight % tileHeight; lastTileHeight > 0 {
		numTilesVertical := imageHeight/tileHeight + 1
		tilePosY := tileIdx / numTilesVertical
		if tilePosY >= numTilesVertical-1 {
			return lastTileHeight
		}
		return tileHeight
	}
	return tileHeight
}
