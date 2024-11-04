package model

import (
	"TiffReader/internal/tiffio/tags"
	"fmt"
)

// --------------------------
// TIFF tags storing metadata
// --------------------------

type TIFFTag interface {
	GetTagID() tags.TagID
	ValuesCount() int
	AsBytes() []byte
	AsStrings() []string
	AsUint16s() []uint16
	AsUint32s() []uint32
	AsUint64s() []uint64
	AsRationals() []Rational
	AsSignedRationals() []SignedRational
	AsInt8s() []int8
	AsInt16s() []int16
	AsInt32s() []int32
	AsFloat32s() []float32
	AsFloat64s() []float64
	GetUintVal(int) uint64
}

type TagType interface {
	byte | string | uint16 | uint32 | uint64 | Rational | int8 | int16 | int32 | int64 | SignedRational | float32 | float64
}

// --------------------------
// DataTag implements Tag
// --------------------------

type DataTag[T TagType] struct {
	TagID  tags.TagID
	Values []T
}

func (t DataTag[T]) GetTagID() tags.TagID {
	return t.TagID
}

func (t DataTag[T]) ValuesCount() int {
	return len(t.Values)
}

func (t DataTag[T]) AsBytes() []byte {
	if byteTag, ok := any(t).(DataTag[byte]); ok {
		return byteTag.Values
	}
	return nil
}

func (t DataTag[T]) AsStrings() []string {
	if stringTag, ok := any(t).(DataTag[string]); ok {
		return stringTag.Values
	}
	return nil
}

func (t DataTag[T]) AsUint16s() []uint16 {
	if uint16Tag, ok := any(t).(DataTag[uint16]); ok {
		return uint16Tag.Values
	}
	return nil
}

func (t DataTag[T]) AsUint32s() []uint32 {
	if uint32Tag, ok := any(t).(DataTag[uint32]); ok {
		return uint32Tag.Values
	}
	return nil
}

func (t DataTag[T]) AsUint64s() []uint64 {
	if uint64Tag, ok := any(t).(DataTag[uint64]); ok {
		return uint64Tag.Values
	}
	return nil
}

func (t DataTag[T]) AsRationals() []Rational {
	if rationalTag, ok := any(t).(DataTag[Rational]); ok {
		return rationalTag.Values
	}
	return nil
}

func (t DataTag[T]) AsSignedRationals() []SignedRational {
	if rationalTag, ok := any(t).(DataTag[SignedRational]); ok {
		return rationalTag.Values
	}
	return nil
}

func (t DataTag[T]) AsInt8s() []int8 {
	if int8Tag, ok := any(t).(DataTag[int8]); ok {
		return int8Tag.Values
	}
	return nil
}

func (t DataTag[T]) AsInt16s() []int16 {
	if int16Tag, ok := any(t).(DataTag[int16]); ok {
		return int16Tag.Values
	}
	return nil
}

func (t DataTag[T]) AsInt32s() []int32 {
	if int32Tag, ok := any(t).(DataTag[int32]); ok {
		return int32Tag.Values
	}
	return nil
}

func (t DataTag[T]) AsFloat32s() []float32 {
	if float32Tag, ok := any(t).(DataTag[float32]); ok {
		return float32Tag.Values
	}
	return nil
}

func (t DataTag[T]) AsFloat64s() []float64 {
	if float64Tag, ok := any(t).(DataTag[float64]); ok {
		return float64Tag.Values
	}
	return nil
}

func (t DataTag[T]) GetUintVal(idx int) uint64 {
	if idx >= len(t.Values) {
		return 0
	}
	switch any(t).(type) {
	case DataTag[uint64]:
		return any(t).(DataTag[uint64]).Values[idx]
	case DataTag[uint32]:
		return uint64(any(t).(DataTag[uint32]).Values[idx])
	case DataTag[uint16]:
		return uint64(any(t).(DataTag[uint16]).Values[idx])
	case DataTag[uint8]:
		return uint64(any(t).(DataTag[uint8]).Values[idx])
	default:
		return 0
	}
}

func (t DataTag[T]) String() string {
	if len(t.Values) > 9 {
		return fmt.Sprintf("%s: %d", tags.IDsLabels[t.TagID], len(t.Values))
	}
	return fmt.Sprintf("%s: %v", tags.IDsLabels[t.TagID], t.Values)
}
