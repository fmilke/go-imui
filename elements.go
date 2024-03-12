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
		gl_Position = vec4(pos.xy + vec2(weights.x * pos.z, 2.0f - weights.y * pos.w), .0f, 1.0f);	
	}

	` + "\x00"

	rectFSSource = `#version 430
	uniform vec4 color = vec4(1.0f);

	out vec4 clr;
	void main() {
        clr = color;
	}
	` + "\x00"
)

type Float = float32
type Pos = int

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
    x := context.ToClipSpaceX(pos.X)
    y := context.ToClipSpaceY(pos.Y)

    w := context.ToClipSpaceX(pos.W) - x
    h := context.ToClipSpaceY(pos.H) - y

	gl.UseProgram(context.RectShader.Program)
	gl.Uniform4f(
		context.RectShader.Ul_Pos,
		x,
		y,
		w,
		h,
	)
    
    c := ColorToGlVec4(color)
	gl.Uniform4f(context.RectShader.Ul_Color, c[0], c[1], c[2], c[3])

	gl.DrawArrays(gl.TRIANGLES, 0, 6)
}
