package matrix

import (
	"encoding/json"
	"os"
	"sync"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"maunium.net/go/mautrix"
	"maunium.net/go/mautrix/event"
	"maunium.net/go/mautrix/id"
)

// DiskStore is a very simple mautrix.Storer implementation
// it is essentially the default InMemoryStore, but writes to disk
// after every write operation
type DiskStore struct {
	sync.Mutex
	path      string
	db        *gorm.DB
	Filters   map[id.UserID]string
	NextBatch map[id.UserID]string
	Rooms     map[id.RoomID]*mautrix.Room

	// not present in InMemoryStore, we want to cache message events
	// but need to figure best way
	Messages map[id.RoomID][]*Message
}

type Room struct {
	gorm.Model
	*mautrix.Room
}

type Message struct {
	gorm.Model
	*event.MessageEventContent
	Sender    string
	Timestamp int64
}

func (s *DiskStore) fromDisk() {
	// s.Lock()
	// defer s.Unlock()
	log.Print("Loading state from disk")
	path := "state.json"
	f, err := os.Open(path)
	defer f.Close()
	if err != nil {
		if os.IsNotExist(err) {
			log.Print("store cache file does not exist, creating")
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
		log.Print(err)
	}
	log.Print("State loaded from disk")
	log.Printf("%d Filters loaded from cache", len(s.Filters))
	log.Printf("%d NextBatch loaded from cache", len(s.NextBatch))
	log.Printf("%d rooms loaded from cache", len(s.Rooms))
	log.Printf("%d messages loaded from cache", len(s.Messages))

}

func (s *DiskStore) toDisk() {
	s.Lock()
	defer s.Unlock()
	log.Print("Writing state to disk")
	f, err := os.Create("state.json")
	defer f.Close()
	enc := json.NewEncoder(f)
	err = enc.Encode(s)
	if err != nil {
		panic(err)
	}
	log.Print("Wrote state to disk")

}

// SaveFilterID to memory.
func (s *DiskStore) SaveFilterID(userID id.UserID, filterID string) {
	log.Print("Saving filterID")
	s.Filters[userID] = filterID
	// s.toDisk()
}

// LoadFilterID from memory.
func (s *DiskStore) LoadFilterID(userID id.UserID) string {
	log.Print("Loading filterID")
	return s.Filters[userID]
}

// SaveNextBatch to memory.
func (s *DiskStore) SaveNextBatch(userID id.UserID, nextBatchToken string) {
	log.Print("Saving Next batch")
	s.NextBatch[userID] = nextBatchToken
	s.toDisk()
}

// LoadNextBatch from memory.
func (s *DiskStore) LoadNextBatch(userID id.UserID) string {
	log.Print("Loading Next batch")
	return s.NextBatch[userID]
}

// SaveRoom to memory.
func (s *DiskStore) SaveRoom(room *mautrix.Room) {
	log.Printf("Saving Room %s", room.ID.String())
	s.Rooms[room.ID] = room
	r := &Room{Room: room}
	s.db.Create(r)
}

// LoadRoom from memory.
func (s *DiskStore) LoadRoom(roomID id.RoomID) *mautrix.Room {
	log.Printf("Loading Room %s", roomID.String())
	return s.Rooms[roomID]
}

// SaveMessage to memory.
func (s *DiskStore) SaveMessage(roomID id.RoomID, m *Message) {
	log.Printf("Saving Message for room %s", roomID.String())
	_, exists := s.Messages[roomID]
	if !exists {
		log.Printf("handling first message for room %s", roomID.String())
		s.Messages[roomID] = make([]*Message, 0)
	}
	s.Messages[roomID] = append(s.Messages[roomID], m)

}

// LoadRoom from memory.
func (s *DiskStore) LoadMessages(roomID id.RoomID) []*Message {
	log.Printf("Loading Messages for Room %s", roomID.String())
	return s.Messages[roomID]
}

// UpdateState stores a state event. This can be passed to DefaultSyncer.OnEvent to keep all room state cached.
func (s *DiskStore) UpdateState(_ mautrix.EventSource, evt *event.Event) {
	if !evt.Type.IsState() {
		log.Printf("skipping store update for event type: %s", evt.Type.String())
		return
	}
	log.Printf("Updating DiskStore State with event: %s", evt.Type.String())
	room := s.LoadRoom(evt.RoomID)
	if room == nil {
		log.Print("Creating new room in store")
		room = mautrix.NewRoom(evt.RoomID)
		s.SaveRoom(room)
	}
	log.Print("Updating room state")
	room.UpdateState(evt)
	s.toDisk()

}

// NewInMemoryStore constructs a new InMemoryStore.
func newDiskStore() *DiskStore {

	s := &DiskStore{
		Filters:   make(map[id.UserID]string),
		NextBatch: make(map[id.UserID]string),
		Rooms:     make(map[id.RoomID]*mautrix.Room),
		Messages:  make(map[id.RoomID][]*Message, 1),
	}
	log.Print("Created diskstore")
	log.Printf("%+v", s.Messages)
	s.fromDisk()
	return s

}
