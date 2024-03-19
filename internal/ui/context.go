package ui

import (
    . "dyiui/internal/types"
)

type Context struct {
	Width uint32
	Height uint32

    PointerState PointerState
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


