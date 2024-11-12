package handlers

import (
	"TiffReader/internal/slides"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/jxskiss/base62"
	"github.com/scalalang2/golang-fifo/sieve"
	"github.com/scalalang2/golang-fifo/types"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
)

const (
	cacheSize = 100
)

var readerCache *sieve.Sieve[string, SlideReaderCacheEntry]

type TileHandlers struct {
	assetsDirectory string
}

func init() {
	readerCache = sieve.New[string, SlideReaderCacheEntry](cacheSize, 0)
	readerCache.SetOnEvicted(func(key string, value SlideReaderCacheEntry, reason types.EvictReason) {
		value.reader.Close()
	})
}

func NewTileHandlers(directory string) *TileHandlers {
	return &TileHandlers{
		assetsDirectory: directory,
	}
}

func (t *TileHandlers) HandleGetTile(c *gin.Context) {
	encoded := c.Param("tiff")
	level := c.Param("level")
	xyParam := c.Param("xy")

	decoded, err := base62.DecodeString(encoded)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to base62 decode path"})
		return
	}
	tiffFile := string(decoded)

	xy := strings.TrimSuffix(xyParam, ".jpeg")
	if xy == xyParam {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid tile format, expected .jpeg"})
		return
	}

	coordinates := strings.Split(xy, "_")
	if len(coordinates) != 2 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid tile coordinates"})
		return
	}

	x, err := strconv.Atoi(coordinates[0])
	if err != nil {
		fmt.Println("Conversion error for coordinate x:", err)
		return
	}

	y, err := strconv.Atoi(coordinates[1])
	if err != nil {
		fmt.Println("Conversion error for coordinate y:", err)
		return
	}

	levelIdx, err := strconv.Atoi(level)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid level"})
		return
	}

	var reader *slides.SlideReader
	var metadata slides.PyramidMetadata

	if entry, ok := readerCache.Get(tiffFile); ok {
		reader = entry.reader
		metadata = entry.metadata
	} else {
		entry, err = t.openFileReader(tiffFile)
		if err != nil {
			slog.Error("Error opening file", "file", tiffFile, "error", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to open file"})
			return
		}
		reader = entry.reader
		metadata = entry.metadata
	}

	tileIdx := metadata.Levels[levelIdx].TileIndex(x, y)
	imageData, err := reader.GetTile(levelIdx, tileIdx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read tile"})
		return
	}

	c.Data(http.StatusOK, "image/jpeg", imageData)
}

func (t *TileHandlers) HandleOpenFile(c *gin.Context) {
	tiffFile := strings.TrimPrefix(c.Param("path"), "/")

	// Encode the resource path in a URL-friendly format
	encoded := base62.EncodeToString([]byte(tiffFile))

	if entry, ok := readerCache.Get(tiffFile); ok {
		c.JSON(200, gin.H{
			"encoded":  encoded,
			"decoded":  tiffFile,
			"metadata": entry.metadata,
		})
		return
	}

	entry, err := t.openFileReader(tiffFile)
	if err != nil {
		slog.Error("Error opening file", "file", tiffFile, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to open file"})
		return
	}

	c.JSON(200, gin.H{
		"encoded":  encoded,
		"decoded":  tiffFile,
		"metadata": entry.metadata,
	})
}

func (t *TileHandlers) HandleOpenS3(c *gin.Context) {

}

func (t *TileHandlers) openFileReader(tiffFile string) (SlideReaderCacheEntry, error) {
	// Determine the full path to the underlying resource (file) to be accessed
	name := fmt.Sprintf("%s/%s", t.assetsDirectory, tiffFile)

	// Open the resource and retrieve its metadata
	reader := slides.NewSlideReader()
	err := reader.OpenFile(name)
	if err != nil {
		return SlideReaderCacheEntry{}, fmt.Errorf("failed to open image file: %w", err)
	}

	metadata, err := reader.GetMetadata()
	if err != nil {
		return SlideReaderCacheEntry{}, fmt.Errorf("failed to read metadata from: %w", err)
	}

	// put in memory cache
	entry := SlideReaderCacheEntry{
		reader:   reader,
		metadata: metadata,
	}
	readerCache.Set(tiffFile, entry)

	return entry, nil
}
