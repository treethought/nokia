package matrix

import (
	"github.com/spf13/viper"
	"github.com/treethought/nokia/logger"
	"maunium.net/go/mautrix"
	"maunium.net/go/mautrix/event"
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
}

func (c *Client) Sync() error {
	log.Print("starting sync")
	err := c.m.Sync()
	if err != nil {
		log.Fatal(err)
	}
	log.Print("Sync complete, calling callback")
	go c.syncCallback()

	return nil
}

func (c *Client) SetSyncCallback(f func()) {
	c.syncCallback = f
}

func (c *Client) SetMessageHandler(handler mautrix.EventHandler) {
	c.messageHandler = handler

	syncer := c.m.Syncer.(*mautrix.DefaultSyncer)
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

}

func (c *Client) ListRooms() ([]string, error) {
	resp, err := c.m.JoinedRooms()
	if err != nil {
		return nil, err
	}
	rooms := []string{}
	for _, r := range resp.JoinedRooms {
		rooms = append(rooms, r.String())
	}
	return rooms, nil
}
