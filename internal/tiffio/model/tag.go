package model

import "fmt"

type IFD struct {
	Tags    []Tag
	NextIFD int64
}

func (d IFD) String() string {
	return fmt.Sprintf("%v", d.Tags)
}

type Tag interface{}

type TagType interface {
	byte | string | uint16 | uint32 | Rational | int8 | int16 | int32 | SignedRational | float32 | float64
}

type DataTag[T TagType] struct {
	TagID  uint16
	Values []T
}

func (t DataTag[T]) String() string {
	if len(t.Values) > 9 {
		return fmt.Sprintf("%s: %d", TagsIDsLabels[t.TagID], len(t.Values))
	}
	return fmt.Sprintf("%s: %v", TagsIDsLabels[t.TagID], t.Values)
}

type Rational struct {
	Numerator   uint32
	Denominator uint32
}

func (r Rational) String() string {
	return fmt.Sprintf("%d/%d", r.Numerator, r.Denominator)
}

type SignedRational struct {
	Numerator   int32
	Denominator int32
}

func (r SignedRational) String() string {
	return fmt.Sprintf("%d/%d", r.Numerator, r.Denominator)
}
