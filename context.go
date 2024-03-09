package main

type RectShader struct {
	Program ProgramId
	Ul_Pos UniformId
	Ul_Color UniformId
}

type TextShader struct {
    Program ProgramId
	Ul_Offset UniformId
    Ul_Wireframe UniformId
    Ul_TextColor UniformId
}

type Context struct {
	TextShader TextShader
	RectShader RectShader

	Width uint32
	Height uint32
}

func (c *Context) ToClipSpaceX(v Float) Float {
	return (2 * v / float32(c.Width)) - 1.0
}

func (c *Context) ToClipSpaceY(v Float) Float {
	return (2 * v / float32(c.Height)) - 1.0
}
