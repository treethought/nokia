package ui

import (
	"github.com/gdamore/tcell/v2"
	"gitlab.com/tslocum/cview"
)

type InputWidget struct {
	*cview.InputField
}

func NewInputWidget() *InputWidget {
	w := &InputWidget{
		InputField: cview.NewInputField(),
	}
	w.SetBackgroundColor(tcell.ColorDefault)
	w.SetLabelColor(tcell.ColorLightCyan)
	w.SetFieldTextColor(tcell.ColorWhite)
	w.SetLabel(">")
	return w
}

func (w *InputWidget) Render(ui *UI) error {
	log.Print("Rendering Input")
	w.SetPlaceholder("type a message")
	w.SetText("Hello")
	return nil

}
