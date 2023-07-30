package main

import (
	"fmt"
	"math"
	"runtime"

	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/danielgatis/go-freetype/freetype"
)


const WIN_WIDTH = 640
const WIN_HEIGHT = 480
const WIN_NAME = "Testing"

const LOG_FONT = true;

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

	app := createApp()
	app.Init(WIN_WIDTH, WIN_HEIGHT, WIN_NAME)
	app.Loop()
}

type App struct {
	window *glfw.Window
	program uint32

	fontFace *freetype.Face 

	glyphTex *GlyphTexture
	glyphView GlyphView
}

func createApp() (*App) {
	return &App {}
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
		"This is text",
	)

	for !a.window.ShouldClose() {
		draw(1, a.window, a.program, a.glyphTex, int32(indices))
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
