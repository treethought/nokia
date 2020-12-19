package ui

import (
	"sync"

	"github.com/treethought/nokia/matrix"
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
	ID       id.RoomID
	Name     string
	StateKey *string
}

type Message struct {
	*event.MessageEventContent
	Sender    string
	Timestamp int64
}

func NewState(ui *UI) *State {
	s := &State{
		Rooms:    make(map[id.RoomID]*Room),
		Messages: make(map[id.RoomID][]*Message),
		ui:       ui,
	}
	return s

}

func (s *State) Store() mautrix.Storer {
	return s.ui.m.Store()
}

func (s *State) loadFromCache(store mautrix.Storer) {
	s.Lock()
	defer s.Unlock()
	ds, ok := store.(*matrix.DiskStore)
	if !ok {
		return
	}
	log.Print("Loading state from cache")

	for i := range ds.Rooms {
		s.Rooms[i] = &Room{ID: i, Name: i.String()}
		cacheMsgs := ds.Messages[i]
		for _, m := range cacheMsgs {
			uiMsg := &Message{
				MessageEventContent: m.MessageEventContent,
				Sender:              m.Sender,
				Timestamp:           m.Timestamp,
			}
			s.Messages[i] = append(s.Messages[i], uiMsg)
		}
		log.Printf("Loaded room %s state from cache", i)

	}

}

func (s *State) handleMessageEvent(src mautrix.EventSource, e *event.Event) {
	s.Lock()
	defer s.Unlock()
	log.Print(("HANDLING MESSAGE IN STATE"))
	if _, ok := s.Rooms[e.RoomID]; !ok {
		s.Rooms[e.RoomID] = &Room{ID: e.RoomID, StateKey: e.StateKey}
		log.Printf("handling first message for room %s", e.RoomID.String())
	}

	sender := e.Sender.String()
	ts := e.Timestamp

	m := &Message{MessageEventContent: e.Content.AsMessage()}
	m.Sender = sender
	m.Timestamp = ts
	s.Messages[e.RoomID] = append(s.Messages[e.RoomID], m)
	s.ui.Render()

}

func (s *State) handleRoomNameEvent(src mautrix.EventSource, e *event.Event) {
	log.Print("Handling StateRoomName event")
	s.Lock()
	defer s.Unlock()

	name := e.Content.AsRoomName().Name
	if r, ok := s.Rooms[e.RoomID]; ok {
		r.Name = name
	} else {
		s.Rooms[e.RoomID] = &Room{ID: e.RoomID, StateKey: e.StateKey, Name: name}
	}
	s.ui.Render()

	return
}

func (s *State) CurrentRoomMessages() []*Message {
	if s.CurrentRoom == "" {
		return []*Message{}
	}
	return s.Messages[s.CurrentRoom]

}
