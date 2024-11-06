package model

import (
	"TiffReader/internal/tiffio/tags"
	"fmt"
)

// --------------------------
// TIFF (Tagged Image File Format)
// --------------------------

type TIFFMetadata struct {
	entries  []TIFFDirectory
	pyramids map[string][]TIFFDirectory
}

func NewTIFFMetadata(entries []TIFFDirectory) TIFFMetadata {
	m := make(map[string][]TIFFDirectory)
	for _, entry := range entries {
		pyramidID := entry.GetPyramidID()
		m[pyramidID] = append(m[pyramidID], entry)
	}
	return TIFFMetadata{entries: entries, pyramids: m}
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
	return fmt.Sprintf("%v", t.pyramids)
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

func (d TIFFDirectory) GetPyramidID() string {
	getOrZero := func(t tags.TagID) int {
		r, _ := d.GetIntTag(t)
		return r
	}
	width := getOrZero(tags.TileWidth)
	height := getOrZero(tags.TileLength)
	rowsPerStrip := getOrZero(tags.RowsPerStrip)
	subFileType := getOrZero(tags.NewSubfileType)
	return fmt.Sprintf("TileWidth:%d, TileLength:%d, RowsPerStrip:%d, NewSubfileType:%s",
		width, height, rowsPerStrip, subFileType)
}

func (d TIFFDirectory) GetImageWidth() (int, error) {
	return d.GetIntTag(tags.ImageWidth)
}

func (d TIFFDirectory) GetImageHeight() (int, error) {
	return d.GetIntTag(tags.ImageLength)
}

func (d TIFFDirectory) GetTileWidth() (int, error) {
	return d.GetIntTag(tags.TileWidth)
}

func (d TIFFDirectory) GetTileHeight() (int, error) {
	return d.GetIntTag(tags.TileLength)
}

func (d TIFFDirectory) GetTileCount() (int, error) {
	tag, err := d.Tag(tags.TileOffsets)
	if err != nil {
		return 0, err
	}
	return tag.ValuesCount(), nil
}

func (d TIFFDirectory) GetStripCount() (int, error) {
	tag, err := d.Tag(tags.StripOffsets)
	if err != nil {
		return 0, err
	}
	return tag.ValuesCount(), nil
}

func (d TIFFDirectory) GetRowsPerStrip() (int, error) {
	return d.GetIntTag(tags.RowsPerStrip)
}

func (d TIFFDirectory) GetCompression() (tags.CompressionType, error) {
	compression, err := d.GetIntTag(tags.Compression)
	if err != nil {
		return 0, err
	}
	return tags.CompressionType(compression), nil
}

func (d TIFFDirectory) GetPhotometricInterpretation() (tags.PhotometricInterpretationType, error) {
	compression, err := d.GetIntTag(tags.PhotometricInterpretation)
	if err != nil {
		return 0, err
	}
	return tags.PhotometricInterpretationType(compression), nil
}

func (d TIFFDirectory) GetPredictor() (tags.PredictorType, error) {
	predictor, err := d.GetIntTag(tags.Predictor)
	if err != nil {
		return 0, err
	}
	return tags.PredictorType(predictor), nil
}

func (d TIFFDirectory) GetIntTag(tagID tags.TagID) (int, error) {
	tag, err := d.Tag(tagID)
	if err != nil {
		return 0, err
	}
	return int(tag.GetUintVal(0)), err
}

func (d TIFFDirectory) Tag(tagID tags.TagID) (TIFFTag, error) {
	if v, ok := d.tags[tagID]; ok {
		return v, nil
	}
	return nil, NewTagNotFoundError(tagID)
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
