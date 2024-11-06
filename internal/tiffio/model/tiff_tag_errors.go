package model

import (
	"TiffReader/internal/tiffio/tags"
	"fmt"
)

type TagNotFoundError struct {
	tagID tags.TagID
}

func NewTagNotFoundError(tagID tags.TagID) TagNotFoundError {
	return TagNotFoundError{tagID: tagID}
}

func (e TagNotFoundError) Error() string {
	return fmt.Sprintf("tag not found: %s", tags.IDsLabels[e.tagID])
}
