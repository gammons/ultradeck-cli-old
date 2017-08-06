package main

// https://github.com/gorilla/websocket/blob/master/examples/echo/client.go

import (
	"os"

	"github.com/gammons/ultradeck-cli/client"
	"github.com/gammons/ultradeck-cli/ultradeck"
)

type Client struct {
	Conn *client.WebsocketConnection
}

func main() {
	client := &Client{}

	switch os.Args[1] {
	case "auth":
		client.doAuth()

	// initialize an existing markdown file to be connected with ultradeck.co
	case "init":

	// creates a new directory wioth a deck.md in it
	// also ties it to ultradeck.co with a .ud.yml file in it
	// also initializes git repo with a .gitignore?
	case "create":
	}
}

func (c *Client) doAuth() {
	c.Conn = &client.WebsocketConnection{}
	c.Conn.DoAuth(c.processAuthResponse)
}

func (c *Client) processAuthResponse(req *ultradeck.Request) {
	writer := client.NewAuthConfigWriter(req.Data["access_token"].(string))
	writer.WriteAuth()
	c.Conn.CloseConnection()
}
