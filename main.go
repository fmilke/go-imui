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


	indices := a.CalculateSegments("This is text")

	for !a.window.ShouldClose() {
		draw(1, a.window, a.program, a.glyphTex, int32(indices))
	}
}

func (app App) CalculateSegments(s string) int {

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
			fmt.Printf("Skipping space")
			continue
		}
		rasterized, metrics := initTex(r, app.fontFace)

		fmt.Printf("Rune: '%v': %v,%v\n", string(r), metrics.Width, metrics.Height)
		app.glyphView.IntoCell(app.glyphTex, rasterized, cellSlotX, 0)

		appendRune(xadv, coi, &co, metrics)
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
