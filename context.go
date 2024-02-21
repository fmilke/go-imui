package main

type RectShader struct {
	Program uint32
	Ul_Pos int32
	Ul_Color int32
}

type Context struct {
	TextShader uint32
	RectShader RectShader

	Width uint32
	Height uint32
}

