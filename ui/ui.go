package ui

import (
	"github.com/gdamore/tcell/v2"
	"github.com/treethought/nokia/logger"
	"github.com/treethought/nokia/matrix"
	"gitlab.com/tslocum/cview"
)

type View string

var log = logger.GetLoggerInstance()

const (
	RoomList    View = "rooms"
	MessageList View = "main"
	Status      View = "status"
	Input       View = "input"
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
	// ui.initPanels()
	log.Print("UI intiialized")
}

func (ui *UI) Render() {
	for _, w := range ui.Widgets {
		go w.Render(ui)
	}

}

func (ui *UI) initWidgets() {

	status := NewStatusWidget()
	ui.Widgets[Status] = status

	rooms := NewRoomsWidget()
	rooms.SetSelectHandler(ui.roomSelectHandler)
	ui.Widgets[RoomList] = rooms

	msgs := NewMessagesWidget()
	ui.Widgets[MessageList] = msgs
	log.Print("widgets initialized")

	input := NewInputBox()
	ui.Widgets[Input] = input

}

func (ui *UI) roomSelectHandler(item *cview.ListItem) {
	ref := item.GetReference()

	room, ok := ref.(Room)
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

	ui.grid.AddItem(ui.Widgets[RoomList], 0, 0, 3, 1, 0, 0, true)
	ui.grid.AddItem(ui.Widgets[MessageList], 0, 1, 3, 3, 0, 0, true)
	ui.grid.AddItem(ui.Widgets[Status], 3, 0, 1, 1, 0, 0, true)
	ui.grid.AddItem(ui.Widgets[Input], 3, 1, 1, 2, 0, 0, true)

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

func (ui *UI) initPanels() {
	panels := cview.NewPanels()
	panels.AddPanel("messages", ui.Widgets[MessageList], true, true)
	// panels.AddPanel("thread", app.thread, true, true)
	// panels.AddPanel("compose", app.compose, true, true)
	panels.SetCurrentPanel("messages")
	// app.panels = panels

	mid := cview.NewFlex()
	mid.SetBackgroundColor(tcell.ColorDefault)
	mid.SetDirection(cview.FlexRow)
	mid.AddItem(panels, 0, 4, true)
	// mid.AddItem(app.statusView, 0, 4, false)
	mid.AddItem(ui.Widgets[Input], 0, 1, false)

	flex := cview.NewFlex()
	flex.SetBackgroundTransparent(false)
	flex.SetBackgroundColor(tcell.ColorDefault)

	left := cview.NewFlex()
	left.SetDirection(cview.FlexRow)
	left.AddItem(ui.Widgets[RoomList], 0, 7, false)
	left.AddItem(ui.Widgets[Status], 0, 1, false)
	// left.AddItem(acctInfo, 0, 1, false)

	flex.AddItem(left, 0, 1, false)
	flex.AddItem(mid, 0, 4, false)

	ui.app.SetRoot(flex, true)
	ui.app.QueueUpdateDraw(func() {})

	ui.app.SetFocus(ui.Widgets[RoomList])
	ui.currentwidget = ui.Widgets[RoomList]
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

	err := ui.app.Run()
	if err != nil {
		panic(err)
	}
	ui.state.toDisk()
}
