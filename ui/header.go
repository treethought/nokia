package ui

import (
	"github.com/gdamore/tcell/v2"
	"gitlab.com/tslocum/cview"
)

type HeaderWidget struct {
	*cview.TextView
}

func NewHeaderWidget() *HeaderWidget {
	return &HeaderWidget{TextView: cview.NewTextView()}
}

func (w *HeaderWidget) Render(ui *UI) error {
	w.Clear()
	w.SetText("bend the spoon")
	w.SetBackgroundColor(tcell.ColorDefault)
	return nil

}
