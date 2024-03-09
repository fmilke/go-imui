package main

import (
	"fmt"
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

const LOG_FONT = false
const USE_DEBUG_UV = false
const DEBUG_SHADERS = true
const DEBUG = true

func init() {
	runtime.LockOSThread()
}

func main() {

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
	window  *glfw.Window
	context *Context

	fontFace *freetype.Face
    hbFont *harfbuzz.Font
    ttf *truetype.Font

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
	a.context = initOpenGL()

	glTex := newGlyphTexture(1024)
	
	if USE_DEBUG_UV {
		png, err := loadPng("./assets/checkered-uvs.png");
		if err == nil {
			replaceGlyphTexture(glTex, png)
		} else {
			fmt.Println("Could not load test uv", err)
		}
	}

	face := initFace(32)

	view := GlyphView{
		size: 32,
		tex:  glTex,
	}

    // setup harfbuzz
	path := GetSomeFont()
	ttf, err := LoadTTF(path)
	if err != nil {
		panic(err)
	}

	hbFont := HBFont(ttf)

	a.glyphTex = glTex
	a.glyphView = view
	a.fontFace = face
    a.hbFont = hbFont
    a.ttf = ttf
}

var done = false;

func (a *App) Loop() {

	for !a.window.ShouldClose() {

        BeginFrame()

		DrawFrame(a, a.context)

        // Handle events
        glfw.PollEvents()
        // Push to display
        a.window.SwapBuffers()
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

func GetSomeFont() string {
	fs, err := findfont.Find("Noto Sans", findfont.FontRegular)

	if err != nil {
		panic(err)
	}

	path := fs[0][2]
	return path
}
