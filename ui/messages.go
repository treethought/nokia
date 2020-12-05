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
	return w
}

func (w *MessagesWidget) Render(ui *UI) error {
	w.Clear()
	msgs := ui.state.CurrentRoomMessages()
	if len(msgs) == 0 {
		item := cview.NewListItem("No messages yet")
		w.AddItem(item)
		ui.app.QueueUpdateDraw(func() {})
		return nil
	}

	for _, m := range msgs {
		item := cview.NewListItem(m.Body)
		item.SetSecondaryText(m.Sender)
		w.AddItem(item)
	}

	ui.app.QueueUpdateDraw(func() {})
	return nil

}
