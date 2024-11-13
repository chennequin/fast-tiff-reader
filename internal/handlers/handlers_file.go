package handlers

import (
	"TiffReader/internal/slides"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/jxskiss/base62"
	"log/slog"
	"net/http"
	"strings"
)

type FileHandlers struct {
	assetsDirectory string
	cache           *SlideReaderCache
}

func NewFileHandlers(directory string, cache *SlideReaderCache) *FileHandlers {
	return &FileHandlers{
		assetsDirectory: directory,
		cache:           cache,
	}
}

func (t *FileHandlers) HandleGetTile(c *gin.Context) {
	tiffFile, levelIdx, x, y, err := handleTileParams(c)
	if err != nil {
		slog.Error("Error opening file", "file", tiffFile, "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	reader, metadata, ok := t.cache.Get(tiffFile)
	if !ok {
		reader, metadata, err = t.openFileReader(tiffFile)
		if err != nil {
			slog.Error("Error opening file", "file", tiffFile, "error", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to open file"})
			return
		}
	}

	tileIdx := metadata.Levels[levelIdx].TileIndex(x, y)
	imageData, err := reader.GetTile(levelIdx, tileIdx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read tile"})
		return
	}

	c.Data(http.StatusOK, "image/jpeg", imageData)
}

func (t *FileHandlers) HandleOpenFile(c *gin.Context) {
	tiffFile := strings.TrimPrefix(c.Param("path"), "/")

	// Encode the resource path in a URL-friendly format
	encoded := base62.EncodeToString([]byte(tiffFile))

	if _, metadata, ok := t.cache.Get(tiffFile); ok {
		c.JSON(200, gin.H{
			"encoded":  encoded,
			"decoded":  tiffFile,
			"metadata": metadata,
		})
		return
	}

	_, metadata, err := t.openFileReader(tiffFile)
	if err != nil {
		slog.Error("Error opening file", "file", tiffFile, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to open file"})
		return
	}

	c.JSON(200, gin.H{
		"encoded":  encoded,
		"decoded":  tiffFile,
		"metadata": metadata,
	})
}

func (t *FileHandlers) openFileReader(tiffFile string) (*slides.SlideReader, *slides.PyramidMetadata, error) {
	// Determine the full path to the underlying resource (file) to be accessed
	name := fmt.Sprintf("%s/%s", t.assetsDirectory, tiffFile)

	// Open the resource and retrieve its metadata
	reader := slides.NewSlideReader()
	err := reader.OpenFile(name)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to open image file: %w", err)
	}

	metadata, err := reader.GetMetadata()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read metadata from: %w", err)
	}

	// put in memory cache
	t.cache.Set(tiffFile, reader, &metadata)

	return reader, &metadata, nil
}
