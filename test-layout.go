package main

type ElementId = uint32

type UI struct {
    Parents []ElementId
    Clicked ElementId
    Context *Context
    App *App
}

func NewUI(context *Context, app *App) UI {
    return UI {
        Context: context,
        App: app,
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

    placements := PlaceSegments(s, ui.App.ttf, ui.App.hbFont, ui.App.fontFace, float32(maxTextWidth), 32.0)

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
    DrawQuad(ui.Context, NewAbsPos(boxX, boxY, float32(maxBoxWidth), float32(maxBoxHeight)),  color)

    textX := boxX + float32(paddingX)
    textY := boxY + float32(paddingY)
    RenderText2(placements, ui.App, NewAbsPos(textX, textY, 0, 0))

    // tell state
    return clicked
}

func End() {}


