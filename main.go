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
const USE_DEBUG_UV = false;

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

	path := GetSomeFont()
	ttf, err := LoadTTF(path)

	if err != nil {
		panic(err)
	}
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

	glTex := newGlyphTexture(1024)
	
	if USE_DEBUG_UV {
		png, err := loadPng("./assets/checkered-uvs.png");
		if err == nil {
			replaceGlpyhTexture(glTex, png)
		} else {
			fmt.Println("Could not load test uv", err)
		}
	}

	face := initFace(32)

	view := GlyphView{
		size: 32,
		tex:  glTex,
	}

	a.glyphTex = glTex
	a.glyphView = view
	a.fontFace = face
}


var done = false;

func (a *App) Loop(
	ttf *truetype.Font,
	hbFont *harfbuzz.Font,
) {

	fmt.Println("======")
	indices := RenderText(
		"Should be on the same but break somewhere Firstverylongline2",
		ttf,
		hbFont,
		a.fontFace,
		a.glyphView,
		a.glyphTex,
	)

	for !a.window.ShouldClose() {
		draw(1, a.window, a.program, a.glyphTex, int32(indices))

		if !done {
			img := readGlyphTexture(a.glyphTex)
			err := writePng("./test.png", img)
			if err != nil {
				fmt.Println(err)
			}
			done = true;
		}
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
