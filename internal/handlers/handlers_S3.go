package handlers

import "github.com/gin-gonic/gin"

type S3Handlers struct {
	cache *SlideReaderCache
}

func NewS3Handlers(cache *SlideReaderCache) *S3Handlers {
	return &S3Handlers{cache}
}

func (t *S3Handlers) HandleGetTile(c *gin.Context) {

}

func (t *S3Handlers) HandleOpenS3(c *gin.Context) {

}
