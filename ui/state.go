package ui

import (
	"encoding/json"
	"os"
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
	ID       id.RoomID
	Name     string
	StateKey *string
}

type Message struct {
	*event.MessageEventContent
	Sender string
}

func NewState(ui *UI) *State {
	s := &State{
		Rooms:    make(map[id.RoomID]*Room),
		Messages: make(map[id.RoomID][]*Message),
		ui:       ui,
	}
	return s

}

func (s *State) handleMessageEvent(src mautrix.EventSource, e *event.Event) {
	s.Lock()
	defer s.Unlock()
	if _, ok := s.Rooms[e.RoomID]; ok {
		log.Debug("Room already processed")
	} else {
		s.Rooms[e.RoomID] = &Room{ID: e.RoomID, StateKey: e.StateKey}
	}

	sender := e.Sender.String()

	m := &Message{e.Content.AsMessage(), sender}
	s.Messages[e.RoomID] = append(s.Messages[e.RoomID], m)
	s.ui.Render()
	go s.toDisk()

}

func (s *State) handleRoomNameEvent(src mautrix.EventSource, e *event.Event) {
	s.Lock()
	defer s.Unlock()

	name := e.Content.AsRoomName().Name
	if r, ok := s.Rooms[e.RoomID]; ok {
		r.Name = name
	} else {
		s.Rooms[e.RoomID] = &Room{ID: e.RoomID, StateKey: e.StateKey, Name: name}
	}
	s.ui.Render()
	go s.toDisk()

	return
}

func (s *State) CurrentRoomMessages() []*Message {
	if s.CurrentRoom == "" {
		return []*Message{}
	}
	return s.Messages[s.CurrentRoom]

}

func (s *State) fromDisk() {
	// logger.Info("Loading state from disk")
	path := "state.json"
	f, err := os.Open(path)
	defer f.Close()
	if err != nil {
		if os.IsNotExist(err) {
			nf, err := os.Create(path)
			defer nf.Close()
			if err != nil {
				panic(err)
			}
			return
		}
	}
	dec := json.NewDecoder(f)
	err = dec.Decode(s)
	if err != nil {
		log.Info(err)
	}
	// logger.Info("State loaded from disk")

}

func (s *State) toDisk() {
	// logger.Info("Writing state to disk")
	f, err := os.Create("state.json")
	enc := json.NewEncoder(f)
	err = enc.Encode(s)
	if err != nil {
		panic(err)
	}
	f.Close()
	// logger.Info("Wrote state to disk")

}
