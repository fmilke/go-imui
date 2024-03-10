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

    PointerState
}

type PointerState struct {
    Active bool
    JustActivated bool
    JustReleased bool
    PosX Float
    PosY Float
}

func NewPointerState() PointerState {
    return PointerState{
        Active: false,
        JustActivated: false,
        JustReleased: false,
        PosX: .0,
        PosY: .0,
    }
}

func (s *PointerState) IsWithin(x, y, w, h Float) bool {
    return x <= s.PosX && s.PosX <= (x + w) &&
        y <= s.PosY && s.PosY <= (y + h)
}

func (c *Context) ToClipSpaceX(v Float) Float {
	return (2 * v / float32(c.Width)) - 1.0
}

func (c *Context) ToClipSpaceY(v Float) Float {
	return (2 * v / float32(c.Height)) - 1.0
}
