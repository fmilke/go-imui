package main

type Geometry struct {
    Width, Height uint32
}

type Offset struct {
    X, Y uint32
}

func NewOffset(x,y uint32) Offset {
    return Offset {
        X: x,
        Y: y,
    }
}

type Constraint struct {
    MinWidth, MaxWidth, MinHeight, MaxHeight uint32
}

func NewExactConstraint(w, h uint32)  Constraint {
    return Constraint {
        MinWidth:  w,
        MaxWidth: w,
        MinHeight: h,
        MaxHeight: h,
    }
}

type Renderable interface {
    GetGeometry(Constraint) Geometry
    Render(*Context, Offset, Geometry)
}

func GetWinConstraint() Constraint {
    return Constraint {
        MinWidth: 0,
        MaxWidth: WIN_WIDTH,
        MinHeight: 0,
        MaxHeight: WIN_HEIGHT,
    }
}

type Box struct {
    Width uint32
    Height uint32
}

func (b Box) GetGeometry(c Constraint) Geometry {
    return Geometry {
        Width: min(b.Width, c.MaxWidth),
        Height: min(b.Height, c.MaxHeight),
    }
}

func (b Box) Render(c *Context, o Offset, g Geometry) {
    DrawQuad(c, NewAbsPos(float32(o.X), float32(o.Y), float32(g.Width), float32(g.Height)), 123)
}

func RenderBox(w, h uint32) Renderable {
    return Box {
        Width: w,
        Height: h,
    }
}

func TestLayout(context *Context) {
    con := GetWinConstraint()
    box := RenderBox(200, 300)
    geo := box.GetGeometry(con)
    box.Render(context, NewOffset(0,0), geo)
}
