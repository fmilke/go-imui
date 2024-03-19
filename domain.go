package main

import (
    . "dyiui/internal/layout"
)

// this file represents how users
// would use the library

type AppState struct {
    Tab int
}

var appState AppState = AppState{}

func RenderUI(
    ui *UI,
) {
    Begin(ui.Context.Width, ui.Context.Height)

    //fmt.Printf("Tab is %v\n", appState.Tab)
    
    if appState.Tab == 0 {
        RenderFirstTab(ui)
    } else {
        RenderSecondTab(ui)
    }

    End()
}

func RenderFirstTab(ui *UI) {
    clicked := ui.DrawButton("Go to second page")

    if clicked {
        appState.Tab = 1
    }
}

func RenderSecondTab(ui *UI) {
    clicked := ui.DrawButton("Go to first page")

    if clicked {
        appState.Tab = 0
    }
}

