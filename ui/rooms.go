package ui

import (
	"time"

	"github.com/gdamore/tcell/v2"
	"gitlab.com/tslocum/cview"
)

type RoomWidget struct {
	*cview.List
	rooms         []string
	selectHandler func(*cview.ListItem)
}

func NewRoomsWidget() *RoomWidget {
	w := &RoomWidget{
		List: cview.NewList(),
	}
	w.SetInputCapture(w.HandleInput)
	w.SetBackgroundColor(tcell.ColorDefault)
	return w
}

func (w *RoomWidget) load(rooms []string) {
	w.Clear()
	for _, r := range rooms {
		item := cview.NewListItem(r)
		w.AddItem(item)
	}

}

func (w *RoomWidget) Render(ui *UI) error {
	w.Clear()

	for len(ui.state.Rooms) == 0 {
		if len(w.GetItems()) > 3 {
			w.Clear()
		}
		w.AddItem(cview.NewListItem("..."))
		ui.app.QueueUpdateDraw(func() {})
		time.Sleep(time.Second * 1)

	}

	if len(ui.state.Rooms) == 0 {
		item := cview.NewListItem("Loading rooms...")
		w.AddItem(item)
		ui.app.QueueUpdateDraw(func() {})
		return nil

	}

	for id, r := range ui.state.Rooms {
		var item *cview.ListItem

		if r.Name != "" {
			item = cview.NewListItem(r.Name)
		} else {
			item = cview.NewListItem(id.String())

		}
		item.SetReference(r)
		w.AddItem(item)
	}
	ui.app.QueueUpdateDraw(func() {})
	return nil

}

func (w *RoomWidget) SetSelectHandler(f func(item *cview.ListItem)) {
	w.selectHandler = f

}

func (w *RoomWidget) HandleInput(event *tcell.EventKey) *tcell.EventKey {
	if len(w.GetItems()) < 2 {
		return event
	}

	key := event.Key()
	switch key {
	case tcell.KeyEnter:
		w.selectHandler(w.GetCurrentItem())
		return nil

	case tcell.KeyRune:
		switch event.Rune() {
		case 'g': // Home.
			w.SetCurrentItem(0)
		case 'G': // End.
			w.SetCurrentItem(-1)
		case 'j': // Down.
			cur := w.GetCurrentItemIndex()
			w.SetCurrentItem(cur + 1)
		case 'k': // Up.
			cur := w.GetCurrentItemIndex()
			w.SetCurrentItem(cur - 1)
		}

		return nil
	}

	return event
}
