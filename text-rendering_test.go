package main

import (
	"testing"
)

func GetSegment(text string) Segment {

	path := GetSomeFont()
	ttf, err := LoadTTF(path)

	if err != nil {
		panic(err)
	}

	hbFont := HBFont(ttf)
	
	segment := CalculateSegment(
		 ttf,
		 text,
		 hbFont,
	)

	return segment
}

func TestStuff(t *testing.T) {
	text := "Sometext"
	segment := GetSegment(text)

	if len(segment.Glyphs) != len(text) {
		t.Fatalf("Calculated segment of %s should have same length", text)
	}
}
/*
func TestRenderSegment(t *testing.T) {

	text := "Sometext"
	segment := GetSegment(text)

}
*/
