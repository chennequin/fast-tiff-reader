package handlers

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/jxskiss/base62"
	"strconv"
	"strings"
)

func handleTileParams(c *gin.Context) (string, int, int, int, error) {
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
