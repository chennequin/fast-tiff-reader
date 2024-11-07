package model

import (
	"TiffReader/internal/tiffio/model"
)

type Pyramid struct {
	metadata model.TIFFMetadata
	pyramids map[string][]model.TIFFDirectory
}
