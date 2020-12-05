package ui

import (
	"github.com/gdamore/tcell/v2"
	"github.com/treethought/spoon/logger"
	"github.com/treethought/spoon/matrix"
	"gitlab.com/tslocum/cview"
)

type View string

var log = logger.GetLoggerInstance()

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
	ui.state.fromDisk()

	ui.Widgets = make(map[View]WidgetRenderer)
	return ui
}

func (ui *UI) init() {
	ui.initWidgets()
	ui.initGrid()
	log.Print("UI intiialized")
}

func (ui *UI) Render() {
	for _, w := range ui.Widgets {
		go w.Render(ui)
	}
}

func (ui *UI) initWidgets() {

	// header := NewHeaderWidget()
	// ui.Widgets[Header] = header

	rooms := NewRoomsWidget()
	rooms.SetSelectHandler(ui.roomSelectHandler)
	ui.Widgets[RoomList] = rooms

	msgs := NewMessagesWidget()
	ui.Widgets[MessageList] = msgs
	log.Print("widgets initialized")

}

func (ui *UI) roomSelectHandler(item *cview.ListItem) {
	ref := item.GetReference()

	room, ok := ref.(*Room)
	if !ok {
		panic("room ref not a room")
	}
	log.Printf("Selected room: %s", room.ID.String())

	roomId := room.ID
	ui.state.CurrentRoom = roomId
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
	ui.grid.SetRows(-1, -5, -1)
	ui.grid.SetColumns(0, -3, 0)
	ui.grid.SetBorders(true)

	// ui.grid.AddItem(ui.Widgets[Header], 0, 0, 1, 3, 0, 0, true)
	ui.grid.AddItem(ui.Widgets[RoomList], 1, 0, 3, 1, 0, 0, true)
	ui.grid.AddItem(ui.Widgets[MessageList], 1, 1, 3, 3, 0, 0, true)

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
	log.Print("grid initialized")

}

func (u *UI) setSyncHandlers() {
	u.m.SetMessageHandler(u.state.handleMessageEvent)
	u.m.SetRoomNameHandler(u.state.handleRoomNameEvent)
	u.m.SetSyncCallback(u.state.toDisk)
}

func Start() {
	ui := New()
	ui.init()
	ui.setSyncHandlers()
	ui.Render()

	ui.m.Login()

	go ui.m.Sync()

	// go func(*UI) {
	// 	err := ui.app.Run()
	// 	if err != nil {
	// 		panic(err)
	// 	}
	// }(ui)

	err := ui.app.Run()
	if err != nil {
		panic(err)
	}

}
