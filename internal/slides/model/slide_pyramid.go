package model

import (
	"TiffReader/internal/tiffio/model"
)

type Pyramid struct {
	metadata model.TIFFMetadata
}

type PyramidImage struct {
	Levels []PyramidImageLevel
}

type PyramidImageLevel struct {
	ImageWidth          int
	ImageHeight         int
	TileWidth           int
	TileHeight          int
	TileCountHorizontal int
	TileCountVertical   int
}
