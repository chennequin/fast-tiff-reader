package slides

import (
	"TiffReader/internal/jpegio"
	"TiffReader/internal/tiffio"
	tiffModel "TiffReader/internal/tiffio/model"
	"TiffReader/internal/tiffio/tags"
	"bytes"
	"errors"
	"fmt"
	"golang.org/x/image/tiff/lzw"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"log/slog"
)

type SlideReader struct {
	pyramid SlideMetadata
	reader  *tiffio.TiffReader
}

func NewSlideReader() *SlideReader {
	return &SlideReader{}
}

func (r *SlideReader) OpenFile(name string) error {
	binaryReader := tiffio.NewFileBinaryReader()
	cacheBinaryReader := tiffio.NewCacheBinaryReader(binaryReader)
	tiffReader := tiffio.NewTiffReader(cacheBinaryReader)

	err := tiffReader.Open(name)
	if err != nil {
		return err
	}

	metadata, err := tiffReader.ReadMetadata()
	if err != nil {
		return fmt.Errorf("unable to read Metadata: %w", err)
	}

	// partition the directories.
	// extract the main pyramid from extra stripped images.
	m := make(map[string]tiffModel.TIFFMetadata)
	for _, directory := range metadata {
		pyramidID := directory.GetPyramidID()
		m[pyramidID] = append(m[pyramidID], directory)
	}

	// extract the largest pyramid of images
	var largestPyramidKey string
	var extra tiffModel.TIFFMetadata
	for k, directories := range m {
		if len(largestPyramidKey) < len(directories) {
			largestPyramidKey = k
		}
	}

	// flatten the rest of images
	for k, directories := range m {
		if k != largestPyramidKey {
			extra = append(extra, directories...)
		}
	}

	r.reader = tiffReader
	r.pyramid = SlideMetadata{
		Directories: m[largestPyramidKey],
		ExtraImages: extra,
	}

	return nil
}

func (r *SlideReader) Close() {
	r.reader.Close()
	r.pyramid = SlideMetadata{}
}

func (r *SlideReader) GetMetadata() (PyramidMetadata, error) {
	var pyramid PyramidMetadata
	pyramid.Levels = make([]PyramidImage, 0)
	for _, level := range r.pyramid.Directories {
		imageTags, err := level.Tags(tags.ImageWidth, tags.ImageLength, tags.TileWidth, tags.TileLength)
		if err != nil {
			return pyramid, fmt.Errorf("missing required tags: %w", err)
		}

		imageWidth := int(imageTags[0].GetUintVal(0))
		imageLength := int(imageTags[1].GetUintVal(0))
		tileWidth := int(imageTags[2].GetUintVal(0))
		tileLength := int(imageTags[3].GetUintVal(0))

		tileCountHorizontal := imageWidth / tileWidth
		if imageWidth%tileWidth > 0 {
			tileCountHorizontal += 1
		}

		tileCountVertical := imageLength / tileLength
		if imageLength%tileLength > 0 {
			tileCountVertical += 1
		}

		l := PyramidImage{
			ImageWidth:          imageWidth,
			ImageHeight:         imageLength,
			TileWidth:           tileWidth,
			TileHeight:          tileLength,
			TileCountHorizontal: tileCountHorizontal,
			TileCountVertical:   tileCountVertical,
		}

		pyramid.Levels = append(pyramid.Levels, l)
	}
	slog.Debug("Pyramid Metadata", "levels", len(pyramid.Levels), "metadata", pyramid)
	return pyramid, nil
}

func (r *SlideReader) GetTile(levelIdx, tileIdx int) ([]byte, error) {
	level, err := r.pyramid.Level(levelIdx)
	if err != nil {
		return nil, fmt.Errorf("unable to get level %d: %w", levelIdx, err)
	}
	tile, err := r.getRawTileJPEG(level, tileIdx)
	if err != nil {
		if errors.Is(err, tiffModel.NewTagNotFoundError(tags.TileOffsets)) {
			return r.recomposeStripImage(level)
		}
	}
	return tile, err
}

func (r *SlideReader) GetExtraImage(levelIdx int) ([]byte, error) {
	level, err := r.pyramid.Extra(levelIdx)
	if err != nil {
		return nil, fmt.Errorf("unable to get level %d: %w", levelIdx, err)
	}
	data, err := r.recomposeStripImage(level)
	if err != nil {
		return nil, fmt.Errorf("unable to get level %d: %w", levelIdx, err)
	}
	return data, nil
}

func (r *SlideReader) getRawTileJPEG(level tiffModel.TIFFDirectory, tileIdx int) ([]byte, error) {
	data, err := r.reader.GetTileData(level, tileIdx)
	if err != nil {
		return nil, fmt.Errorf("getRawTileJPEG: unable to obtain tile data: %w", err)
	}

	var tileWidth, tileHeight int
	var encoded []byte
	var jpegTablesErr, iccProfileErr error
	var jpegTables, iccProfile []byte

	jpegTables, jpegTablesErr = level.GetJPEGTables()
	iccProfile, iccProfileErr = r.retrieveIccProfile(level)

	if jpegTablesErr == nil || iccProfileErr == nil {
		tileWidth, tileHeight, encoded, err = jpegio.MergeSegments(data, jpegTables, iccProfile)
		if err != nil {
			return nil, fmt.Errorf("getRawTileJPEG: unable to merge JPEG segments: %w", err)
		}
	} else {
		// JPEGTables and ICC Profile are not present in Metadata
		tileWidth, tileHeight, err = jpegio.DecodeSOF(data)
		if err != nil {
			return nil, fmt.Errorf("getRawTileJPEG: unable to decode JPEG segment: %w", err)
		}
		encoded = data
	}

	expectedWidth, expectedHeight, err := r.calculateTileWidthHeight(level, tileIdx)
	if err != nil {
		return nil, fmt.Errorf("getRawTileJPEG: unable to calculate expected tile size: %w", err)
	}

	if tileWidth != expectedWidth || tileHeight != expectedHeight {
		encoded, err = r.cropImageJPEG(expectedWidth, expectedHeight, encoded)
	}

	return encoded, err
}

func (r *SlideReader) recomposeStripImage(level tiffModel.TIFFDirectory) ([]byte, error) {
	stripCount, err := level.GetStripCount()
	if err != nil {
		return nil, err
	}
	widthImage, err := level.GetImageWidth()
	if err != nil {
		return nil, err
	}
	heightImage, err := level.GetImageHeight()
	if err != nil {
		return nil, err
	}
	rowsPerStrip, err := level.GetRowsPerStrip()
	if err != nil {
		return nil, err
	}
	compression, err := level.GetCompression()
	if err != nil {
		return nil, err
	}
	finalImage := image.NewRGBA(image.Rect(0, 0, widthImage, heightImage))
	for stripIdx := range stripCount {
		data, err := r.getRawStrip(level, stripIdx)
		if err != nil {
			return nil, err
		}

		switch compression {
		case tags.CompressionTypeJPEG:
			img, err := jpeg.Decode(bytes.NewReader(data))
			if err != nil {
				return nil, fmt.Errorf("recomposeStripImage: unable to decode JPEG: %w", err)
			}
			rect := image.Rect(0, rowsPerStrip*stripIdx, 0+widthImage, rowsPerStrip*stripIdx+rowsPerStrip)
			draw.Draw(finalImage, rect, img, image.Point{}, draw.Over)

		case tags.CompressionTypeLZW:
			lzwReader := lzw.NewReader(bytes.NewReader(data), lzw.MSB, 8)

			var decompressedData []byte
			buf := make([]byte, 4096)
			for {
				n, err := lzwReader.Read(buf)
				if err != nil {
					if err.Error() != "EOF" {
						return nil, fmt.Errorf("recomposeStripImage: unable to read lzw: %w", err)
					}
					break
				}
				decompressedData = append(decompressedData, buf[:n]...)
			}
			err = lzwReader.Close()
			if err != nil {
				return nil, fmt.Errorf("recomposeStripImage: unable to close lzw reader: %w", err)
			}

			photometricInterpretation, err := level.GetPhotometricInterpretation()
			if err != nil {
				return nil, err
			}

			predictor, err := level.GetPredictor()
			if err != nil {
				return nil, err
			}

			switch photometricInterpretation {
			case tags.PhotometricInterpretationTypeRGB:
				img := image.NewRGBA(image.Rect(0, 0, widthImage, rowsPerStrip))
				for y := 0; y < rowsPerStrip; y++ {
					var previousRed, previousGreen, previousBlue byte
					for x := 0; x < widthImage; x++ {
						if index := (y*widthImage + x) * 3; index < len(decompressedData) {
							var red, green, blue byte
							if predictor == tags.PredictorTypeHorizontalDifferencing {
								red = previousRed + decompressedData[index]
								green = previousGreen + decompressedData[index+1]
								blue = previousBlue + decompressedData[index+2]
							} else {
								red = decompressedData[index]
								green = decompressedData[index+1]
								blue = decompressedData[index+2]
							}
							c := color.RGBA{
								R: red,
								G: green,
								B: blue,
								A: 255,
							}
							img.Set(x, y, c)

							previousRed = red
							previousGreen = green
							previousBlue = blue
						}
					}
				}
				rect := image.Rect(0, rowsPerStrip*stripIdx, 0+widthImage, rowsPerStrip*stripIdx+rowsPerStrip)
				draw.Draw(finalImage, rect, img, image.Point{}, draw.Over)

			case tags.PhotometricInterpretationTypeYCbCr:
				img := image.NewYCbCr(image.Rect(0, 0, widthImage, rowsPerStrip), image.YCbCrSubsampleRatio422)

				copy(img.Y, decompressedData[:len(img.Y)])
				copy(img.Cb, decompressedData[len(img.Y):len(img.Y)+len(img.Cb)])
				copy(img.Cr, decompressedData[len(img.Y)+len(img.Cb):])

				rect := image.Rect(0, rowsPerStrip*stripIdx, 0+widthImage, rowsPerStrip*stripIdx+rowsPerStrip)
				draw.Draw(finalImage, rect, img, image.Point{}, draw.Over)

				return nil, fmt.Errorf("recomposeStripImage: unsupported PhotometricInterpretation type: %v", compression)
			}

		default:
			return nil, fmt.Errorf("recomposeStripImage: unsupported Compression type: %v", compression)
		}
	}

	buf := bytes.NewBuffer(make([]byte, 0))
	if err = jpeg.Encode(buf, finalImage, nil); err != nil {
		return nil, fmt.Errorf("recomposeStripImage: unable to encode JPEG: %w", err)
	}
	return buf.Bytes(), err
}

func (r *SlideReader) getRawStrip(level tiffModel.TIFFDirectory, stripIdx int) ([]byte, error) {
	compression, err := level.GetCompression()
	if err != nil {
		return nil, err
	}

	var data []byte
	switch compression {
	case tags.CompressionTypeJPEG:
		data, err = r.getRawStripJPEG(level, stripIdx)
	case tags.CompressionTypeLZW:
		data, err = r.getRawStripLZW(level, stripIdx)
	default:
		return nil, fmt.Errorf("getRawStrip: unsupported compression type: %v", compression)
	}
	return data, err
}

func (r *SlideReader) getRawStripJPEG(level tiffModel.TIFFDirectory, stripIdx int) ([]byte, error) {
	data, err := r.reader.GetStripData(level, stripIdx)
	if err != nil {
		return nil, fmt.Errorf("getRawStripJPEG: unable to obtain strip data: %w", err)
	}

	var jpegTablesErr, iccProfileErr error
	var jpegTables, iccProfile []byte

	jpegTables, jpegTablesErr = level.GetJPEGTables()
	iccProfile, iccProfileErr = r.retrieveIccProfile(level)

	if jpegTablesErr == nil || iccProfileErr == nil {
		_, _, data, err = jpegio.MergeSegments(data, jpegTables, iccProfile)
		if err != nil {
			return nil, fmt.Errorf("getRawStripJPEG: unable to merge JPEG segments: %w", err)
		}
	}

	// no crop operation to do in strips
	return data, err
}

func (r *SlideReader) getRawStripLZW(level tiffModel.TIFFDirectory, stripIdx int) ([]byte, error) {
	data, err := r.reader.GetStripData(level, stripIdx)
	if err != nil {
		return nil, fmt.Errorf("getRawStripLZW: unable to obtain strip data: %w", err)
	}
	return data, err
}

func (r *SlideReader) retrieveIccProfile(level tiffModel.TIFFDirectory) ([]byte, error) {
	iccProfile, err := level.GetIccProfile() // ICC profile at this level
	if err != nil {
		// get the ICC profile from the main image
		for i := range len(r.pyramid.Directories) {
			l, err := r.pyramid.Level(i)
			if err != nil {
				return nil, err
			}
			iccProfile, err = l.GetIccProfile()
			if err == nil {
				return iccProfile, nil
			}
		}
	}
	return iccProfile, nil
}

func (r *SlideReader) cropImageJPEG(expectedWidth, expectedHeight int, tileData []byte) ([]byte, error) {

	img, err := jpeg.Decode(bytes.NewReader(tileData))
	if err != nil {
		return nil, fmt.Errorf("cropImageJPEG: unable to decode JPEG: %w", err)
	}

	cropRect := image.Rect(0, 0, expectedWidth, expectedHeight)
	croppedImg, ok := img.(interface {
		SubImage(r image.Rectangle) image.Image
	})
	if !ok {
		return nil, fmt.Errorf("cropImageJPEG: image does not support cropping")
	}

	cropped := croppedImg.SubImage(cropRect)

	buf := bytes.NewBuffer(make([]byte, 0, len(tileData)))
	if err = jpeg.Encode(buf, cropped, nil); err != nil {
		return nil, fmt.Errorf("unable to encode JPEG: %w", err)
	}

	return buf.Bytes(), nil
}

func (r *SlideReader) calculateTileWidthHeight(level tiffModel.TIFFDirectory, tileIdx int) (int, int, error) {
	imageTags, err := level.Tags(tags.ImageWidth, tags.ImageLength, tags.TileWidth, tags.TileLength)
	if err != nil {
		return -1, -1, fmt.Errorf("missing required tags: %w", err)
	}

	imageWidth := int(imageTags[0].GetUintVal(0))
	imageLength := int(imageTags[1].GetUintVal(0))
	tileWidth := int(imageTags[2].GetUintVal(0))
	tileLength := int(imageTags[3].GetUintVal(0))

	actualWidth := actualTileWidth(imageWidth, tileWidth, tileIdx)
	actualHeight := actualTileHeight(imageLength, tileLength, tileIdx)

	return actualWidth, actualHeight, nil
}

func actualTileWidth(imageWidth, tileWidth, tileIdx int) int {
	if lastTileWidth := imageWidth % tileWidth; lastTileWidth > 0 {
		numTilesHorizontal := imageWidth/tileWidth + 1
		tilePosX := tileIdx % numTilesHorizontal
		if tilePosX >= numTilesHorizontal-1 {
			return lastTileWidth
		}
		return tileWidth
	}
	return tileWidth
}

func actualTileHeight(imageHeight, tileHeight, tileIdx int) int {
	if lastTileHeight := imageHeight % tileHeight; lastTileHeight > 0 {
		numTilesVertical := imageHeight/tileHeight + 1
		tilePosY := tileIdx / numTilesVertical
		if tilePosY >= numTilesVertical-1 {
			return lastTileHeight
		}
		return tileHeight
	}
	return tileHeight
}
