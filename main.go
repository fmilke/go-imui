package main

import (
	"fmt"
	"image"
	"io/ioutil"
	"log"
	"runtime"
	"strings"
	"unsafe"

	"github.com/danielgatis/go-findfont/findfont"
	"github.com/danielgatis/go-freetype/freetype"
	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
)

const (
	vertexShaderSource = `#version 410
    in vec3 vp;
	out vec2 uv;
    void main() {
        gl_Position = vec4(vp, 1.0);
		uv = vec2(vp.x + 1.0, 1.0 - vp.y) / 2.0;
    }
` + "\x00"

	fragmentShaderSource = `#version 410
	in vec2 uv;
	uniform sampler2D glyphTexture;
    out vec4 clr;
    void main() {
        clr = vec4(texture(glyphTexture, uv));
    }
` + "\x00"
)

var verts = []float32{
	-1, -1, 0,
	1, -1, -0,
	-1, 1, 0,

	1, -1, -0,
	-1, 1, 0,
	1, 1, 0,
}

func init() {
	runtime.LockOSThread()
}

func main() {
	err := glfw.Init()

	if err != nil {
		panic(err)
	}

	defer glfw.Terminate()

	window, err := glfw.CreateWindow(640, 480, "Testing", nil, nil)

	if err != nil {
		panic(err)
	}

	window.MakeContextCurrent()

	program := initOpenGL()
	vao := makeVertexArrayObject(verts)
	CheckGLErrorsPrint("Pre tex")

	glTex := newGlyphTexture(1024)
	face := initFace()

	view := GlyphView{
		size: 32,
		tex:  glTex,
	}

	for i, r := range "asdfjoisdfj" {
		rasterized := initTex(r, face)
		view.IntoCell(glTex, rasterized, int32(i), 0)
	}

	CheckGLErrors()

	log.Println("Program", program)

	for !window.ShouldClose() {
		draw(vao, window, program, glTex)
	}
}

func initFace() *freetype.Face {
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

	err = face.Pt(32, 72)
	if err != nil {
		panic(err)
	}

	return face
}

func initTex(rn rune, face *freetype.Face) *image.RGBA {

	img, _, err := face.Glyph(rn)
	if err != nil {
		panic(err)
	}

	// err = face.Done()
	// if err != nil {
	// 	panic(err)
	// }

	// err = lib.Done()
	// if err != nil {
	// 	panic(err)
	// }

	return img
}

func initOpenGL() uint32 {
	if err := gl.Init(); err != nil {
		panic(err)
	}

	var data [4]uint8

	gl.DebugMessageCallback(func(source uint32,
		gltype uint32,
		id uint32,
		severity uint32,
		length int32,
		message string,
		userParam unsafe.Pointer,
	) {
		log.Printf("OpenGL Error: %v\n", message)
	}, gl.Ptr(&data[0]))

	version := gl.GoStr(gl.GetString(gl.VERSION))
	log.Println("OpenGL Version", version)

	vertexShader, err := compileShader(vertexShaderSource, gl.VERTEX_SHADER)

	if err != nil {
		panic(err)
	}

	fragmentShader, err := compileShader(fragmentShaderSource, gl.FRAGMENT_SHADER)
	if err != nil {
		panic(err)
	}

	prog := gl.CreateProgram()

	gl.AttachShader(prog, vertexShader)
	gl.AttachShader(prog, fragmentShader)
	gl.LinkProgram(prog)
	return prog
}

func draw(
	vao uint32,
	window *glfw.Window,
	program uint32,
	tex *GlyphTexture,
) {
	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
	gl.UseProgram(program)

	gl.BindTexture(tex.target, tex.handle)
	gl.ActiveTexture(gl.TEXTURE0)

	gl.BindVertexArray(vao)
	gl.DrawArrays(gl.TRIANGLES, 0, int32(len(verts)/3))

	glfw.PollEvents()
	window.SwapBuffers()
}

func makeVertexArrayObject(vertices []float32) uint32 {
	var buffer uint32
	gl.GenBuffers(1, &buffer)
	gl.BindBuffer(gl.ARRAY_BUFFER, buffer)
	gl.BufferData(gl.ARRAY_BUFFER, 4*len(vertices), gl.Ptr(vertices), gl.STATIC_DRAW)

	var vao uint32
	gl.GenVertexArrays(1, &vao)
	gl.BindVertexArray(vao)
	gl.EnableVertexAttribArray(0)
	gl.BindBuffer(gl.ARRAY_BUFFER, buffer)
	gl.VertexAttribPointer(0, 3, gl.FLOAT, false, 0, nil)

	return vao
}

func compileShader(source string, shaderType uint32) (uint32, error) {
	shader := gl.CreateShader(shaderType)

	asCStr, free := gl.Strs(source)
	gl.ShaderSource(shader, 1, asCStr, nil)
	free()

	gl.CompileShader(shader)

	var status int32
	gl.GetShaderiv(shader, gl.COMPILE_STATUS, &status)
	if status == gl.FALSE {
		var logLength int32
		gl.GetShaderiv(shader, gl.INFO_LOG_LENGTH, &logLength)

		log := cString(int(logLength))
		gl.GetShaderInfoLog(shader, logLength, nil, gl.Str(log))
		return 0, fmt.Errorf("Failed to compile shader %v: %v", source, log)
	}

	return shader, nil
}

func cString(len int) string {
	return strings.Repeat("\x00", len+1)
}

const GLYTEX_INTERNAL_FMT = int32(gl.SRGB_ALPHA)

type GlyphTexture struct {
	handle  uint32
	target  uint32
	width   int32
	height  int32
	pixType uint32
	intFmt  int32
}

type GlyphView struct {
	size int32
	tex  *GlyphTexture
}

func (c *GlyphView) IntoCell(
	t *GlyphTexture,
	i *image.RGBA,
	cx int32,
	cy int32,
) {

	iw := int32(i.Rect.Size().X)
	ih := int32(i.Rect.Size().Y)

	cw := t.width / c.size
	ch := t.height / c.size

	fmt.Printf("csize, cx,cy,x,y,w,h: %v %v,%v,%v, %v, %v, %v\n", c.size, cx, cy, cx*cw,
		cy*ch,
		iw,
		ih)

	gl.BindTexture(t.target, t.handle)
	CheckGLErrorsPrint("BindTexture")

	gl.TexSubImage2D(
		t.target,
		0,
		cx*cw,
		cy*ch,
		int32(iw),
		int32(ih),
		gl.RGBA,
		gl.UNSIGNED_BYTE,
		gl.Ptr(i.Pix),
	)
	CheckGLErrorsPrint("TexSubImage2D")
}

func newGlyphTexture(size int32) *GlyphTexture {
	var handle uint32
	gl.GenTextures(1, &handle)

	target := uint32(gl.TEXTURE_2D)

	texture := GlyphTexture{
		handle: handle,
		target: target,
		width:  size,
		height: size,
	}

	gl.BindTexture(target, handle)

	gl.TexParameteri(texture.target, gl.TEXTURE_WRAP_R, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(texture.target, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(texture.target, gl.TEXTURE_MIN_FILTER, gl.NEAREST)
	gl.TexParameteri(texture.target, gl.TEXTURE_MAG_FILTER, gl.NEAREST)

	gl.TexStorage2D(
		target,
		1,
		gl.SRGB8_ALPHA8,
		size,
		size,
	)

	CheckGLErrorsPrint("TexStorage2D")

	return &texture
}

func uploadGlyphToTexture(t *GlyphTexture, i *image.RGBA) {
	w := i.Bounds().Size().X
	h := i.Bounds().Size().Y

	gl.BindTexture(t.target, t.handle)
	CheckGLErrorsPrint("BindTexture")

	gl.TexSubImage2D(
		t.target,
		0,
		0,
		0,
		int32(w),
		int32(h),
		gl.RGBA,
		gl.UNSIGNED_BYTE,
		gl.Ptr(i.Pix),
	)
	CheckGLErrorsPrint("TexSubImage2D")
}

func CheckGLErrors() {
	CheckGLErrorsPrint("")
}

func CheckGLErrorsPrint(s string) {
	glerror := gl.GetError()
	if glerror == gl.NO_ERROR {
		return
	}

	fmt.Printf("%v gl.GetError() reports", s)
	for glerror != gl.NO_ERROR {
		fmt.Printf(" ")
		switch glerror {
		case gl.INVALID_ENUM:
			fmt.Printf("GL_INVALID_ENUM")
		case gl.INVALID_VALUE:
			fmt.Printf("GL_INVALID_VALUE")
		case gl.INVALID_OPERATION:
			fmt.Printf("GL_INVALID_OPERATION")
		case gl.STACK_OVERFLOW:
			fmt.Printf("GL_STACK_OVERFLOW")
		case gl.STACK_UNDERFLOW:
			fmt.Printf("GL_STACK_UNDERFLOW")
		case gl.OUT_OF_MEMORY:
			fmt.Printf("GL_OUT_OF_MEMORY")
		default:
			fmt.Printf("%d", glerror)
		}
		glerror = gl.GetError()
	}
	fmt.Printf("\n")
}

type Tex struct {
	width  uint32
	height uint32
	data   []uint8
}

type Dimensions struct {
	width  uint32
	height uint32
}
