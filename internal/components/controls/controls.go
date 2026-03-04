package controls

import (
    "github.com/rivo/tview"
)

type ControlsView struct {
    *tview.Table
}

type CellThemer interface {
    SetTableCellTheme(table *tview.Table, row int, col int, foreground, background string)
}

func NewControlsView(themer CellThemer) *ControlsView {
    cv := &ControlsView{}
    cv.Table = tview.NewTable()
    cv.SetBorder(true)
    cv.SetTitle("Controls")

    col := 0
    index := 0
    for _, v := range cv.CreateControlsHelp() {
        if index == 3 {
            index = 0
            col += 2
        }

        cv.SetCell(index, col, tview.NewTableCell(v.Message))
        cv.SetCell(index, col+1, tview.NewTableCell(v.Key))
        themer.SetTableCellTheme(cv.Table, index, col+1, "orange", "background")
        index++
    }

    return cv
}

func (cv ControlsView) CreateControlsHelp() []ControlsHelp {
    return []ControlsHelp{
        {Key: "<q>", Message: "Quit"},
        {Key: "<d> / <Enter>", Message: "Show Secret"},
        {Key: "<b> / <Esc>", Message: "Close Secret"},
        {Key: "</>", Message: "Search"},
        {Key: "<ctrl+r>", Message: "Reload"},
        {Key: "<y>", Message: "Copy Value"},
        {Key: "<r>", Message: "Re-fetch Secret"},
        {Key: "<1-9>", Message: "Select Vault"},
    }
}

type ControlsHelp struct {
    Key     string
    Message string
}
