package layout

import (
	. "dyiui/internal/color"
	. "dyiui/internal/gl"
	. "dyiui/internal/types"
	. "dyiui/internal/ui"
	"fmt"
)

type ElementId = uint32

type UI struct {
    Parents []ElementId
    Clicked ElementId
    Context *Context
    Renderer *Renderer
}

func NewUI(context *Context, r *Renderer) UI {
    return UI {
        Context: context,
        Renderer: r,
    }
}

func Begin(x, y uint32) {
}

func (ui *UI) GenerateNewId() ElementId {
    if ui.Parents == nil || len(ui.Parents) == 0 {
        return 1;
    }

    parentId := ui.Parents[len(ui.Parents)-1];
    return parentId *  139123 % 45837;
}

func (ui *UI) DrawButton(s string) bool {
    id := ui.GenerateNewId()

    // get from outside
    // collect input, refactor
    paddingX := 15.0
    paddingY := 10.0
    maxBoxWidth := 400.0
    maxBoxHeight := 200.0


    // calculate layout
    boxX := float32(0.0)
    boxY := float32(0.0)
    maxTextWidth := maxBoxWidth - 2.0 * paddingX

    font := ui.Renderer.Fonts.Get()
    if font == nil {
        // not ready to render font yet
        fmt.Printf("not yet ready to render\n")
        return false
    }
    placements := PlaceSegments(s, font.Ttf, font.HbFont, font.FontFace, float32(maxTextWidth), 32.0)

    maxBoxWidth = min(maxBoxWidth, float64(placements.Width) + paddingX * 2.0)
    maxBoxHeight = min(maxBoxHeight, float64(placements.Height + float32(paddingY) * 2.0))

    // determine state
    var color Color
    mouseOver := ui.Context.PointerState.IsWithin(boxX, boxY, float32(maxBoxWidth), float32(maxBoxHeight))
    clicked := mouseOver && ui.Context.PointerState.JustActivated

    if clicked {
        color = 0x00ff00ff
        ui.Clicked = id
    } else if mouseOver {
        color = 0xff0000ff
    } else {
        color = 0x00ffffff
    }

    // render elements

    q := NewQuad(boxX, boxY, float32(maxBoxWidth), float32(maxBoxHeight))
    ui.Renderer.MapToClipSpace(&q)
    DrawQuad(ui.Renderer, q,  color)

    textX := boxX + float32(paddingX)
    textY := boxY + float32(paddingY)

    atlas := ui.Renderer.GetAtlas()
    args := RenderTextArgs {
        FontFace: font.FontFace,
        GlyphView: atlas.GlyphView,
        GlyphTex: atlas.GlyphTexture,
    }

    ui.Renderer.RenderText2(placements, &args, NewQuad(textX, textY, 0, 0))

    // tell state
    return clicked
}

func End() {}


