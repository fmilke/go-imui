package gl

import (
	"fmt"
	"image"
	"os"
	"unicode"

	"github.com/benoitkugler/textlayout/fonts"
	"github.com/benoitkugler/textlayout/fonts/truetype"
	"github.com/benoitkugler/textlayout/harfbuzz"
	"github.com/benoitkugler/textlayout/language"
	"github.com/danielgatis/go-freetype/freetype"
	"github.com/go-gl/gl/v4.1-core/gl"

	. "dyiui/internal/types"
	. "dyiui/internal/units"
)

const DEBUG_GLYPH_PLACEMENT = false

const DEB_UV = 2
const DEB_POS = 1
const DEBUG_GLYPH_COMPONENTS = 0
const LOG_FONT = false

const WIN_WIDTH = 640
const WIN_HEIGHT = 480

func InsertGlyph(
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
	y := (base_line - char_hbear_y + yOffset) * px_y

	w := char_width * px_x
	h := char_height * px_y

	pos := Quad{
		X: x,
		Y: y,
		W: w,
		H: h,
	}

	if DEBUG_GLYPH_PLACEMENT {
		fmt.Printf("Inserting Quad: offset: %d, i: %d\n", offset, i)
	}

	if DEBUG_GLYPH_COMPONENTS > 0 {
		if DEBUG_GLYPH_COMPONENTS&DEB_UV > 0 {
			fmt.Printf("Location In GlpyhTex: %v\n", uvs)
		}

		fmt.Println()
	}

	insertGlyphComponents(
		verts,
		i+offset,
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

	pt := int(PxToPt(px))

	if LOG_FONT {
		fmt.Printf("Retrieving font face of size %vpt\n", pt)
	}

	err = face.Pt(pt, int(PIXELS_PER_LOGICAL_INCH))
	if err != nil {
		return nil, err
	}

	return face, nil
}

func GetGlyphBitmap(gid fonts.GID, a *freetype.Face) (*image.RGBA, *freetype.Metrics, error) {
    return a.GlyphByGid(int(gid))
}

func InitTex(rn fonts.GID, face *freetype.Face) (*image.RGBA, *freetype.Metrics) {

	img, metrics, err := face.Glyph(rune(rn))
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
	XOffset  float32
	YOffset  float32
    GID      fonts.GID
}

type Segment struct {
	Glyphs []Glyph
	Width  float32
}

/*
Split text into array of words,
while ignoring whitespace
*/
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

	buf.AddRunes(rs, 0, -1)
	buf.Props.Direction = harfbuzz.LeftToRight
	buf.Props.Language = language.DefaultLanguage()
	buf.Props.Script = language.Latin
	buf.Shape(hbFont, []harfbuzz.Feature{})
	metric := GetDefaultMetric()
	factor := float32(FontScaleFactor(ttf, metric, Sp(fontSize)))

	var segment Segment

	for i, g := range buf.Pos {

		segment.Glyphs = append(segment.Glyphs, Glyph{
			XAdvance: float32(g.XAdvance) * factor,
			YAdvance: float32(g.YAdvance) * factor,
			XOffset:  float32(g.XOffset),
			YOffset:  float32(g.YOffset),
			GID:      buf.Info[i].Glyph,
		})

		segment.Width += float32(g.XAdvance) * factor
	}


	return segment
}

const VERTS_PER_GLYPH = 6
const COMPS_PER_VERT = 4
const COMPS_PER_GLYPH = VERTS_PER_GLYPH * COMPS_PER_VERT

func CopyGlyphDataIntoVertexBuffer(
	placement *PlacedSegment,
	fontFace *freetype.Face,
	atlas *Atlas,
	offset int,
	vertices []float32,
) int {

	segment := placement.Segment
	xadv := placement.XOffset
	coi := 0

	for _, g := range segment.Glyphs {

		// Upload rasterized image
		rasterized, metrics, err := GetGlyphBitmap(g.GID, fontFace)
        // TODO: use fallback?
        if err != nil {
            panic(err)
        }

		// TODO: Don't upload, if still in texture
		q, cached := atlas.GetSlot(g.GID)
		if !cached {
			atlas.GlyphView.IntoCell(atlas.GlyphView.tex, rasterized, q)
		}

		widthRatio := float32(metrics.Width) / 32.0
		heightRatio := float32(metrics.Height) / 32.0

		// TODO: Remove mapping to uv-space. Should be done within caching data structure
		uvs := Quad{
			X: q.X / 1024,
			Y: q.Y / 1024,
			W: q.W / 1024 * float32(widthRatio),
			H: q.H / 1024 * float32(heightRatio),
		}

		InsertGlyph(xadv, placement.YOffset, coi, &vertices, metrics, offset, uvs)

		if DEBUG_GLYPH_PLACEMENT {
			fmt.Printf("Rune: %v, xadv: %v\n", g.GID, xadv)
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

func PlaceSegments(
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

	var currentWidth float32 = -whiteSpacesWidth

	// TODO: Handle case where text is empty and we have no line at all
	var totalHeight = lineHeight
	var totalWidth float32
	var placedSegs []PlacedSegment
	var yOffset float32 = 0
	var xOffset float32 = 0

	for _, seg := range segs {

		run := CalculateSegment(ttf, seg, hbFont, 32)

		breakLine := currentWidth+run.Width+whiteSpacesWidth > allowedWidth

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

		placed := PlacedSegment{
			Segment: run,
			XOffset: xOffset,
			YOffset: yOffset,
		}

		indicesToRender += len(run.Glyphs)
		placedSegs = append(placedSegs, placed)
	}

	return RenderTextResult{
		Width:          totalWidth,
		Height:         totalHeight,
		Indices:        indicesToRender,
		PlacedSegments: placedSegs,
	}
}

type PlacedSegment struct {
	Segment Segment
	XOffset float32
	YOffset float32
}

type RenderTextResult struct {
	Height         float32
	Width          float32
	Indices        int
	PlacedSegments []PlacedSegment
}

func (renderer *Renderer) UploadTextVertices(vertices []float32) {
}

// TODO: caching can be removed, once glyph atlas is working
// but still need to figure out, what happening, when too many glyphs
var indicesToRenderCached int = 0
var verticesCached []float32

type RenderTextArgs struct {
	GlyphView *GlyphView
	GlyphTex  *GlyphTexture
	FontFace  *freetype.Face
}

func (renderer *Renderer) RenderText(placement RenderTextResult, args *RenderTextArgs, pos Quad) {
	indicesToRender := 0
	offset := 0

	verticesCached = make([]float32, placement.Indices*COMPS_PER_GLYPH)

	for _, p := range placement.PlacedSegments {
		indicesToRender += CopyGlyphDataIntoVertexBuffer(&p, args.FontFace, renderer.GetAtlas(), offset, verticesCached)
		offset += len(p.Segment.Glyphs) * COMPS_PER_GLYPH
	}

	makeSegmentVaos(verticesCached)

	x := renderer.ToClipSpaceX(pos.X)
	y := renderer.ToClipSpaceY(pos.Y)

	gl.UseProgram(renderer.shaders.TextShader.Program)

	// TODO: use some method to retrieve texture atlas
	atlas := renderer.GetAtlas()
	if atlas != nil {
		gl.BindTexture(atlas.GlyphTexture.target, atlas.GlyphTexture.handle)
	}

	// set offset
	gl.Uniform2f(
		renderer.shaders.TextShader.Ul_Offset,
		x,
		y,
	)

	gl.Uniform1f(renderer.shaders.TextShader.Ul_Wireframe, .0) // read from context
	gl.Uniform3f(renderer.shaders.TextShader.Ul_TextColor, .3, .3, .3)

	gl.ActiveTexture(gl.TEXTURE0)
	gl.DrawArrays(gl.TRIANGLES, 0, int32(indicesToRender))
}
