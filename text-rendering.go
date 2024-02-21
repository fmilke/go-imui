package main

import (
	"fmt"
	"image"
	"os"
	"unicode"

	"github.com/benoitkugler/textlayout/fonts/truetype"
	"github.com/benoitkugler/textlayout/harfbuzz"
	"github.com/danielgatis/go-findfont/findfont"
	"github.com/danielgatis/go-freetype/freetype"
)

const PT_PER_LOGICAL_INCH = 72.0
const PIXELS_PER_LOGICAL_INCH = 96.0 // aka. DPI

const DEBUG_GLYPH_PLACEMENT = true;

const DEB_UV = 2;
const DEB_POS = 1;
const DEBUG_GLYPH_COMPONENTS = 0;

type Quad struct {
	X float32
	Y float32
	W float32
	H float32
}
func insertGlyph(
	xAdv float32,
	yOffset float32,
	i int,
	verts *[]float32,
	metrics *freetype.Metrics,
	offset int,
	uvs Quad,
) {
	px_x := 1.0 / float32(WIN_WIDTH)
	px_y := 1.0 / float32(WIN_HEIGHT)

	char_width := float32(metrics.Width)
	char_height := float32(metrics.Height)

	char_hbear_y := float32(metrics.HorizontalBearingY)

	// TODO: properly calculate baseline
	base_line := float32(30.0)

	x := xAdv * px_x
	y := (base_line-char_hbear_y + yOffset) * px_y

	w := char_width * px_x
	h := char_height * px_y

	pos := Quad {
		X: x,
		Y: y,
		W: w,
		H: h,
	}

	if DEBUG_GLYPH_PLACEMENT {
		fmt.Printf("Inserting Quad: offset: %d, i: %d\n", offset, i)
	}

	if DEBUG_GLYPH_COMPONENTS > 0 {
		if DEBUG_GLYPH_COMPONENTS & DEB_UV > 0 {
			fmt.Printf("Location In GlpyhTex: %v\n", uvs)
		}

		fmt.Println()
	}

	insertGlyphComponents(
		verts,
		i + offset,
		pos,
		uvs,
	)
}

func insertGlyphComponents(
	verts *[]float32,
	i int,
	pos Quad,
	loc Quad,
) {

	u := loc.X
	v := loc.Y
	uw := loc.W
	vh := loc.H

	x := pos.X
	y := pos.Y
	w := pos.W
	h := pos.H

	u_min := u
	v_min := v

	u_max := u + uw
	v_max := v + vh

	(*verts)[i] = x
	(*verts)[i+1] = y

	(*verts)[i+2] = u_min
	(*verts)[i+3] = v_min

	(*verts)[i+4] = x + w
	(*verts)[i+5] = y

	(*verts)[i+6] = u_max
	(*verts)[i+7] = v_min

	(*verts)[i+8] = x
	(*verts)[i+9] = y + h

	(*verts)[i+10] = u_min
	(*verts)[i+11] = v_max

	//

	(*verts)[i+12] = x
	(*verts)[i+13] = y + h

	(*verts)[i+14] = u_min
	(*verts)[i+15] = v_max

	(*verts)[i+16] = x + w
	(*verts)[i+17] = y

	(*verts)[i+18] = u_max
	(*verts)[i+19] = v_min

	(*verts)[i+20] = x + w
	(*verts)[i+21] = y + h

	(*verts)[i+22] = u_max
	(*verts)[i+23] = v_max
}

func GetFace(path string, px float32) (*freetype.Face, error) {
	data, err := os.ReadFile(path)

	if err != nil {
		return nil, err
	}

	lib, err := freetype.NewLibrary()
	if err != nil {
		return nil, err
	}

	face, err := freetype.NewFace(lib, data, 0)
	if err != nil {
		return nil, err
	}

	pt := int(pxToPt(px))

	if LOG_FONT {
		fmt.Printf("Retrieving font face of size %vpt\n", pt)
	}

	err = face.Pt(pt, int(PIXELS_PER_LOGICAL_INCH))
	if err != nil {
		return nil, err
	}

	return face, nil
}

func initFace(px float32) *freetype.Face {
	fonts, err := findfont.Find("Arial", findfont.FontRegular)

	if err != nil {
		panic(err)
	}

	data, err := os.ReadFile(fonts[0][2])

	if err != nil {
		panic(err)
	}

	lib, err := freetype.NewLibrary()
	if err != nil {
		panic(err)
	}

	face, err := freetype.NewFace(lib, data, 0)
	if err != nil {
		panic(err)
	}

	pt := int(pxToPt(px))

	if LOG_FONT {
		fmt.Printf("Retrieving font face of size %vpt\n", pt)
	}
	
	err = face.Pt(pt, int(PIXELS_PER_LOGICAL_INCH))
	if err != nil {
		panic(err)
	}

	return face
}

func ptToPx(pt float32) float32 {
	return pt * (PIXELS_PER_LOGICAL_INCH / PT_PER_LOGICAL_INCH)
}

func pxToPt(px float32) float32 {
	return px * (PT_PER_LOGICAL_INCH / PIXELS_PER_LOGICAL_INCH)
}

func initTex(rn rune, face *freetype.Face) (*image.RGBA, *freetype.Metrics) {

	img, metrics, err := face.Glyph(rn)
	if err != nil {
		panic(err)
	}

	// TODO: Add cleanup?
	// err = face.Done()
	// if err != nil {
	// 	panic(err)
	// }

	// err = lib.Done()
	// if err != nil {
	// 	panic(err)
	// }

	return img, metrics
}

type Glyph struct {
	XAdvance float32
	YAdvance float32
	XOffset float32
	YOffset float32
	R rune
}

type Segment struct {
	Glyphs []Glyph
	Width float32
}

func LoadTTF(path string) (*truetype.Font, error) {
	fs, err := findfont.Find("Arial", findfont.FontRegular)

	if err != nil {
		return nil, err
	}

	f := fs[0][2]

	file, err := os.Open(f)
	
	if err != nil {
		return nil, err
	}

	font, err := truetype.Parse(file)

	return font, err
}

func SplitIntoSegments(text string) []string {

	var segs []string

	withinWhitespaceRun := true
	start := 0

	for i, r := range text {
		if unicode.IsSpace(r) {
			if !withinWhitespaceRun {
				segs = append(segs, text[start:i])
				withinWhitespaceRun = true
			}
		} else if withinWhitespaceRun {
			withinWhitespaceRun = false
			start = i
		}
	}

	if !withinWhitespaceRun {
		segs = append(segs, text[start:])
	}

	return segs
}

func CalculateSegment(
	ttf *truetype.Font,
	text string,
	hbFont *harfbuzz.Font,
	fontSize int,
) Segment {
	buf := harfbuzz.NewBuffer()
	rs := []rune(text)

	buf.AddRunes(rs, 0, len(rs))
	buf.Props.Direction = harfbuzz.LeftToRight
	buf.Shape(hbFont, []harfbuzz.Feature{})

	metric := GetDefaultMetric()
	factor := float32(FontScaleFactor(ttf, metric, Sp(fontSize)))

	var segment Segment

	for i, g := range buf.Pos {
		segment.Glyphs = append(segment.Glyphs, Glyph{
			XAdvance: float32(g.XAdvance) * factor,
			YAdvance: float32(g.YAdvance) * factor,
			XOffset: float32(g.XOffset),
			YOffset: float32(g.YOffset),
			R: rs[i],
		})

		segment.Width += float32(g.XAdvance) * factor
	}

	return segment
}

const VERTS_PER_GLYPH = 6
const COMPS_PER_VERT = 4
const COMPS_PER_GLYPH = VERTS_PER_GLYPH * COMPS_PER_VERT

var cid = 0

func CopyGlyphDataIntoVertexBuffer(
	placement *PlacedSegment,
	fontFace *freetype.Face,
	glyphView GlyphView,
	glyphTex *GlyphTexture,
	offset int,
	vertices []float32,
) int {

	segment := placement.Segment
	xadv := placement.XOffset
	coi := 0

	for _, g := range segment.Glyphs {

		if g.R == ' ' {
			// TODO: Is this generalizable? What is the actual
			// error to catch here?
			fmt.Println("Whitespace found in segment. This should never happen")
			continue
		}

		// Copy rasterized image
		rasterized, metrics := initTex(g.R, fontFace)

		// Copy into texture
		x := int32(cid % 32)
		y := int32(cid / 32)
		glyphView.IntoCell(glyphTex, rasterized, x, y)

		// Write vertex buffer
		// TODO: Properly get uvs from texture placement
		cellWidth := float32(1024 / 32)
		cellHeight := float32(1024 / 32)
	
		widthRatio := float32(metrics.Width) / cellWidth
		heightRatio := float32(metrics.Height) / cellHeight

		cx := float32(int(float32(cid)) % 32.0)
		cy := float32(int(float32(cid)) / 32.0)
	
		uSize := 1.0/cellWidth * widthRatio
		vSize := 1.0/cellHeight * heightRatio
	
		uOffset := float32(cx)*float32(32.0/1024.0)
		vOffset := float32(cy)*float32(32.0/1024.0)

		uvs := Quad {
			X: uOffset,
			Y: vOffset,
			W: uSize,
			H: vSize,
		}

		fmt.Printf("uvs: %v\n", uvs)

		insertGlyph(xadv, placement.YOffset, coi, &vertices, metrics, offset, uvs)
		cid++

		if DEBUG_GLYPH_PLACEMENT {
			fmt.Printf("Rune: %v, xadv: %v\n", string(g.R), xadv)
		}

		xadv += g.XAdvance
		coi += COMPS_PER_GLYPH
	}

	return len(segment.Glyphs) * VERTS_PER_GLYPH
}

func FontScaleFactor(font *truetype.Font, m Metric, size Sp) float32 {
	sizePx := m.Sp(size)

	upem := font.Upem()
	factor := float32(sizePx) / float32(upem)
	return factor
}

func PlaceSegments (
	text string,
	ttf *truetype.Font,
	hbFont *harfbuzz.Font,
	fontFace *freetype.Face,
	allowedWidth float32,
	lineHeight float32,
) RenderTextResult {
	indicesToRender := 0
	whiteSpacesWidth := float32(32.0)

	segs := SplitIntoSegments(text)

	var currentWidth float32 = - whiteSpacesWidth

	// TODO: Handle case where text is empty and we have no line at all
	var totalHeight = lineHeight
	var totalWidth float32
	var placedSegs []PlacedSegment
	var yOffset float32 = lineHeight * 2.0
	var xOffset float32 = 0

	for _, seg := range segs {

		run := CalculateSegment(ttf, seg, hbFont, 32)

		breakLine := currentWidth + run.Width + whiteSpacesWidth > allowedWidth
		//fmt.Printf("Segs: '%s', CurrentWidth: %v, RunWidth: %v, WhiteSpacesWidth: %v, AllowedWidth: %v, Break line: %v\n", seg, currentWidth, run.Width, whiteSpacesWidth, allowedWidth, breakLine)

		if breakLine {
			currentWidth = run.Width
			totalHeight += lineHeight
			xOffset = 0
			totalWidth = max(run.Width, totalWidth)
			yOffset += lineHeight
		} else {
			xOffset = currentWidth + whiteSpacesWidth
			currentWidth = xOffset + run.Width
			totalWidth = max(totalWidth, currentWidth)
		}

		//fmt.Printf("xOffset: %f, yOffset: %f\n", xOffset, yOffset)

		placed := PlacedSegment {
			Segment: run,
			XOffset: xOffset,
			YOffset: yOffset,
		}

		indicesToRender += len(run.Glyphs)
		placedSegs = append(placedSegs, placed)
	}

	return RenderTextResult {
		Width: totalWidth,
		Height: totalHeight,
		Indices: indicesToRender,
		PlacedSegments: placedSegs,
	}
}

type PlacedSegment struct {
	Segment Segment
	XOffset float32
	YOffset float32
}

type RenderTextResult struct {
	Height float32
	Width float32
	Indices int
	PlacedSegments []PlacedSegment
}

func RenderText(
	text string,
	ttf *truetype.Font,
	hbFont *harfbuzz.Font,
	fontFace *freetype.Face,
	glyphView GlyphView,
	glyphTex *GlyphTexture,
) int {
	indicesToRender := 0
	offset := 0

	placement := PlaceSegments(text, ttf, hbFont, fontFace, 640.0, 32.0)
	vertices := make([]float32, placement.Indices*COMPS_PER_GLYPH)

	for _, p := range placement.PlacedSegments {
		indicesToRender += CopyGlyphDataIntoVertexBuffer(&p, fontFace, glyphView, glyphTex, offset, vertices)
		fmt.Printf("Offset in vao: %v\n", offset)
		offset += len(p.Segment.Glyphs) * COMPS_PER_GLYPH

		fmt.Println("-----")
	}

	makeSegmentVaos(vertices)
	CheckGLErrorsPrint("RenderSegment: makeSegmentVaos")

	return indicesToRender
}

