package ui

import (
	"fmt"
	"time"

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

func unixToTime(unix int64) time.Time {
	timestamp := time.Now()
	if unix != 0 {
		timestamp = time.Unix(unix/1000, unix%1000*1000)
	}
	return timestamp
}

func timestampToString(t time.Time) string {
	// log.Print(ts)
	// t := time.Unix(ts, 0)
	return fmt.Sprintf(t.Format("02/01/2006, 15:04:05"))
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
		t := unixToTime(m.Timestamp)
		title := fmt.Sprintf("%s: %s", m.Sender, timestampToString(t))
		item := cview.NewListItem(title)

		item.SetSecondaryText(m.Body)
		w.AddItem(item)
	}

	title := fmt.Sprintf("%d Messages", len(msgs))
	w.SetTitle(title)
	w.SetBorder(true)
	w.SetCurrentItem(-1)
	ui.app.QueueUpdateDraw(func() {})
	return nil

}
