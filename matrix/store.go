package matrix

import (
	"encoding/json"
	"os"
	"sync"

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
	Filters   map[id.UserID]string
	NextBatch map[id.UserID]string
	Rooms     map[id.RoomID]*mautrix.Room

	// not present in InMemoryStore, we want to cache message events
	// but need to figure best way
	Messages map[id.RoomID]*event.Event
}

func (s *DiskStore) fromDisk() {
	log.Print("Loading state from disk")
	path := "state.json"
	f, err := os.Open(path)
	defer f.Close()
	if err != nil {
		if os.IsNotExist(err) {
			log.Print("store cache file does not exiswt, creating")
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

}

func (s *DiskStore) toDisk() {
	log.Print("Writing state to disk")
	f, err := os.Create("state.json")
	enc := json.NewEncoder(f)
	err = enc.Encode(s)
	if err != nil {
		panic(err)
	}
	f.Close()
	log.Print("Wrote state to disk")

}

// SaveFilterID to memory.
func (s *DiskStore) SaveFilterID(userID id.UserID, filterID string) {
	s.Lock()
	defer s.Unlock()
	log.Print("Saving filterID")
	s.Filters[userID] = filterID
	s.toDisk()
}

// LoadFilterID from memory.
func (s *DiskStore) LoadFilterID(userID id.UserID) string {
	log.Print("Loading filterID")
	return s.Filters[userID]
}

// SaveNextBatch to memory.
func (s *DiskStore) SaveNextBatch(userID id.UserID, nextBatchToken string) {
	s.Lock()
	defer s.Unlock()
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
	s.Lock()
	defer s.Unlock()
	log.Printf("Saving Room %s", room.ID.String())
	s.Rooms[room.ID] = room
	s.toDisk()
}

// LoadRoom from memory.
func (s *DiskStore) LoadRoom(roomID id.RoomID) *mautrix.Room {
	log.Printf("Loading Room %s", roomID.String())
	return s.Rooms[roomID]
}

// UpdateState stores a state event. This can be passed to DefaultSyncer.OnEvent to keep all room state cached.
func (s *DiskStore) UpdateState(_ mautrix.EventSource, evt *event.Event) {
	s.Lock()
	defer s.Unlock()
	if !evt.Type.IsState() {
		return
	}
	room := s.LoadRoom(evt.RoomID)
	if room == nil {
		room = mautrix.NewRoom(evt.RoomID)
		s.SaveRoom(room)
	}
	room.UpdateState(evt)
	s.toDisk()

}

// NewInMemoryStore constructs a new InMemoryStore.
func newDiskStore() *DiskStore {
	s := &DiskStore{
		Filters:   make(map[id.UserID]string),
		NextBatch: make(map[id.UserID]string),
		Rooms:     make(map[id.RoomID]*mautrix.Room),
	}
	s.fromDisk()
	return s

}
