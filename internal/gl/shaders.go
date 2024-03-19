package gl

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

type TextShader struct {
    Program ProgramId
	Ul_Offset UniformId
    Ul_Wireframe UniformId
    Ul_TextColor UniformId
}

func CreateTextShader() TextShader {
	r, err := CompileProgram(vertexShaderSource, fragmentShaderSource)
	if err != nil {
		panic(err)
	}

	//if DEBUG_SHADERS {
	//	DebugPrintUniformInfos(GetUniformInfos(r))
	//}

    uniforms := GetUniformInfos(r)

    return TextShader{
        Program: r,
        Ul_Offset: FindUniformOrPanic("offset", uniforms).Index,
        Ul_Wireframe: FindUniformOrPanic("bWireframe", uniforms).Index,
        Ul_TextColor: FindUniformOrPanic("textColor", uniforms).Index,
    }
}


