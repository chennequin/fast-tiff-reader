package main

import (
	"TiffReader/internal/handlers"
	"context"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
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
	defer cache.Close()

	dir := viper.GetString("assets.directory")
	hf := handlers.NewFileHandlers(dir, cache)

	hs3 := handlers.NewS3Handlers(cache)

	r.GET("/open/file/*path", hf.HandleOpenFile)
	r.GET("/open/S3/*path", hs3.HandleOpenS3)
	r.GET("files/:tiff/levels/:level/tiles/:xy", hf.HandleGetTile)
	r.GET("S3/:tiff/levels/:level/tiles/:xy", hs3.HandleOpenS3)

	server := &http.Server{
		Handler: r,
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := r.Run(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("Error in HTTP server", "error", err)
		}
	}()

	sigReceived := <-stop
	slog.Info(fmt.Sprintf("Signal reçu: %s, beginning graceful shutdown", sigReceived))

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		slog.Error("Error while stopping server", "error", err)
	} else {
		slog.Info("Server stopped.")
	}
}
