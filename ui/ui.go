package ui

import (
	"os"
	"os/signal"
	"syscall"

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

	input := NewInputWidget()
	ui.Widgets[Input] = input
	input.SetDoneFunc(func(key tcell.Key) {
		room := ui.state.CurrentRoom
		roomName := ui.state.Rooms[room].Name
		ui.m.SendMessage(roomName, room, input.GetText())
		input.SetText("")
	})

	status := NewStatusWidget()
	ui.Widgets[Status] = status

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

	ui.grid.AddItem(ui.Widgets[Status], 0, 2, 1, 3, 0, 0, true)
	ui.grid.AddItem(ui.Widgets[RoomList], 1, 0, 3, 1, 0, 0, true)
	ui.grid.AddItem(ui.Widgets[MessageList], 1, 1, 2, 3, 0, 0, true)
	ui.grid.AddItem(ui.Widgets[Input], 3, 1, 1, 3, 0, 0, true)

	ui.app.SetRoot(ui.grid, true)
	ui.app.QueueUpdateDraw(func() {})

	ui.app.SetFocus(ui.Widgets[RoomList])
	ui.currentwidget = ui.Widgets[RoomList]

	ui.app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {

		key := event.Key()
		switch key {
		case tcell.KeyTab:
			ui.toggleFocus()
			return nil

		case tcell.KeyEscape:
			ui.app.SetFocus(ui.Widgets[RoomList])
			ui.currentwidget = ui.Widgets[RoomList]
			return nil

		case tcell.KeyRune:
			switch event.Rune() {
			case 'i': // Home.
				ui.app.SetFocus(ui.Widgets[Input])
				ui.currentwidget = ui.Widgets[Input]
				return nil
			}

			return event
		}

		return event
	})
	log.Print("grid initialized")

}

func (u *UI) setSyncHandlers() {
	u.m.SetMessageHandler(u.state.handleMessageEvent)
	u.m.SetRoomNameHandler(u.state.handleRoomNameEvent)
}

func Start() {
	ui := New()
	ui.init()
	ui.setSyncHandlers()
	ui.Render()

	ui.m.Login()
	ui.state.loadFromCache(ui.m.Store())

	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func(ui *UI) {
		<-c
		log.Print("program interrupted, saving state")
		ui.m.CacheState()
		os.Exit(1)
	}(ui)

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
