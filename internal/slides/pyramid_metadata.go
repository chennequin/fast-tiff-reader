package slides

type PyramidMetadata struct {
	Levels []PyramidImage
}

type PyramidImage struct {
	ImageWidth          int
	ImageHeight         int
	TileWidth           int
	TileHeight          int
	TileCountHorizontal int
	TileCountVertical   int
}

func (i PyramidImage) TileIndex(tileX, tileY int) int {
	numTilesHorizontal := i.ImageWidth / i.TileWidth
	if i.ImageWidth%i.TileWidth > 0 {
		numTilesHorizontal += 1
	}
	idx := tileY*numTilesHorizontal + tileX
	return idx
}
