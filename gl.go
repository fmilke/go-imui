package main

import (
	"errors"
	"fmt"
	"image"
	imageDraw "image/draw"
	"log"
	"strings"
	"unsafe"

	"github.com/go-gl/gl/v4.1-core/gl"
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

    uniform vec2 offset;

	out vec2 uv;
	out vec3 bc; // barycentric coordinates

	const vec3 bcs[3] = vec3[3](
		vec3(1.0, 0.0, 0.0),
		vec3(0.0, 1.0, 0.0),
		vec3(0.0, 0.0, 1.0)
	);

    void main() {
		vec2 pos_normalized = vec2(a_pos.x, a_pos.y)* 2.0 + offset;
        pos_normalized.y = -pos_normalized.y;

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
    uniform float bWireframe;
    uniform vec3 textColor;

    out vec4 clr;
    void main() {

		float b = 16.0 * bc[0] * bc[1] * bc[2];
		if (b < .2) {
			b = 1.0;
		} else {
			b = 0.0;
		}
		vec4 wire_frame = vec4(1.0, .0, .0, 1.0) * b * bWireframe;
        float opacity = vec4(texture(glyphTexture, uv)).r;
        vec4 frag_clr = vec4(textColor, opacity);

		clr = wire_frame + frag_clr;
    }
` + "\x00"
)

func ToGlClipSpace(value Float, dim Float) Float {
	return (2 * value / dim) - 1.0
}

type ProgramId = uint32
type UniformId = int32

const INACTIVE_UNIFORM int32 = -1

type GlyphTexture struct {
	handle uint32
	target uint32
	width  int32
	height int32
}

func cString(len int) string {
	return strings.Repeat("\x00", len+1)
}

func GetUniformInfos(program ProgramId) []UniformInfo {
    var infos []UniformInfo
    var count int32
    gl.GetProgramiv(program, gl.ACTIVE_UNIFORMS, &count)

	for i := int32(0); i < count; i++ {
        infos = append(infos, GetUniformInfo(program, i))
	}
    
    return infos
}

func DebugPrintUniformInfos(infos []UniformInfo) {
	log.Printf("Has %d uniforms:\n", len(infos))
    for _, info := range(infos) {
		log.Printf("Info: %+v\n", info)
	}
}

func GetUniformInfo(program ProgramId, uniform UniformId) UniformInfo {

	if uniform == INACTIVE_UNIFORM {
		panic("Trying to get info for inactive uniform")
	}

	maxNameLen := gl.UNIFORM_NAME_LENGTH
	buf := cString(maxNameLen)

	var written int32
	var size int32
	var t uint32
	gl.GetActiveUniform(program, uint32(uniform), int32(maxNameLen), &written, &size, &t, gl.Str(buf))
	err := GetWrappedGlError()
	if err != nil {
		panic(err)
	}

	return UniformInfo{
        Program: program,
		Index: uniform,
		Name:  strings.TrimRight(buf, "\x00"), 
		Size:  size,
		Type:  t,
	}
}

type UniformInfo struct {
    Program ProgramId
	Index UniformId
	Name  string
	Size  int32
	Type  uint32
}

func initOpenGL() *Context {
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

	glslVersion := gl.GoStr(gl.GetString(gl.SHADING_LANGUAGE_VERSION))
	log.Println("GLSL Version", glslVersion)

	ctx := Context{
		Width:  WIN_WIDTH,
		Height: WIN_HEIGHT,
        PointerState: NewPointerState(),
	}

	ctx.TextShader = CreateTextShader()
	ctx.RectShader = CreateRectShader()

    gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)
    gl.Enable(gl.BLEND)

	return &ctx
}

func CreateTextShader() TextShader {
	r, err := CompileProgram(vertexShaderSource, fragmentShaderSource)
	if err != nil {
		panic(err)
	}

	if DEBUG_SHADERS {
		DebugPrintUniformInfos(GetUniformInfos(r))
	}


    uniforms := GetUniformInfos(r)

    return TextShader{
        Program: r,
        Ul_Offset: FindUniformOrPanic("offset", uniforms).Index,
        Ul_Wireframe: FindUniformOrPanic("bWireframe", uniforms).Index,
        Ul_TextColor: FindUniformOrPanic("textColor", uniforms).Index,
    }
}

func FindUniformOrPanic(name string, infos []UniformInfo) UniformInfo {
    for _, i := range(infos) {
        if i.Name == name {
            return i;
        }
    }

    panic("could not find uniform " + name)
}

func CreateRectShader() RectShader {
	r, err := CompileProgram(rectVSSource, rectFSSource)
	if err != nil {
		panic(err)
	}

    uniforms := GetUniformInfos(r)

	if DEBUG_SHADERS {
		DebugPrintUniformInfos(uniforms)
	}

	return RectShader{
		Program:  r,
		Ul_Pos:   FindUniformOrPanic("pos", uniforms).Index,
		Ul_Color: FindUniformOrPanic("color", uniforms).Index,
	}
}

func CompileProgram(vs string, fs string) (ProgramId, error) {
	vertexShader, err := compileShader(vs, gl.VERTEX_SHADER)
	if err != nil {
		return 0, err
	}

	fragmentShader, err := compileShader(fs, gl.FRAGMENT_SHADER)
	if err != nil {
		return 0, err
	}

	prog := gl.CreateProgram()

	gl.AttachShader(prog, vertexShader)
	CheckGLErrors()
	gl.AttachShader(prog, fragmentShader)
	CheckGLErrors()
	gl.LinkProgram(prog)

	var isLinked int32
	gl.GetProgramiv(prog, gl.LINK_STATUS, &isLinked)
	if isLinked == gl.FALSE {
		gl.DeleteProgram(prog)
		return 0, errors.New("could not link program")
	}

	return prog, nil
}

func GetUniformLocation(prog uint32, name string) int32 {
	asCStr, free := gl.Strs(name)
	loc := gl.GetUniformLocation(prog, *asCStr)
	free()

    log.Printf("program: %d, uniform: %s, loc: %d\n", prog, name, loc)
	if loc == INACTIVE_UNIFORM {
		//panic(fmt.Sprintf("could not find uniform location %s", name))
	}

	return loc
}

func BeginFrame() {
	// Clear previous buffer
	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
}

func FinishFrame() {

}

func DrawFrame(
    app *App,
    context *Context,
) {
    ui := NewUI(context, app)
    ui.Draw()
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
		return 0, fmt.Errorf("failed to compile shader %v: %v", source, log)
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

	x := cx * cw
	//	y := t.height - cy*ch - h
	y := cy * ch

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

	gl.BindTexture(GL_TEXTURE_2D, NULL_TEX_HANDLE)

	return &texture
}

func readGlyphTexture(tex *GlyphTexture) image.Image {
	gl.BindTexture(GL_TEXTURE_2D, tex.handle)
	img := image.NewRGBA(image.Rect(0, 0, int(tex.width), int(tex.height)))
	//gl.ReadPixels(0, 0, tex.width, tex.height, gl.RGBA, gl.UNSIGNED_BYTE, gl.Ptr(img.Pix))
	gl.GetTexImage(GL_TEXTURE_2D, 0, gl.RGBA, gl.UNSIGNED_BYTE, gl.Ptr(img.Pix))

	return img
}

func replaceGlyphTexture(tex *GlyphTexture, img image.Image) {
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

func GetWrappedGlError() error {
	glerror := gl.GetError()
	if glerror == gl.NO_ERROR {
		return nil
	}

	var sb strings.Builder

	for glerror != gl.NO_ERROR {
		sb.WriteString(GetErrorAsString(glerror))
		glerror = gl.GetError()
	}

	return errors.New(sb.String())
}

func GetErrorAsString(glError uint32) string {
	switch glError {
	case gl.INVALID_ENUM:
		return "GL_INVALID_ENUM"
	case gl.INVALID_VALUE:
		return "GL_INVALID_VALUE"
	case gl.INVALID_OPERATION:
		return "GL_INVALID_OPERATION"
	case gl.STACK_OVERFLOW:
		return "GL_STACK_OVERFLOW"
	case gl.STACK_UNDERFLOW:
		return "GL_STACK_UNDERFLOW"
	case gl.OUT_OF_MEMORY:
		return "GL_OUT_OF_MEMORY"
	default:
		return fmt.Sprintf("<errno: %d>", glError)
	}
}

func CheckGLErrorsPrint(s string) {
	glError := gl.GetError()
	if glError == gl.NO_ERROR {
		return
	}

	log.Printf("%v gl.GetError() reports", s)
	for glError != gl.NO_ERROR {
		log.Printf(" %s\n", GetErrorAsString(glError))
		glError = gl.GetError()
	}
    log.Println("")
}
