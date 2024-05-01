package gl

import (
	. "dyiui/internal/types"
	. "dyiui/internal/text"
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

func ToGlClipSpace(value Float, dim Float) Float {
	return (2 * value / dim) - 1.0
}

type ProgramId = uint32
type UniformId = int32
type BufferId = uint32
type VaoId = uint32

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

type shaders struct {
    RectShader RectShader
    TextShader TextShader
}

type Renderer struct {
    shaders shaders
    Width int
    Height int
    Fonts FontRepo
    atlases AtlasRepo
}

func InitRenderer(initWidth, initHeight int) *Renderer {
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
        if !strings.HasPrefix(message, "Shader Stats") {
            log.Printf("OpenGL Error: %v\n", message)
        }
	}, gl.Ptr(&data[0]))

	version := gl.GoStr(gl.GetString(gl.VERSION))
	log.Println("OpenGL Version", version)

	glslVersion := gl.GoStr(gl.GetString(gl.SHADING_LANGUAGE_VERSION))
	log.Println("GLSL Version", glslVersion)

	r := Renderer{
		Width:  initWidth,
		Height: initHeight,
        shaders: shaders{},
        Fonts: NewFontRepo(),
	}

    path := GetSomeFont()
    fmt.Printf("loading font from %s\n", path)
    r.Fonts.Load(path)

	r.shaders.TextShader = CreateTextShader()
	r.shaders.RectShader = CreateRectShader()

    gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)
    gl.Enable(gl.BLEND)

	return &r
}

func (c *Renderer) ToClipSpaceX(v Float) Float {
	return (2 * v / float32(c.Width)) - 1.0
}

func (c *Renderer) ToClipSpaceY(v Float) Float {
	return (2 * v / float32(c.Height)) - 1.0
}

func (c *Renderer) MapToClipSpace(q *Quad) {
    q.X = c.ToClipSpaceX(q.X)
    q.Y = c.ToClipSpaceY(q.Y)
    q.W = c.ToClipSpaceX(q.W) + 1.0
    q.H = c.ToClipSpaceY(q.H) + 1.0
}

func (r *Renderer) GetAtlas() *Atlas {
    var id AtlasId = 0
    atlas := r.atlases.Get(id)

    if atlas == nil {
        fmt.Printf("atlas not existing yet, creating one\n")
        tex := NewGlyphTexture(1024)
        atlas = r.atlases.Add(id, *tex)
    }

    return atlas
}

func FindUniformOrPanic(name string, infos []UniformInfo) UniformInfo {
    for _, i := range(infos) {
        if i.Name == name {
            return i;
        }
    }

    panic("could not find uniform " + name)
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

func makeTextVertexArrays(vertices []float32) {
    var buffer BufferId
    gl.GenBuffers(1, &buffer)
    gl.BindBuffer(gl.ARRAY_BUFFER, buffer)
	gl.BufferData(gl.ARRAY_BUFFER, 4*len(vertices), gl.Ptr(vertices), gl.STATIC_DRAW)

    var vao VaoId
    gl.GenVertexArrays(1, &vao)
    gl.BindVertexArray(vao)
    gl.BindBuffer(gl.ARRAY_BUFFER, buffer)
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

const GL_TEXTURE_2D = uint32(gl.TEXTURE_2D)
const NULL_TEX_HANDLE uint32 = 0

func NewGlyphTexture(size int32) *GlyphTexture {
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

	gl.BindTexture(GL_TEXTURE_2D, NULL_TEX_HANDLE)

	return &texture
}

func ReadGlyphTexture(tex *GlyphTexture) image.Image {
	gl.BindTexture(GL_TEXTURE_2D, tex.handle)
	img := image.NewRGBA(image.Rect(0, 0, int(tex.width), int(tex.height)))
	gl.GetTexImage(GL_TEXTURE_2D, 0, gl.RGBA, gl.UNSIGNED_BYTE, gl.Ptr(img.Pix))

	return img
}

func ReplaceGlyphTexture(tex *GlyphTexture, img image.Image) {
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
