package types

type Float = float32
type Pos = int

type Quad struct {
    X Float
    Y Float
    W Float
    H Float
}

func NewQuad(x,y,w,h Float) Quad {
    return Quad {
        X: x,
        Y: y,
        W: w,
        H: h,
    }
}
