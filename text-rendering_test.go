package main

import (
	"math"
	"testing"

	"github.com/benoitkugler/textlayout/fonts/truetype"
	"github.com/benoitkugler/textlayout/harfbuzz"
)

/*
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
*/

/*
func TestStuff(t *testing.T) {
	text := "Sometext"
	segment := GetSegment(text)

	if len(segment.Glyphs) != len(text) {
		t.Fatalf("Calculated segment of %s should have same length", text)
	}
}
*/
/*
func TestRenderSegment(t *testing.T) {

	text := "Sometext"
	segment := GetSegment(text)

}
*/


type TestFont struct {
	path string
	ttf *truetype.Font
	hbFont *harfbuzz.Font
}

func LoadTestFont () TestFont {
	path := GetSomeFont()
	ttf, err := LoadTTF(path)

	if err != nil {
		panic(err)
	}

	hbFont := HBFont(ttf)

	return TestFont {
		path,
		ttf,
		hbFont,
	}
}

func StringsShouldEqual(t *testing.T, a string, b string) {
	if a != b {
		t.Fatalf("Values should equal but don't: a: '%s' b: '%s'", a, b)
	}
}


func TestSegmentation1(t *testing.T) {
	ws := []string {
		"   ",
		"\t   ",
		"\n\t\r",
		"\r  \t",
	}

	for i, s := range ws {
		segs := SplitIntoSegments(s)
		
		if len(segs) > 0 {
			t.Fatalf("Segmentation of whitespaces should not return any segments. %d-th fixture failed", i)
		}
	}
}

func TestSegmentation2(t *testing.T) {
	s := "   This should be four"

	segs := SplitIntoSegments(s)

	StringsShouldEqual(t, "This", segs[0])
	StringsShouldEqual(t, "should", segs[1])
	StringsShouldEqual(t, "be", segs[2])
	StringsShouldEqual(t, "four", segs[3])
}

func TestNaiveLineBreaking(t *testing.T) {

	text := "Break inbetween"
	f := LoadTestFont()
	face, err := GetFace(f.path, 32.0)
	if err != nil {
		panic(err)
	}

	lineHeight := float32(32.0)

	placement := PlaceSegments(
		text,
		f.ttf,
		f.hbFont,
		face,
		1.0,
		lineHeight,
	)

	if len(placement.PlacedSegments) != 2 {
		t.Fatalf("Expected to have placed two segments from '%s'\n", text)
	}

	for i := range [2]int{} {
		xOffset := placement.PlacedSegments[i].XOffset
		if xOffset != 0 {
			t.Fatalf("Expected to be have both lines start at 0. %d-nth starts at: %f\n", i, xOffset)
		}
	}

	a := float64(placement.Height)
	b := float64(lineHeight * 2)
	if math.Abs(a) - math.Abs(b) < .001 {
		t.Fatalf("Expected height to be two line heights. Was %f. Expected %f\n", a, b)
	}
}


