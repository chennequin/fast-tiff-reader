package handlers

import (
	"TiffReader/internal/slides"
	"github.com/scalalang2/golang-fifo/sieve"
	"github.com/scalalang2/golang-fifo/types"
)

type SlideReaderCacheEntry struct {
	reader   *slides.SlideReader
	metadata *slides.PyramidMetadata
}

type SlideReaderCache struct {
	cache *sieve.Sieve[string, SlideReaderCacheEntry]
}

func NewSlideReaderCache(cacheSize int) *SlideReaderCache {
	cache := sieve.New[string, SlideReaderCacheEntry](cacheSize, 0)
	cache.SetOnEvicted(func(key string, value SlideReaderCacheEntry, reason types.EvictReason) {
		value.reader.Close()
	})
	return &SlideReaderCache{cache: cache}
}

func (c *SlideReaderCache) Get(tiffFile string) (*slides.SlideReader, *slides.PyramidMetadata, bool) {
	if entry, ok := c.cache.Get(tiffFile); ok {
		return entry.reader, entry.metadata, true
	}
	return nil, nil, false
}

func (c *SlideReaderCache) Set(tiffFile string, reader *slides.SlideReader, metadata *slides.PyramidMetadata) {
	entry := SlideReaderCacheEntry{
		reader:   reader,
		metadata: metadata,
	}
	c.cache.Set(tiffFile, entry)
}

func (c *SlideReaderCache) Close() {
	c.cache.Close()
}
