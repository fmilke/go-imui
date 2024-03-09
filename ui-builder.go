package main

type Constraints struct {
	MaxWidth  uint32
	MaxHeight uint32
}

type Cursor struct{}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////

// type Constraints struct {
// 	MaxWidth  uint32
// 	MaxHeight uint32
// }

// type UIElement interface {
// 	Render(Constraints)
// }

// type UIBuilder struct{}

// func (*UIBuilder) ColumnLayout() {}

// type RootElement struct{}

// func Render(c Constraints) {

// }

// type ColumnLayout struct {
// 	Cols  uint32
// 	Basis uint32
// }

// func NewColumnLayout() *ColumnLayout {
// 	return &ColumnLayout{
// 		Cols:  0,
// 		Basis: 1,
// 	}
// }

// func (cl *ColumnLayout) AddColumn(basis uint32) {
// 	cl.Cols += 1
// 	cl.Basis += basis
// }
