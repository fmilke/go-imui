package main

import (
	"fmt"
	"math"
	"os"
	"runtime"
	"strings"

	"github.com/benoitkugler/textlayout/fonts/truetype"
	"github.com/benoitkugler/textlayout/harfbuzz"
	"github.com/danielgatis/go-findfont/findfont"
	"github.com/danielgatis/go-freetype/freetype"
	"github.com/go-gl/glfw/v3.3/glfw"
)

const WIN_WIDTH = 640
const WIN_HEIGHT = 480
const WIN_NAME = "Testing"

const LOG_FONT = true

func init() {
	runtime.LockOSThread()
}

func main() {
	testConversions()

	err := glfw.Init()

	if err != nil {
		panic(err)
	}

	defer glfw.Terminate()

	getBuffer()

	app := createApp()
	app.Init(WIN_WIDTH, WIN_HEIGHT, WIN_NAME)

	path := GetSomeFont()
	ttf, err := LoadTTF(path)
	hbFont := HBFont(ttf)

	app.Loop(ttf, hbFont)
}

type App struct {
	window  *glfw.Window
	program uint32

	fontFace *freetype.Face

	glyphTex  *GlyphTexture
	glyphView GlyphView
}

func createApp() *App {
	return &App{}
}

func (a *App) Init(width int, height int, name string) {

	window, err := glfw.CreateWindow(width, height, name, nil, nil)

	if err != nil {
		panic(err)
	}

	window.MakeContextCurrent()

	a.window = window

	a.program = initOpenGL()
	CheckGLErrorsPrint("Pre tex")

	glTex := newGlyphTexture(1024)
	face := initFace(32)

	view := GlyphView{
		size: 32,
		tex:  glTex,
	}

	a.glyphTex = glTex
	a.glyphView = view
	a.fontFace = face
}

func (a *App) Loop(
	ttf *truetype.Font,
	hbFont *harfbuzz.Font,
) {

	fmt.Println("======")
	indices := renderText(
		"Someetext",
		ttf,
		hbFont,
		a.fontFace,
		a.glyphView,
		a.glyphTex,
	)
	for !a.window.ShouldClose() {
		draw(1, a.window, a.program, a.glyphTex, int32(indices))
	}
}

func (a *App) Loop2() {

	indices := CalculateSegments(
		a.fontFace,
		a.glyphView,
		a.glyphTex,
//		"This is text",
"a",
	)
	for !a.window.ShouldClose() {
		draw(1, a.window, a.program, a.glyphTex, int32(indices))
	}
}

func HBFont(font *truetype.Font) *harfbuzz.Font {
	return harfbuzz.NewFont(font)
}

type GlyphInfo struct {
	XOffset int
	YOffset int
	XAdvance int
}

func getBuffer() {

	fs, err := findfont.Find("Arial", findfont.FontRegular)

	if err != nil {
		panic(err)
	}

	f := fs[0][2]

	file, err := os.Open(f)

	if err != nil {
		panic(err)
	}

	font, err := truetype.Parse(file)

	if err != nil {
		panic(err)
	}

	buf := harfbuzz.NewBuffer()
	s := "This is harfbuzz"
	text := []rune(s)
	buf.AddRunes(text, 0, len(text))

	hbFont := HBFont(font)
	buf.Shape(hbFont, []harfbuzz.Feature{})

	for _, g := range buf.Pos {
		fmt.Printf("glyph: %v\n", g)
	}
}

func GetSomeFont() string {
	fs, err := findfont.Find("Arial", findfont.FontRegular)

	if err != nil {
		panic(err)
	}

	path := fs[0][2]
	return path
}

func getHbFont() *harfbuzz.Font {
	fs, err := findfont.Find("Arial", findfont.FontRegular)

	if err != nil {
		panic(err)
	}

	f := fs[0][2]

	file, err := os.Open(f)
	
	if err != nil {
		panic(err)
	}

	font, err := truetype.Parse(file)

	if err != nil {
		panic(err)
	}

	hbFont := HBFont(font)

	return hbFont
}

func renderText(
	text string,
	ttf *truetype.Font,
	hbFont *harfbuzz.Font,
	fontFace *freetype.Face,
	glyphView GlyphView,
	glyphTex *GlyphTexture,
) int {
	// TODO: Check if pointer into string based
	// solution is more efficient

	indicesToRender := 0

	segments :=	strings.Fields(text)

	for _, s := range segments {
		//indicesToRender += renderSegment(fontFace, s, hbFont, glyphView, glyphTex)
		seg := CalculateSegment(ttf, s, hbFont)
		indicesToRender += RenderSegment(&seg, fontFace, glyphView, glyphTex)
	}

	return indicesToRender
}


func testConversions() {

	r := math.Abs(float64(pxToPt(16)) - (12))
	if r > 0.0001 {
		fmt.Printf("pxToPt is off\n: %v", r)
	}

	r2 := math.Abs(float64(ptToPx(12)) - (16))
	if r2 > 0.0001 {
		fmt.Printf("ptToPx is off: %v\n", r2)
	}
}
