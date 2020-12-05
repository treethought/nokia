package matrix

import (
	"github.com/spf13/viper"
	"github.com/treethought/nokia/logger"
	"maunium.net/go/mautrix"
	"maunium.net/go/mautrix/event"
)

var log = logger.GetLoggerInstance()

type Client struct {
	m            *mautrix.Client
	syncCallback func()
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
	log.Print("syncing....\n")
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
	syncer := c.m.Syncer.(*mautrix.DefaultSyncer)
	syncer.OnEventType(event.EventMessage, handler)
}

func (c *Client) SetRoomNameHandler(handler mautrix.EventHandler) {
	syncer := c.m.Syncer.(*mautrix.DefaultSyncer)
	syncer.OnEventType(event.StateRoomName, handler)
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
