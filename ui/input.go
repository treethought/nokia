package ui

import (
	"github.com/gdamore/tcell/v2"
	"gitlab.com/tslocum/cview"
)

type InputBox struct {
	*cview.InputField
	submitHandler func(*cview.ListItem)
}

func NewInputBox() *InputBox {
	w := &InputBox{
		InputField: cview.NewInputField(),
	}
	// w.SetInputCapture(w.HandleInput)
	w.SetBackgroundColor(tcell.ColorDefault)
	return w
}


func (w *InputBox) Render(ui *UI) error {
    return nil
}
