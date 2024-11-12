package main

import (
	"TiffReader/internal/handlers"
	"TiffReader/internal/slides"
	"github.com/gin-gonic/gin"
	"github.com/scalalang2/golang-fifo/sieve"
	"github.com/spf13/viper"
	"log/slog"
)

const (
	assetsDirectory = "assets"
)

func init() {
	viper.SetDefault("assets.directory", "assets")
	readerCache = sieve.New[string, *slides.SlideReader](100, 0)
}

var readerCache *sieve.Sieve[string, *slides.SlideReader]

func main() {
	slog.SetLogLoggerLevel(slog.LevelDebug)
	r := gin.Default()

	dir := viper.GetString("assets.directory")
	h := handlers.NewTileHandlers(dir)

	r.GET("/open/file/*path", h.HandleOpenFile)
	r.GET("/open/S3/*path", h.HandleOpenS3)
	r.GET("files/:tiff/levels/:level/tiles/:xy", h.HandleGetTile)

	err := r.Run()
	if err != nil {
		return
	}
}
