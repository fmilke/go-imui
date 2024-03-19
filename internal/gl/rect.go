package gl

import (
	"github.com/go-gl/gl/v4.1-core/gl"
    . "dyiui/internal/color"
    . "dyiui/internal/types"
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

type RectShader struct {
	Program ProgramId
	Ul_Pos UniformId
	Ul_Color UniformId
}

func CreateRectShader() RectShader {
	r, err := CompileProgram(rectVSSource, rectFSSource)
	if err != nil {
		panic(err)
	}

    uniforms := GetUniformInfos(r)

	//if DEBUG_SHADERS {
	//	DebugPrintUniformInfos(uniforms)
	//}

	return RectShader{
		Program:  r,
		Ul_Pos:   FindUniformOrPanic("pos", uniforms).Index,
		Ul_Color: FindUniformOrPanic("color", uniforms).Index,
	}
}

func DrawQuad(
	renderer *Renderer,
	pos Quad,
	color Color,
) {
    gl.UseProgram(renderer.shaders.RectShader.Program)
	gl.Uniform4f(
		renderer.shaders.RectShader.Ul_Pos,
		pos.X,
		pos.Y,
		pos.W,
		pos.H,
	)
    
    c := ColorToGlVec4(color)
	gl.Uniform4f(renderer.shaders.RectShader.Ul_Color, c[0], c[1], c[2], c[3])

	gl.DrawArrays(gl.TRIANGLES, 0, 6)
}
