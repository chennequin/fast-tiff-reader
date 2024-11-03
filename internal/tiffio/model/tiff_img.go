package model

import (
	"TiffReader/internal/tiffio/tags"
	"fmt"
)

// --------------------------
// TIFF (Tagged Image File Format)
// --------------------------

type TIFF struct {
	IFDs []IFD
}

func (t TIFF) Level(level int) IFD {
	return t.IFDs[level]
}

func (t TIFF) String() string {
	return fmt.Sprintf("%v", t.IFDs)
}

// --------------------------
// IFD (Image File Directory)
// --------------------------

type IFD struct {
	Tags    map[tags.TagID]Tag
	NextIFD uint64
}

func (d IFD) Tag(tagID tags.TagID) Tag {
	return d.Tags[tagID]
}

func (d IFD) String() string {
	return fmt.Sprintf("%v", d.Tags)
}
