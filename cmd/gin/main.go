package main

import (
	"TiffReader/internal/slides"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/jxskiss/base62"
	"github.com/scalalang2/golang-fifo/sieve"
	"net/http"
	"strconv"
	"strings"
)

const (
	assetsDirectory = "assets"
	cacheSize       = 100
)

var readerCache *sieve.Sieve[string, *slides.SlideReader]

func init() {
	readerCache = sieve.New[string, *slides.SlideReader](cacheSize, 0)
}

func main() {
	r := gin.Default()

	r.GET("/open/*path", handleOpenFile)
	r.GET("/files/:tiff/metadata", handleMetadata)

	r.GET("files/:tiff/levels/:level/tiles/:xy", func(c *gin.Context) {
		encoded := c.Param("tiff")
		level := c.Param("level")
		xyParam := c.Param("xy")

		decoded, err := base62.DecodeString(encoded)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to base62 decode"})
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
			fmt.Println("Erreur de conversion pour la coordonnée x:", err)
			return
		}

		y, err := strconv.Atoi(coordinates[1])
		if err != nil {
			fmt.Println("Erreur de conversion pour la coordonnée y:", err)
			return
		}

		leveld, err := strconv.Atoi(level)
		if err != nil {
			// Si la conversion échoue, retourner une erreur
			c.JSON(400, gin.H{"error": "invalid level"})
			return
		}

		reader, err := getReader(tiffFile)
		if err != nil {
			c.JSON(http.StatusInternalServerError,
				gin.H{"error": fmt.Sprintf("Failed to read image file: %s", tiffFile)})
			return
		}

		metadata, err := reader.GetMetadata()
		if err != nil {
			c.JSON(http.StatusInternalServerError,
				gin.H{"error": fmt.Sprintf("Failed to read metadata from: %s", tiffFile)})
			return
		}

		levelIdx := leveld
		tileIdx := metadata.Levels[levelIdx].TileIndex(x, y)

		imageData, err := reader.GetTile(levelIdx, tileIdx)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read tile"})
			return
		}

		//c.Header("Content-Type", "image/jpeg")
		c.Data(http.StatusOK, "image/jpeg", imageData)

		c.JSON(200, gin.H{
			"message": "pong",
			"tiff":    tiffFile,
			"level":   level,
			"x":       coordinates[0],
			"y":       coordinates[1],
		})
	})

	err := r.Run() // listen and serve on 0.0.0.0:8080
	if err != nil {
		return
	}
}

func handleOpenFile(c *gin.Context) {
	tiffFile := strings.TrimPrefix(c.Param("path"), "/")
	encoded := base62.EncodeToString([]byte(tiffFile))
	name := fmt.Sprintf("assets/%s", tiffFile)
	reader := slides.NewSlideReader()
	err := reader.OpenFile(name)
	if err != nil {
		c.JSON(http.StatusInternalServerError,
			gin.H{"error": fmt.Sprintf("Failed to open image file: %s", tiffFile)})
		return
	}

	readerCache.Set(tiffFile, reader)

	c.Redirect(http.StatusFound, fmt.Sprintf("/files/%s/levels/%d/tiles/0_0.jpeg", encoded, reader.LevelCount()-1))
}

func handleServeTile(c *gin.Context) {

}

func handleMetadata(c *gin.Context) {
	//metadata, err := reader.GetMetadata()
	//if err != nil {
	//	c.JSON(http.StatusInternalServerError,
	//		gin.H{"error": fmt.Sprintf("Failed to read metadata from: %s", tiffFile)})
	//	return
	//}
}

func getReader(tiffFile string) (*slides.SlideReader, error) {
	if reader, ok := readerCache.Get(tiffFile); ok {
		return reader, nil
	}
	reader, err := openReader(tiffFile)
	if err != nil {
		return nil, err
	}
	readerCache.Set(tiffFile, reader)
	return reader, nil
}

func openReader(tiffFile string) (*slides.SlideReader, error) {
	path := fmt.Sprintf("%s/%s", assetsDirectory, tiffFile)
	reader := slides.NewSlideReader()
	err := reader.OpenFile(path)
	return reader, err
}
