package main

import (
	"math"
	"testing"
    . "dyiui/internal/units"
)

func TestConversions(t *testing.T) {

	r := math.Abs(float64(PxToPt(16)) - (12))
	if r > 0.0001 {
		t.Fatalf("pxToPt is off: %v\n", r)
	}

	r2 := math.Abs(float64(PtToPx(12)) - (16))
	if r2 > 0.0001 {
		t.Fatalf("ptToPx is off: %v\n", r2)
	}
}
