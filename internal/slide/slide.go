package slide

import (
	"TiffReader/internal/tiffio/model"
	"TiffReader/internal/tiffio/tags"
)

func TileDimensions(img model.TIFF, level, tileIndex int) (int, int) {
	//TODO avoid panic
	imageWidth := int(img.Level(level).Tag(tags.ImageWidth).AsUint16s()[0])
	imageLength := int(img.Level(level).Tag(tags.ImageLength).AsUint16s()[0])
	tileWidth := int(img.Level(level).Tag(tags.TileWidth).AsUint16s()[0])
	tileLength := int(img.Level(level).Tag(tags.TileLength).AsUint16s()[0])

	numTilesHorizontal := imageWidth / tileWidth
	numTilesVertical := imageLength / tileLength
	lastTileWidth := imageWidth % tileWidth
	lastTileLength := imageLength % tileLength

	if lastTileWidth > 0 {
		numTilesHorizontal += 1
	}
	if lastTileLength > 0 {
		numTilesVertical += 1
	}

	tilePosX := tileIndex % numTilesHorizontal
	tilePosY := tileIndex / numTilesVertical

	tileActualWidth := tileWidth
	tileActualLength := tileLength

	if tilePosX == numTilesHorizontal-1 {
		tileActualWidth = lastTileWidth
	}
	if tilePosY == numTilesVertical-1 {
		tileActualLength = lastTileLength
	}

	return tileActualWidth, tileActualLength
}
