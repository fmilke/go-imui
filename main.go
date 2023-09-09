package main

import (
	"fmt"
	"math"
	"os"
	"runtime"

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
	app.Loop()
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

func (a *App) Loop() {

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
