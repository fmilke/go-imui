package main

import (
	"fmt"
	"image"
	"io/ioutil"
	"log"
	"math"
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

	in vec2 a_pos;
	in vec2 a_uv;

	out vec2 uv;
    void main() {
		vec2 pos_normalized = vec2(a_pos.x - .5, .5 - a_pos.y)* 2.0;
		gl_Position = vec4(pos_normalized, .0, 1.0);
		uv = vec2(a_uv.x, a_uv.y);
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

const GL_S_FLOAT = 4
const PT_PER_LOGICAL_INCH = 72.0
const PIXELS_PER_LOGICAL_INCH = 96.0 // aka. DPI

const WIN_WIDTH = 640
const WIN_HEIGHT = 480

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

	window, err := glfw.CreateWindow(WIN_WIDTH, WIN_HEIGHT, "Testing", nil, nil)

	if err != nil {
		panic(err)
	}

	window.MakeContextCurrent()

	program := initOpenGL()
	CheckGLErrorsPrint("Pre tex")

	glTex := newGlyphTexture(1024)
	face := initFace(32)

	view := GlyphView{
		size: 32,
		tex:  glTex,
	}

	s := "Asdfjoidfj"

	verticesPerRune := 6
	componentPerVertex := 4
	runeStride := componentPerVertex * verticesPerRune
	co := make([]float32, len(s)*runeStride)

	xadv := 0.0
	coi := 0

	indices := 0
	for i, r := range s {
		rasterized, metrics := initTex(r, face)
		fmt.Printf("%v: %v,%v\n", string(r), metrics.Width, metrics.Height)
		view.IntoCell(glTex, rasterized, int32(i), 0)

		appendRune(xadv, coi, &co, metrics)
		coi += runeStride

		xadv += float64(metrics.Width)
		indices += verticesPerRune
	}

	fmt.Printf("co: %+v\n", co)

	makeSegmentVaos(co)

	CheckGLErrors()

	log.Println("Program", program)

	for !window.ShouldClose() {
		draw(1, window, program, glTex, int32(indices))
	}
}

func appendRune(
	_xAdv float64,
	i int,
	verts *[]float32,
	metrics *freetype.Metrics,
) {
	px_x := 1.0 / float32(WIN_WIDTH)
	px_y := 1.0 / float32(WIN_HEIGHT)

	xAdv := float32(_xAdv)

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

	fmt.Printf("uv: ux: %v, uy %v, uw: %v, uh: %v\n", uOffset, vOffset, uWidth, vWidth)

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
	indices int32,
) {
	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
	gl.UseProgram(program)

	gl.BindTexture(tex.target, tex.handle)
	gl.ActiveTexture(gl.TEXTURE0)

	gl.DrawArrays(gl.TRIANGLES, 0, indices)

	glfw.PollEvents()
	window.SwapBuffers()
}

func makeSegmentVaos(vertices []float32) (uint32, uint32) {
	var buffer uint32
	gl.GenBuffers(1, &buffer)
	gl.BindBuffer(gl.ARRAY_BUFFER, buffer)
	gl.BufferData(gl.ARRAY_BUFFER, 4*len(vertices), gl.Ptr(vertices), gl.STATIC_DRAW)

	var g_pos uint32
	gl.GenVertexArrays(1, &g_pos)
	gl.BindVertexArray(g_pos)
	gl.EnableVertexAttribArray(0)
	gl.BindBuffer(gl.ARRAY_BUFFER, buffer)
	gl.VertexAttribPointer(0, 2, gl.FLOAT, false, GL_S_FLOAT*4, nil)

	gl.EnableVertexAttribArray(1)
	gl.BindBuffer(gl.ARRAY_BUFFER, buffer)
	gl.VertexAttribPointer(1, 2, gl.FLOAT, false, GL_S_FLOAT*4, gl.PtrOffset(GL_S_FLOAT*2))

	return g_pos, g_pos
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

	// fmt.Printf("csize, cx,cy,x,y,w,h: %v %v,%v,%v, %v, %v, %v\n", c.size, cx, cy, cx*cw,
	// 	cy*ch,
	// 	iw,
	// 	ih)

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
