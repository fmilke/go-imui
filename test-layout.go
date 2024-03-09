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

func (ui *UI) Draw() {
    Begin(ui.Context.Width, ui.Context.Height)
    ui.DrawButton("Click me")
    End()
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
    paddingX := 15.0
    paddingY := 10.0
    maxBoxWidth := 400.0
    maxBoxHeight := 200.0

    x := float32(0.0)
    y := float32(0.0)
    maxTextWidth := maxBoxWidth - 2.0 * paddingX

    placements := PlaceSegments(s, ui.App.ttf, ui.App.hbFont, ui.App.fontFace, float32(maxTextWidth), 32.0)

    maxBoxWidth = min(maxBoxWidth, float64(placements.Width))
    maxBoxHeight = min(maxBoxHeight, float64(placements.Height + float32(paddingY) * 2.0))
    DrawQuad(ui.Context, NewAbsPos(x, y, float32(maxBoxWidth), float32(maxBoxHeight)),  123)

    RenderText2(placements, ui.App, NewAbsPos(x, y, 0, 0))

    return id == ui.Clicked
}

func End() {}


