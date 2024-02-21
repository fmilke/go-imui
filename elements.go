package main

import (
	"github.com/go-gl/gl/v4.1-core/gl"
)

const (
	rectVSSource = `#version 430
	uniform vec4 pos = vec4(.0f, .0f, .0f, 1.0f);
	
	const vec2 size_weights[6] = vec2[6](
		vec2(0.0, 1.0),
		vec2(1.0, 1.0),
		vec2(0.0, 0.0),

		vec2(1.0, 1.0),
		vec2(0.0, 0.0),
		vec2(1.0, 0.0)
	);

	void main() {
		vec2 weights = size_weights[gl_VertexID];
		gl_Position = vec4(pos.xy + vec2(weights.x * pos.z, weights.y + pos.w), .0f, 1.0f);	
	}

	` + "\x00"

	rectFSSource = `#version 430
	uniform vec4 color = vec4(1.0f);

	out vec4 clr;
	void main() {
		clr = color * .31f + vec4(1.0f, 0.5f, .5f, 1.0f);
	}
	` + "\x00"
)

type Float = float32
type Pos = int
type Color = uint32

const (
	Absolute Pos = 0
	Relative Pos = 1
)

type Position struct {
	Pos Pos
	X   Float
	Y   Float
	W   Float
	H   Float
}

func NewRelPos(X, Y, W, H Float) Position {
	return Position{
		X:   X,
		Y:   Y,
		W:   W,
		H:   H,
		Pos: Relative,
	}
}

func NewAbsPos(X, Y, W, H Float) Position {
	return Position{
		X:   X,
		Y:   Y,
		W:   W,
		H:   H,
		Pos: Absolute,
	}
}

func DrawQuad(
	context *Context,
	pos Position,
	color Color,
) {
	x := ToGlClipSpace(pos.X, float32(context.Width))
	y := ToGlClipSpace(pos.Y, float32(context.Height))

	w := ToGlClipSpace(pos.W, float32(context.Width)) - x
	h := ToGlClipSpace(pos.H, float32(context.Height)) - y

	gl.UseProgram(context.RectShader.Program)
	gl.Uniform4f(
		context.RectShader.Ul_Pos,
		x,
		y,
		w,
		h,
	)

	gl.Uniform4f(context.RectShader.Ul_Color, 1.0, 1.0, 0.0, 1.0)
	gl.DrawArrays(gl.TRIANGLES, 0, 6)
}
