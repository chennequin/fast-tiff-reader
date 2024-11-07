package main

import (
	"TiffReader/internal/slides"
	"fmt"
	"log"
	"log/slog"
	"os"
	"time"
)

func main() {
	//slog.SetLogLoggerLevel(slog.LevelInfo)
	slog.SetLogLoggerLevel(slog.LevelDebug)
	//slog.SetLogLoggerLevel(tiffio.LogLevelTrace)

	//name := "assets/generic/CMU-1.tiff"
	//name := "assets/philips/Philips-1.tiff"
	//name := "assets/philips/Philips-2.tiff"
	//name := "assets/philips/Philips-3.tiff"
	//name := "assets/philips/Philips-4.tiff"
	//name := "assets/trestle/CMU-1.tif"
	//name := "assets/trestle/CMU-2.tif"
	//name := "assets/trestle/CMU-3.tif"
	//name := "assets/leica/Leica-1.scn"
	//name := "assets/leica/Leica-2.scn"
	//name := "assets/leica/Leica-3.scn"
	//name := "assets/leica/Leica-Fluorescence-1.scn"
	//name := "assets/aperio/CMU-1.svs"
	//name := "assets/aperio/CMU-2.svs"
	//name := "assets/aperio/CMU-3.svs"
	//name := "assets/aperio/CMU-1-JP2K-33005.svs"
	//name := "assets/aperio/CMU-1-Small-Region.svs"
	//name := "assets/aperio/JP2K-33003-1.svs"
	//name := "assets/aperio/JP2K-33003-2.svs"
	//name := "assets/ventana/OS-1.bif"
	//name := "assets/ventana/OS-2.bif"
	//name := "assets/camelyon16/test_001.tif"
	name := "assets/camelyon16/test_002.tif"

	start := time.Now()

	reader := slides.NewSlideReader()
	err := reader.OpenFile(name)
	if err != nil {
		log.Fatalf("%v", err)
	}

	elapsed := time.Since(start)
	fmt.Printf("Opening execution time: %s\n", elapsed)

	levelIdx := reader.LevelCount() - 1

	start = time.Now()

	_, err = reader.GetMetaData()
	if err != nil {
		log.Fatalf("%v", err)
	}

	tile, err := reader.GetTile(levelIdx, 0)
	if err != nil {
		log.Fatalf("unable to read tile %d at level %d/%d: %v", 0, levelIdx, reader.LevelCount()-1, err)
	}

	output := fmt.Sprintf("tile_%d_%d.jpeg", levelIdx, 0)
	err = os.WriteFile(output, tile, os.ModePerm)
	if err != nil {
		log.Fatalf("Error writing file: %v", err)
	}

	elapsed = time.Since(start)
	fmt.Printf("Reading execution time: %s\n", elapsed)

	println("OK")
}
