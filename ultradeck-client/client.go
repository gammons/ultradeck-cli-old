package main

// https://github.com/gorilla/websocket/blob/master/examples/echo/client.go

import (
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"os"
	"os/signal"

	ultradeckcli "gitlab.com/gammons/ultradeck-cli"

	"github.com/gorilla/websocket"
	"github.com/skratchdot/open-golang/open"
	"github.com/twinj/uuid"
)

// Client is the client
type Client struct {
	Conn      *websocket.Conn
	Done      chan struct{}
	Interrupt chan os.Signal
}

/*

{ request: "auth", token: "abcd1234", tokenType: "intermediate" }

*/

func main() {
	client := &Client{}
	client.OpenConnection()

	switch os.Args[1] {
	case "auth":
		client.DoAuth(string(uuid.NewV4()))
	case "testauth":
		client.DoAuth("abcd1234")
	}
}

func (c *Client) OpenConnection() {
	c.Interrupt = make(chan os.Signal, 1)
	signal.Notify(c.Interrupt, os.Interrupt)

	log.Printf("connecting to %s", c.serverURL())

	var err error
	log.Println("Dialing...")
	c.Conn, _, err = websocket.DefaultDialer.Dial(c.serverURL(), nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	log.Println("Dialed")
	// defer c.Conn.Close()

	c.Done = make(chan struct{})

}

// DoAuth does auth
func (c *Client) DoAuth(token string) {

	var auth = make(map[string]interface{})
	auth["token"] = token
	auth["tokenType"] = "intermediate"

	req := &ultradeckcli.Request{Request: ultradeckcli.AuthRequest, Data: auth}
	authMsg, _ := json.Marshal(req)
	log.Printf("authMsg = %s", auth)

	err := c.Conn.WriteMessage(websocket.TextMessage, []byte(authMsg))
	if err != nil {
		log.Println("write err: ", err)
	}

	url := fmt.Sprintf("http://localhost:3000/start_auth?token=%s", auth["token"])
	open.Start(url)

	c.listen()
}

func (c *Client) listen() {
	go func() {
		log.Println("Listening..")
		defer c.Conn.Close()
		defer close(c.Done)
		for {
			_, message, err := c.Conn.ReadMessage()
			if err != nil {
				log.Println("read error:", err)
				return
			}

			req := &ultradeckcli.Request{}
			json.Unmarshal(message, req)
			c.processMessage(req)
		}
	}()

	log.Println("after setupMessageReader")

	select {
	case <-c.Done:
		log.Println("Got done msg")
	case <-c.Interrupt:
		c.closeConnection()
		log.Println("interrupt")
	}
}

func (c *Client) closeConnection() {
	close(c.Done)
	err := c.Conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	if err != nil {
		log.Println("write close err:", err)
		return
	}
	c.Conn.Close()
}

func (c *Client) serverURL() string {
	addr := "localhost:8080"
	u := url.URL{Scheme: "ws", Host: addr, Path: "/"}
	return u.String()
}

func (c *Client) processMessage(req *ultradeckcli.Request) {
	log.Println("in processMessage with ", req)
	switch req.Request {
	case ultradeckcli.AuthResponse:
		c.processAuthResponse(req)
	}
}

func (c *Client) processAuthResponse(req *ultradeckcli.Request) {
	log.Println("in processAuthResponse with ", req)
	log.Println("closing connection")
	c.closeConnection()
}
