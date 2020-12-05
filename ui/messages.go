package ui

import (
	"github.com/gdamore/tcell/v2"
	"gitlab.com/tslocum/cview"
)

type MessagesWidget struct {
	*cview.List
	messages []*Message
}

func NewMessagesWidget() *MessagesWidget {
	w := &MessagesWidget{
		List:     cview.NewList(),
		messages: make([]*Message, 0),
	}
	w.SetBackgroundColor(tcell.ColorDefault)
	w.SetMainTextColor(tcell.ColorLightCyan)
	w.SetSecondaryTextColor(tcell.ColorWhite)
	return w
}

func (w *MessagesWidget) Render(ui *UI) error {
	log.Print("Rendering Messages")
	w.Clear()
	msgs := ui.state.CurrentRoomMessages()
	if len(msgs) == 0 {
		item := cview.NewListItem("No messages yet")
		w.AddItem(item)
		ui.app.QueueUpdateDraw(func() {})
		return nil
	}

	for _, m := range msgs {
		item := cview.NewListItem(m.Sender)
		item.SetSecondaryText(m.Body)
		w.AddItem(item)
	}

	ui.app.QueueUpdateDraw(func() {})
	return nil

}
