package ui

import (
	"sync"

	log "github.com/sirupsen/logrus"
	"maunium.net/go/mautrix"
	"maunium.net/go/mautrix/event"
	"maunium.net/go/mautrix/id"
)

type State struct {
	sync.Mutex
	Rooms       map[id.RoomID]*Room
	Messages    map[id.RoomID][]*Message
	ui          *UI
	CurrentRoom id.RoomID
}

type Room struct {
	id       id.RoomID
	alias    id.RoomAlias
	stateKey *string
}

type Message struct {
	*event.MessageEventContent
}

func NewState(ui *UI) *State {
	s := &State{
		Rooms:    make(map[id.RoomID]*Room),
		Messages: make(map[id.RoomID][]*Message),
		ui:       ui,
	}
	return s

}

func (s *State) ProcessMessage(src mautrix.EventSource, e *event.Event) {
	s.Lock()
	defer s.Unlock()

	if _, ok := s.Rooms[e.RoomID]; ok {
		log.Debug("Room already processed")
	} else {
		s.Rooms[e.RoomID] = &Room{id: e.RoomID, stateKey: e.StateKey}
	}

	m := &Message{e.Content.AsMessage()}
	s.Messages[e.RoomID] = append(s.Messages[e.RoomID], m)
	s.ui.Render()
}

func (s *State) CurrentRoomMessages() []*Message {
	if s.CurrentRoom == "" {
		return []*Message{}
	}
	return s.Messages[s.CurrentRoom]

}
