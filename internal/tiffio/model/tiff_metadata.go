package model

import (
	"fmt"
)

type TIFFMetadata struct {
	Entries []TIFFDirectory
}

func NewTIFFMetadata(entries []TIFFDirectory) TIFFMetadata {
	return TIFFMetadata{Entries: entries}
}

func (t TIFFMetadata) Level(level int) (TIFFDirectory, error) {
	if level >= len(t.Entries) {
		return TIFFDirectory{}, fmt.Errorf("level out of range: %d", level)
	}
	return t.Entries[level], nil
}

func (t TIFFMetadata) LevelCount() int {
	return len(t.Entries)
}

func (t TIFFMetadata) String() string {
	return fmt.Sprintf("%v", t.Entries)
}
