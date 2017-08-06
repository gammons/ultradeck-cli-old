package main

// https://github.com/gorilla/websocket/blob/master/examples/echo/client.go

import (
	"os"

	"gitlab.com/gammons/ultradeck-cli/client"
	"gitlab.com/gammons/ultradeck-cli/ultradeck"
)

type Client struct {
	Conn *client.WebsocketConnection
}

func main() {
	client := &Client{}

	switch os.Args[1] {
	case "auth":
		client.doAuth()
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
