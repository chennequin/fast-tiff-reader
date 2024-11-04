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
	slog.SetLogLoggerLevel(slog.LevelDebug)

	//name := "assets/CMU-1.tiff"
	name := "assets/Philips-1.tiff"

	reader := slides.NewSlideReader()
	err := reader.OpenFile(name)
	if err != nil {
		log.Fatalf("%v", err)
	}

	levelIdx := reader.LevelCount() - 2
	tileIdx := 0

	start := time.Now()

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
