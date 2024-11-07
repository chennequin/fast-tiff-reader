package model

import (
	"TiffReader/internal/tiffio/model"
	"fmt"
)

type SlideMetadata struct {
	Directories model.TIFFMetadata // contains the pyramid Metadata
	ExtraImages model.TIFFMetadata // contains a few optional extra images
}

func (t SlideMetadata) Level(level int) (model.TIFFDirectory, error) {
	if level >= len(t.Directories) {
		return model.TIFFDirectory{}, fmt.Errorf("level out of range: %d", level)
	}
	return t.Directories[level], nil
}

func (t SlideMetadata) Extra(idx int) (model.TIFFDirectory, error) {
	if idx >= len(t.ExtraImages) {
		return model.TIFFDirectory{}, fmt.Errorf("directory index out of range: %d", idx)
	}
	return t.ExtraImages[idx], nil
}

type PyramidImageMetadata struct {
	Levels []PyramidImage
}

type PyramidImage struct {
	ImageWidth          int
	ImageHeight         int
	TileWidth           int
	TileHeight          int
	TileCountHorizontal int
	TileCountVertical   int
}

func (i PyramidImage) TileIndex(tileX, tileY int) int {
	numTilesHorizontal := i.ImageWidth / i.TileWidth
	if i.ImageWidth%i.TileWidth > 0 {
		numTilesHorizontal += 1
	}
	idx := tileY*numTilesHorizontal + tileX
	return idx
}
