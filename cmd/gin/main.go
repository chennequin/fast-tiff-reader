package main

import (
	"TiffReader/internal/handlers"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"log/slog"
)

const (
	assetsDirectory = "assets"
	cacheSize       = 100
)

func init() {
	viper.SetDefault("assets.directory", assetsDirectory)
	viper.SetDefault("reader.cache.size", cacheSize)
}

func main() {
	slog.SetLogLoggerLevel(slog.LevelDebug)
	r := gin.Default()

	size := viper.GetInt("reader.cache.size")
	cache := handlers.NewSlideReaderCache(size)

	dir := viper.GetString("assets.directory")
	h := handlers.NewTileHandlers(dir, cache)

	r.GET("/open/file/*path", h.HandleOpenFile)
	r.GET("/open/S3/*path", h.HandleOpenS3)
	r.GET("files/:tiff/levels/:level/tiles/:xy", h.HandleGetFileTile)

	err := r.Run()
	if err != nil {
		return
	}
}
