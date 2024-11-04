package model

import (
	"TiffReader/internal/tiffio/tags"
	"fmt"
)

// --------------------------
// TIFF (Tagged Image File Format)
// --------------------------

type TIFFMetadata struct {
	entries []TIFFDirectory
}

func NewTIFFMetadata(entries []TIFFDirectory) TIFFMetadata {
	return TIFFMetadata{entries: entries}
}

func (t TIFFMetadata) Level(level int) (TIFFDirectory, error) {
	if level >= len(t.entries) {
		return TIFFDirectory{}, fmt.Errorf("level out of range: %d", level)
	}
	return t.entries[level], nil
}

func (t TIFFMetadata) LevelCount() int {
	return len(t.entries)
}

func (t TIFFMetadata) String() string {
	return fmt.Sprintf("%v", t.entries)
}

// --------------------------
// IFD (Image File Directory)
// --------------------------

type TIFFDirectory struct {
	tags map[tags.TagID]TIFFTag
}

func NewTIFFDirectory(tags map[tags.TagID]TIFFTag) TIFFDirectory {
	return TIFFDirectory{tags: tags}
}

func (d TIFFDirectory) Tag(tagID tags.TagID) (TIFFTag, error) {
	if v, ok := d.tags[tagID]; ok {
		return v, nil
	}
	return nil, fmt.Errorf("tag not found: %s", tags.IDsLabels[tagID])
}

func (d TIFFDirectory) Tags(tagIDs ...tags.TagID) ([]TIFFTag, error) {
	r := make([]TIFFTag, len(tagIDs))
	for i, tID := range tagIDs {
		t, err := d.Tag(tID)
		if err != nil {
			return r, err
		}
		r[i] = t
	}
	return r, nil
}

func (d TIFFDirectory) String() string {
	return fmt.Sprintf("%v", d.tags)
}
