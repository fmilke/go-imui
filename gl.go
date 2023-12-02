package main

import (
	"strings"
	"fmt"
	"image"
	imageDraw "image/draw"
	"log"
	"unsafe"
	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
)

const GL_S_FLOAT = 4

// texture dump shaders
const (
	texDumpVSSource = `#version 410

	out vec2 uv;
	out vec2 pos;

	void main() {
	}

	` + "\x00"

	texDumpFSSource = `#version 410

	in vec2 uv;
	in vec2 pos;

	uniform sampler2D tex;

	out vec4 clr;
	void main() {
        vec4 clr = texture(tex, uv);
	}
	` + "\x00"
)

const (
	vertexShaderSource = `#version 410

	in vec2 a_pos;
	in vec2 a_uv;

	out vec2 uv;
	out vec3 bc; // barycentric coordinates

	const vec3 bcs[3] = vec3[3](
		vec3(1.0, 0.0, 0.0),
		vec3(0.0, 1.0, 0.0),
		vec3(0.0, 0.0, 1.0)
	);

    void main() {
		vec2 pos_normalized = vec2(a_pos.x - .5, .5 - a_pos.y)* 2.0;
		gl_Position = vec4(pos_normalized, .0, 1.0);
		uv = vec2(a_uv.x, a_uv.y);

		// Add barycentric 
		bc = bcs[gl_VertexID % 3];
    }
` + "\x00"

	fragmentShaderSource = `#version 410
	in vec2 uv;
	in vec3 bc; // barycentric coordinates

	uniform sampler2D glyphTexture;

    out vec4 clr;
    void main() {

		float b = 16.0 * bc[0] * bc[1] * bc[2];
		if (b < .2) {
			b = 1.0;
		} else {
			b = 0.0;
		}
		vec4 wire_frame = vec4(1.0, .0, .0, 1.0) * b;
        vec4 frag_clr = vec4(texture(glyphTexture, uv));

		clr = wire_frame * .5 + frag_clr;
    }
` + "\x00"
)

type GlyphTexture struct {
	handle  uint32
	target  uint32
	width   int32
	height  int32
}

func cString(len int) string {
	return strings.Repeat("\x00", len+1)
}

func initOpenGL() uint32 {
	if err := gl.Init(); err != nil {
		panic(err)
	}

	var data [4]uint8

	gl.Enable(gl.DEBUG_OUTPUT)

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

func (c *GlyphView) IntoCell(
	t *GlyphTexture,
	i *image.RGBA,
	cx int32,
	cy int32,
) {

	if USE_DEBUG_UV {
		return
	}

	iw := int32(i.Rect.Size().X)
	ih := int32(i.Rect.Size().Y)

	cw := t.width / c.size
	ch := t.height / c.size

	gl.BindTexture(t.target, t.handle)
	CheckGLErrorsPrint("BindTexture")

	w := int32(iw)
	h := int32(ih)

	x := cx*cw 
	y := t.height - cy*ch - h

	gl.TexSubImage2D(
		t.target,
		0,
		x,
		y,
		w,
		h,
		gl.RGBA,
		gl.UNSIGNED_BYTE,
		gl.Ptr(i.Pix),
	)
	CheckGLErrorsPrint("TexSubImage2D")
}

const GL_TEXTURE_2D = uint32(gl.TEXTURE_2D)
const NULL_TEX_HANDLE uint32 = 0

func newGlyphTexture(size int32) *GlyphTexture {
	var handle uint32
	gl.GenTextures(1, &handle)

	texture := GlyphTexture{
		handle: handle,
		target: GL_TEXTURE_2D,
		width:  size,
		height: size,
	}

	gl.BindTexture(GL_TEXTURE_2D, handle)

	gl.TexParameteri(texture.target, gl.TEXTURE_WRAP_R, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(texture.target, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(texture.target, gl.TEXTURE_MIN_FILTER, gl.NEAREST)
	gl.TexParameteri(texture.target, gl.TEXTURE_MAG_FILTER, gl.NEAREST)

	gl.TexStorage2D(
		GL_TEXTURE_2D,
		1,
		gl.SRGB8_ALPHA8,
		size,
		size,
	)

	CheckGLErrorsPrint("TexStorage2D")

	gl.BindTexture(GL_TEXTURE_2D,  NULL_TEX_HANDLE)

	return &texture
}

func replaceGlpyhTexture(tex *GlyphTexture, img image.Image) {
	r := img.Bounds()
	rgba := image.NewRGBA(r)
	imageDraw.Draw(rgba, r, img, r.Min, imageDraw.Src)
	data := gl.Ptr(rgba.Pix)

	gl.BindTexture(GL_TEXTURE_2D, tex.handle)

	gl.TexSubImage2D(
		GL_TEXTURE_2D,
		0,
		int32(r.Min.X),
		int32(r.Min.Y),
		int32(r.Max.X),
		int32(r.Max.Y),
		gl.RGBA,
		gl.UNSIGNED_BYTE,
		data,
	)

	gl.BindTexture(GL_TEXTURE_2D, NULL_TEX_HANDLE)
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

