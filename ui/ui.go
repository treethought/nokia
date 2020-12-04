package ui

import (
	"github.com/gdamore/tcell/v2"
	log "github.com/sirupsen/logrus"
	"github.com/treethought/spoon/matrix"
	"gitlab.com/tslocum/cview"
	"maunium.net/go/mautrix/id"
)

type View string

const (
	RoomList    View = "rooms"
	MessageList View = "main"
	Header      View = "header"
)

type UI struct {
	app           *cview.Application
	grid          *cview.Grid
	Widgets       map[View]WidgetRenderer
	m             *matrix.Client
	state         *State
	currentwidget WidgetRenderer
}

type WidgetRenderer interface {
	cview.Primitive
	Render(ui *UI) error
}

func New() *UI {
	tapp := cview.NewApplication()
	m, err := matrix.New()
	if err != nil {
		panic(err)
	}

	ui := &UI{app: tapp, m: m}

	ui.state = NewState(ui)

	ui.Widgets = make(map[View]WidgetRenderer)
	return ui
}

func (ui *UI) init() {
	ui.initWidgets()
	ui.initGrid()
}

func (ui *UI) Render() {
	for _, w := range ui.Widgets {
		go w.Render(ui)
	}
}

func (ui *UI) initWidgets() {

	header := NewHeaderWidget()

	rooms := NewRoomsWidget()
	rooms.SetBackgroundColor(tcell.ColorIndianRed)
	rooms.SetSelectHandler(ui.roomSelectHandler)

	msgs := NewMessagesWidget()

	ui.Widgets[RoomList] = rooms
	ui.Widgets[Header] = header
	ui.Widgets[MessageList] = msgs

}

func (ui *UI) roomSelectHandler(item *cview.ListItem) {
	roomText := item.GetMainText()
	ui.state.CurrentRoom = id.RoomID(roomText)
	ui.Widgets[MessageList].Render(ui)
	ui.app.SetFocus(ui.Widgets[MessageList])

}

func (ui *UI) toggleFocus() {
	if ui.currentwidget == ui.Widgets[RoomList] {
		ui.app.SetFocus(ui.Widgets[MessageList])
		ui.currentwidget = ui.Widgets[MessageList]
	} else {
		ui.app.SetFocus(ui.Widgets[RoomList])
		ui.currentwidget = ui.Widgets[RoomList]

	}
}

func (ui *UI) initGrid() {
	ui.grid = cview.NewGrid()
	ui.grid.SetRows(-1, -3, -1)
	ui.grid.SetColumns(0, -3, 0)
	ui.grid.SetBorders(false)

	ui.grid.AddItem(ui.Widgets[Header], 0, 0, 1, 3, 0, 0, true)
	ui.grid.AddItem(ui.Widgets[RoomList], 1, 0, 3, 1, 0, 0, true)
	ui.grid.AddItem(ui.Widgets[MessageList], 1, 1, 3, 3, 0, 0, true)

	log.Info("GRIDD")
	log.Info(ui.grid)

	ui.app.SetRoot(ui.grid, true)
	ui.app.QueueUpdateDraw(func() {})

	ui.app.SetFocus(ui.Widgets[RoomList])
	ui.currentwidget = ui.Widgets[RoomList]

	ui.app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {

		key := event.Key()
		switch key {
		case tcell.KeyTAB:
			ui.toggleFocus()
			return nil
		}
		return event
	})

}

func (u *UI) setSyncHandlers() {
	u.m.SetMessageHandler(u.state.ProcessMessage)
	u.m.SetSyncCallback(u.state.toDisk)
}

func Start() {
	ui := New()
	ui.init()
	ui.setSyncHandlers()
	ui.Render()

	ui.m.Login()

	go ui.m.Sync()

	err := ui.app.Run()
	if err != nil {
		panic(err)
	}

}
