package handlers

import (
	"TiffReader/internal/slides"
)

type SlideReaderCacheEntry struct {
	reader   *slides.SlideReader
	metadata slides.PyramidMetadata
}

type SlideReaderCache struct {
}

func NewSlideReaderCache() *SlideReaderCache {
	return &SlideReaderCache{}
}
