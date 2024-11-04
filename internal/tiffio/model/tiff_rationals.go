package model

import "fmt"

// --------------------------
// specialized structs
// --------------------------

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
