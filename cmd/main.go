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
	name := "assets/aperio/CMU-2.svs"
	//name := "assets/aperio/CMU-3.svs"
	//name := "assets/aperio/CMU-1-JP2K-33005.svs"
	//name := "assets/aperio/CMU-1-Small-Region.svs"
	//name := "assets/aperio/JP2K-33003-1.svs"
	//name := "assets/aperio/JP2K-33003-2.svs"
	//name := "assets/ventana/OS-1.bif"
	//name := "assets/ventana/OS-2.bif"
	//name := "assets/ventana/Ventana-1.bif"

	start := time.Now()

	reader := slides.NewSlideReader()
	err := reader.OpenFile(name)
	if err != nil {
		log.Fatalf("%v", err)
	}

	levelIdx := reader.LevelCount() - 1
	tileIdx := 0

	tile, err := reader.GetTile(levelIdx, tileIdx)
	if err != nil {
		log.Fatalf("unable to read tile %d at level %d/%d: %v", tileIdx, levelIdx, reader.LevelCount()-1, err)
	}

	elapsed := time.Since(start) // Calculate elapsed time
	fmt.Printf("Execution time: %s\n", elapsed)

	output := fmt.Sprintf("tile_%d_%d.jpeg", levelIdx, tileIdx)
	err = os.WriteFile(output, tile, os.ModePerm)
	if err != nil {
		log.Fatalf("Error writing file: %v", err)
	}

	println("OK")
}
