package slides

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
