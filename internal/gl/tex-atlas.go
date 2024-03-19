package gl

import (
    "image"
    . "dyiui/internal/types"
)

type AtlasId = int

type AtlasRepo struct {
    entries []Atlas
}

func (at *AtlasRepo) Get(id AtlasId) *Atlas {
    if len(at.entries) == 0 {
        return nil
    }

    return &at.entries[0]
}

func (at *AtlasRepo) Add(id AtlasId, tex GlyphTexture) *Atlas {
	view := GlyphView{
		size: 32,
		tex:  &tex,
	}

    a := Atlas {
        GlyphView: view,
        GlyphTexture: tex,
    }

    at.entries = append(at.entries, a)

    return &a
}

type Atlas struct {
    GlyphView GlyphView
    GlyphTexture GlyphTexture
}

type GlyphView struct {
	size int32
	tex  *GlyphTexture
}

var EmptyQuad = Quad{}

type GlyphAtlas struct {
	index map[rune]Quad
	tex GlyphTexture
}

func NewGlyphAtlas(tex *GlyphTexture) GlyphAtlas {
	return GlyphAtlas {
		index: make(map[rune]Quad),
		tex: *tex,
	}
}

func (at GlyphAtlas) AddQuad(r rune, q Quad) {
	at.index[r] = q
}

// TODO: Add Pixel size and font family
func (at GlyphAtlas) GetQuad(r rune) (Quad, bool) {
	q, ok := at.index[r]

	if ok {
		return q, true
	} else {
		return EmptyQuad, false
	}
}


// Alternate text atlas


type Node struct
{
	Child [2]*Node
	Rect Quad
	ImageId int
}

func (n *Node) IsLeaf() bool {
	return n.Child[0] == nil && n.Child[1] == nil
}

func (n *Node) CanContainImage(i int) {
}

func (n *Node) Insert(i image.Image) *Node {
	if n.IsLeaf() {
		isOccupied := n.ImageId != 0
		if isOccupied {
			return nil
		}


	} else {
	}

	return nil
}

