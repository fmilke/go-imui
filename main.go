package main

import (
	"runtime"
    . "dyiui/internal/gl"
    . "dyiui/internal/ui"
    . "dyiui/internal/layout"

	"github.com/go-gl/glfw/v3.3/glfw"
)

const WIN_NAME = "Testing"

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
    renderer *Renderer
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
	a.renderer = InitRenderer(WIN_WIDTH, WIN_HEIGHT)

    a.context = &Context{
        Width: WIN_WIDTH,
        Height: WIN_HEIGHT,
        PointerState: NewPointerState(),
    }
}

func (a *App) Loop() {
	for !a.window.ShouldClose() {

        a.UpdatePointerState()

        BeginFrame()
        ui := NewUI(a.context, a.renderer)

        RenderUI(&ui)

        FinishFrame()

        // Push to display
        a.window.SwapBuffers()

        // Handle events
        glfw.PollEvents()
	}
}

func (a *App) UpdatePointerState() {
    ev := a.window.GetMouseButton(glfw.MouseButtonLeft)

    s := &a.context.PointerState

    if ev == glfw.Press {
        s.JustActivated = !s.Active
        s.Active = true
        s.JustReleased = false
    } else if ev == glfw.Repeat {
        s.Active = true
        s.JustReleased = false
        s.JustActivated = false
    } else {
        s.JustReleased = s.Active
        s.Active = false
        s.JustActivated = false
    }

    x, y := a.window.GetCursorPos()

    s.PosX = float32(x)
    s.PosY = float32(y)
}

type GlyphInfo struct {
	XOffset int
	YOffset int
	XAdvance int
}


