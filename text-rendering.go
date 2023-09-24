package main

import (
	"fmt"
	"os"
	"image"
	"io/ioutil"

	"github.com/danielgatis/go-findfont/findfont"
	"github.com/benoitkugler/textlayout/harfbuzz"
	"github.com/benoitkugler/textlayout/fonts/truetype"
	"github.com/danielgatis/go-freetype/freetype"
)


const PT_PER_LOGICAL_INCH = 72.0
const PIXELS_PER_LOGICAL_INCH = 96.0 // aka. DPI


type GlyphTexture struct {
	handle  uint32
	target  uint32
	width   int32
	height  int32
}

type GlyphView struct {
	size int32
	tex  *GlyphTexture
}

func NewGlyphTexture() {
}

func NewGlyphView() {
}

func appendRune(
	xAdv float32,
	i int,
	verts *[]float32,
	metrics *freetype.Metrics,
) {
	px_x := 1.0 / float32(WIN_WIDTH)
	px_y := 1.0 / float32(WIN_HEIGHT)

	char_width := float32(metrics.Width)
	char_height := float32(metrics.Height)

	char_hbear_y := float32(metrics.HorizontalBearingY)

	fixed_offset := float32(30.0)

	x := xAdv * px_x
	y := (fixed_offset-char_hbear_y) * px_y

	w := char_width * px_x
	h := char_height * px_y

	fmt.Printf("x: %v,y: %v, w: %v, h: %v\n", x, y, w, h)

	cellWidth := float32(1024 / 32)
	cellHeight := float32(1024 / 32)

	widthRatio := float32(metrics.Width) / cellWidth
	heightRatio := float32(metrics.Height) / cellHeight

	uOffset := float32(i)/24.0*float32(32.0/1024.0)
	vOffset := float32(0.0)
	uWidth := 1.0/cellWidth * widthRatio
	vWidth := 1.0/cellHeight * heightRatio

	fmt.Printf("segment data: uv: ux: %v, uy %v, uw: %v, uh: %v\n", uOffset, vOffset, uWidth, vWidth)

	insertQuad(
		verts,
		i,
		x,
		y,
		w,
		h,
		uOffset,
		vOffset,
		uWidth,
		vWidth,
	)
}

func insertQuad(
	verts *[]float32,
	i int,
	x float32,
	y float32,
	w float32,
	h float32,
	u float32,
	v float32,
	uw float32,
	vh float32,
) {

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

func initFace(px float32) *freetype.Face {
	fonts, err := findfont.Find("Arial", findfont.FontRegular)

	if err != nil {
		panic(err)
	}

	data, err := ioutil.ReadFile(fonts[0][2])

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


func CalculateSegments(
	fontFace *freetype.Face,
	glyphView GlyphView,
	glyphTex *GlyphTexture,
	s string,
) int {

	verticesPerRune := 6
	componentPerVertex := 4
	runeStride := componentPerVertex * verticesPerRune
	co := make([]float32, len(s)*runeStride)

	xadv := 0.0
	coi := 0

	indices := 0
	cellSlotX := int32(0)
	for _, r := range s {

		if r == ' ' {
			xadv += 24.0
			fmt.Printf("Skipping space\n")
			continue
		}
		rasterized, metrics := initTex(r, fontFace)

		fmt.Printf("Rune: '%v': %v,%v\n", string(r), metrics.Width, metrics.Height)
		glyphView.IntoCell(glyphTex, rasterized, cellSlotX, 0)

		appendRune(float32(xadv), coi, &co, metrics)
		coi += runeStride

		xadv += float64(metrics.Width)
		indices += verticesPerRune
		cellSlotX++
	}

	fmt.Printf("co: %+v\n", co)

	makeSegmentVaos(co)
	CheckGLErrors()

	return indices
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

	fmt.Println("Font Scale Factor: ", factor)
	var segment Segment

	for i, g := range buf.Pos {
		fmt.Printf("Calculating for glyph: %+v \n", g)

		segment.Glyphs = append(segment.Glyphs, Glyph{
			XAdvance: float32(g.XAdvance) * factor,
			YAdvance: float32(g.YAdvance) * factor,
			XOffset: float32(g.XOffset),
			YOffset: float32(g.YOffset),
			R: rs[i],
		})

		segment.Width += float32(g.XAdvance)
	}

	return segment
}

const VERTS_PER_GLYPH = 6
const COMPS_PER_VERT = 4
const COMPS_PER_GLYPH = VERTS_PER_GLYPH * COMPS_PER_VERT

func RenderSegment(
	segment *Segment,
	fontFace *freetype.Face,
	glyphView GlyphView,
	glyphTex *GlyphTexture,
) int {
	glyphCount := len(segment.Glyphs)
	vertices := make([]float32, glyphCount*COMPS_PER_GLYPH)

	cellSlotX := int32(0)
	xadv := float32(0.0)
	coi := 0

	for _, g := range segment.Glyphs {

		if g.R == ' ' {
			xadv += 24.0
			continue
		}

		// Copy rasterized image
		rasterized, metrics := initTex(g.R, fontFace)

		// Copy into texture
		glyphView.IntoCell(glyphTex, rasterized, cellSlotX, 0)

		// Write vertex buffer
		appendRune(xadv, coi, &vertices, metrics)

		fmt.Printf("xadv: %v\n", xadv)

		xadv += g.XAdvance
		coi += COMPS_PER_GLYPH

		cellSlotX++
	}
 
	makeSegmentVaos(vertices)
	CheckGLErrorsPrint("RenderSegment: makeSegmentVaos")

	return glyphCount * VERTS_PER_GLYPH
}

func renderSegment(
	fontFace *freetype.Face,
	segment string,
	hbFont *harfbuzz.Font,
	glyphView GlyphView,
	glyphTex *GlyphTexture,
) int {
	verticesPerRune := 6
	componentPerVertex := 4
	runeStride := componentPerVertex * verticesPerRune
	indicesToRender := 0
	// TODO: Reuse buffer?
	
	buf := harfbuzz.NewBuffer()
	text := []rune(segment)
	buf.AddRunes(text, 0, len(text))
	cellSlotX := int32(0)
	xadv := 0.0

	co := make([]float32, len(segment)*runeStride)
	buf.Shape(hbFont, []harfbuzz.Feature{})

	//metric := GetDefaultMetric()
	//factor := FontScaleFactor(ttf, metric, 14)

	for i, g := range buf.Pos {
		fmt.Printf("glyph: %v\n", g)
		// TODO: Check for faster way to access rune
		// TODO: Also this access does not respect unicode
		r := rune(segment[i])
		
		rasterized, _ := initTex(r, fontFace)
		glyphView.IntoCell(glyphTex, rasterized, cellSlotX, 0)

		xadv += float64(g.XAdvance)
		indicesToRender += verticesPerRune
		cellSlotX++
	}

	makeSegmentVaos(co)
	CheckGLErrors()

	return indicesToRender
}

func FontScaleFactor(font *truetype.Font, m Metric, size Sp) float32 {
	sizePx := m.Sp(size)

	upem := font.Upem()
	factor := float32(sizePx) / float32(upem)
	return factor
}
