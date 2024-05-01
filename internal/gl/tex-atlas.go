package gl

import (
	"dyiui/internal/lru"
	"dyiui/internal/types"
	"image"

	"github.com/benoitkugler/textlayout/fonts"
	"github.com/go-gl/gl/v4.1-core/gl"
)

type AtlasId = int

// AtlasRepo

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

    storeableGlyphs:=1024
    a := Atlas {
        GlyphView: &view,
        GlyphTexture: &tex,
        Cache: lru.NewLRUCache[CacheEntry](uint(storeableGlyphs)),
    }

    at.entries = append(at.entries, a)

    return &a
}


// Atlas

type CacheEntry = types.Quad

type Atlas struct {
    GlyphView *GlyphView
    GlyphTexture *GlyphTexture
    Cache *lru.LRUCache[CacheEntry]
}

func (at *Atlas) GetSlot(r fonts.GID) (types.Quad, bool) {
    if e := at.Cache.Get(r); e != nil {
        return *e, true
    } else {
        q := at.GlyphView.Next()
        at.Cache.Store(r, q)
        return q, false
    }
}

// Glyph View

type GlyphView struct {
	size int32
	tex  *GlyphTexture
    next int
}

func (view *GlyphView) Next() types.Quad {

    cellsPerRow := int(view.tex.width / view.size)
    cellsPerCol := int(view.tex.height / view.size)

    x := view.next % cellsPerRow
    y := view.next / cellsPerCol

    view.next += 1

    return types.Quad {
        X: float32(x * int(view.size)),
        Y: float32(y * int(view.size)),
        W: float32(view.size),
        H: float32(view.size),
    }
}

func (view *GlyphView) IntoCell(
	t *GlyphTexture,
	i *image.RGBA,
    pos types.Quad,
) {
    // todo handle difference between
    // rect size and cell size
	iw := int32(i.Rect.Size().X)
	ih := int32(i.Rect.Size().Y)

	gl.BindTexture(t.target, t.handle)
	CheckGLErrorsPrint("BindTexture")

	gl.TexSubImage2D(
		t.target,
		0,
		int32(pos.X),
		int32(pos.Y),
		iw,
		ih,
		gl.RGBA,
		gl.UNSIGNED_BYTE,
		gl.Ptr(i.Pix),
	)
}
