package ui

import (
	"fmt"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/spf13/viper"
	"gitlab.com/tslocum/cview"
)

type StatusWidget struct {
	*cview.TextView
}

func NewStatusWidget() *StatusWidget {
	return &StatusWidget{TextView: cview.NewTextView()}
}

func (w *StatusWidget) Render(ui *UI) error {
	w.Clear()

	homeserver := viper.GetString("homeserver")
	username := viper.GetString("user")

	hostdomain := strings.Split(homeserver, "://")[1]

	s := fmt.Sprintf("username: @%s:%s\nhomerserver: %s", username, hostdomain, homeserver)

	w.SetText(s)
	w.SetBackgroundColor(tcell.ColorDefault)
	return nil

}
