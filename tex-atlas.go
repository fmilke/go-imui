package main

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



