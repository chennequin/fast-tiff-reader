package handlers

import (
	"TiffReader/internal/slides"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/jxskiss/base62"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
)

type TileHandlers struct {
	assetsDirectory string
	cache           *SlideReaderCache
}

func NewTileHandlers(directory string, cache *SlideReaderCache) *TileHandlers {
	return &TileHandlers{
		assetsDirectory: directory,
		cache:           cache,
	}
}

func (t *TileHandlers) HandleGetFileTile(c *gin.Context) {
	tiffFile, levelIdx, x, y, err := t.handleTileParams(c)
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

func (t *TileHandlers) HandleOpenFile(c *gin.Context) {
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

func (t *TileHandlers) HandleOpenS3(c *gin.Context) {

}

func (t *TileHandlers) handleTileParams(c *gin.Context) (string, int, int, int, error) {
	encoded := c.Param("tiff")
	level := c.Param("level")
	xyParam := c.Param("xy")

	decoded, err := base62.DecodeString(encoded)
	if err != nil {
		return "", 0, 0, 0, fmt.Errorf("failed to base62 decode path")
	}
	tiffFile := string(decoded)

	xy := strings.TrimSuffix(xyParam, ".jpeg")
	if xy == xyParam {
		return "", 0, 0, 0, fmt.Errorf("invalid tile format, expected .jpeg")
	}

	coordinates := strings.Split(xy, "_")
	if len(coordinates) != 2 {
		return "", 0, 0, 0, fmt.Errorf("invalid tile coordinates")
	}

	x, err := strconv.Atoi(coordinates[0])
	if err != nil {
		return "", 0, 0, 0, fmt.Errorf("conversion error for coordinate x")
	}

	y, err := strconv.Atoi(coordinates[1])
	if err != nil {
		return "", 0, 0, 0, fmt.Errorf("conversion error for coordinate y")
	}

	levelIdx, err := strconv.Atoi(level)
	if err != nil {
		return "", 0, 0, 0, fmt.Errorf("invalid level")
	}

	return tiffFile, levelIdx, x, y, nil
}

func (t *TileHandlers) openFileReader(tiffFile string) (*slides.SlideReader, *slides.PyramidMetadata, error) {
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
