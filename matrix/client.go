package matrix

import (
	"github.com/spf13/viper"
	"github.com/treethought/nokia/logger"
	"maunium.net/go/mautrix"
	"maunium.net/go/mautrix/event"
	"maunium.net/go/mautrix/id"
)

var log = logger.GetLoggerInstance()

type Client struct {
	m *mautrix.Client
	// syncCallback    func()
	messageHandler  mautrix.EventHandler
	roomNameHandler mautrix.EventHandler
}

func New() (c *Client, err error) {
	c = &Client{}
	homeserver := viper.GetString("homeserver")
	c.m, err = mautrix.NewClient(homeserver, "", "")
	c.m.Store = newDiskStore()
	// c.setInternalHandlers()
	c.setInternalEventHandler()
	return
}

func (c *Client) Login() {
	homeserver := viper.GetString("homeserver")
	username := viper.GetString("user")
	password := viper.GetString("password")

	log.Printf("Logging into %s as %s", homeserver, username)
	_, err := c.m.Login(&mautrix.ReqLogin{
		Type:             "m.login.password",
		Identifier:       mautrix.UserIdentifier{Type: mautrix.IdentifierTypeUser, User: username},
		Password:         password,
		StoreCredentials: true,
	})
	if err != nil {
		log.Fatalf("Failed to login to homeserver %s: %v", homeserver, err)
	}
	log.Print("Successfully logged in")
	// c.loadJoinedRooms()
}

func (c *Client) Sync() error {
	log.Print("starting sync")
	err := c.m.Sync()
	if err != nil {
		log.Fatal(err)
	}
	log.Print("Sync complete, calling callback")
	// go c.syncCallback()

	return nil
}

func (c *Client) internalRoomNameHandler(_ mautrix.EventSource, evt *event.Event) {
	log.Print("Handling room name event with internal handler")
	// name := evt.Content.AsRoomName().Name
	room := mautrix.NewRoom(evt.RoomID)
	room.UpdateState(evt)
	c.m.Store.SaveRoom(room)
}

func (c *Client) internalMessageHandler(_ mautrix.EventSource, evt *event.Event) {
	log.Print("MESSAGE EVENT")
	log.Print("Handling message event with internal handler")
	ds, ok := c.m.Store.(*DiskStore)
	if !ok {
		return
	}
	sender := evt.Sender.String()
	ts := evt.Timestamp

	m := &Message{
		MessageEventContent: evt.Content.AsMessage(),
		Sender:              sender,
		Timestamp:           ts,
	}
	ds.SaveMessage(evt.RoomID, m)
}

func (c *Client) setInternalHandlers() {
	_, ok := c.m.Store.(*DiskStore)
	if !ok {
		return
	}
	log.Print("setting internal event handlers for DiskStore")
	syncer := c.m.Syncer.(*mautrix.DefaultSyncer)
	syncer.OnEventType(event.StateRoomName, c.internalRoomNameHandler)
	syncer.OnEventType(event.EventMessage, c.internalMessageHandler)

}

func (c *Client) setInternalEventHandler() {
	ds, ok := c.m.Store.(*DiskStore)
	if !ok {
		return
	}
	log.Print("Setting event handler to cache with diskStore")

	syncer := c.m.Syncer.(*mautrix.DefaultSyncer)
	syncer.OnEvent(ds.UpdateState)
}

func (c *Client) CacheState() {
	ds, ok := c.m.Store.(*DiskStore)
	if ok {
		ds.toDisk()

	}
}

func (c *Client) SetMessageHandler(handler mautrix.EventHandler) {
	c.messageHandler = handler

	syncer := c.m.Syncer.(*mautrix.DefaultSyncer)
	log.Print("SETTING EXTERNAL MESSAGE HANDLER")
	syncer.OnEventType(event.EventMessage, c.messageHandler)

}

func (c *Client) SetRoomNameHandler(handler mautrix.EventHandler) {
	c.roomNameHandler = handler
	syncer := c.m.Syncer.(*mautrix.DefaultSyncer)
	syncer.OnEventType(event.StateRoomName, c.roomNameHandler)
}

func (c *Client) SendMessage(roomName string, room id.RoomID, text string) {
	log.Printf("Sending message to room %s:\n%s", roomName, text)
	resp, err := c.m.SendText(room, text)
	if err != nil {
		log.Fatalf("Failed to send message %v", err)
	}
	log.Printf("Send message with event id: %s", resp.EventID.String())

}

func (c *Client) ProcessEvent(roomID id.RoomID, eventID id.EventID) error {
	log.Print("Processing event outside of sync process")
	e, err := c.m.GetEvent(roomID, eventID)
	if err != nil {
		return err
	}

	// TODO either process event without use of source
	// or detemrine which type it should be
	// we simply want to include the new sent message into
	// the systems (via the handler func) collection of message
	src := mautrix.EventSource(0)

	go c.messageHandler(src, e)
	return nil

}

func (c *Client) Store() mautrix.Storer {
	return c.m.Store
}

func (c *Client) isDiskStore() bool {
	_, ok := c.m.Store.(*DiskStore)
	return ok

}

func (c *Client) SaveRoom(room *mautrix.Room) {
	c.m.Store.SaveRoom(room)
}

func (c *Client) loadJoinedRooms() ([]string, error) {
	resp, err := c.m.JoinedRooms()
	if err != nil {
		return nil, err
	}
	rooms := []string{}
	for _, r := range resp.JoinedRooms {
		rooms = append(rooms, r.String())
		cache := c.m.Store.LoadRoom(r)
		if cache == nil {
			cache = mautrix.NewRoom(r)
		}
		c.m.Store.SaveRoom(cache)
	}
	return rooms, nil
}

// func (c *Client) Rooms() []*mautrix.Room {
// 	return c.m.Store.
// }
