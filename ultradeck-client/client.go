package main

// https://github.com/gorilla/websocket/blob/master/examples/echo/client.go

import (
	"fmt"
	"log"
	"net/url"
	"os"
	"os/signal"

	"github.com/gorilla/websocket"
	"github.com/twinj/uuid"
)

// Client is the client
type Client struct {
	Conn      *websocket.Conn
	Done      chan struct{}
	Interrupt chan os.Signal
}

func main() {
	client := &Client{}
	switch os.Args[1] {
	case "auth":
		client.DoAuth()
	}
}

// DoAuth does auth
func (c *Client) DoAuth() {
	c.openConnection()

	msg := fmt.Sprintf("AUTH %s", uuid.NewV4())
	err := c.Conn.WriteMessage(websocket.TextMessage, []byte(msg))
	if err != nil {
		log.Println("write err: ", err)
	}
}

func (c *Client) openConnection() {
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
	defer c.Conn.Close()

	c.Done = make(chan struct{})

	c.setupMessageReader()

	select {
	case <-c.Done:
		log.Println("Got done msg")
	case <-c.Interrupt:
		c.closeConnection()
		log.Println("interrupt")
	}
}

func (c *Client) setupMessageReader() {
	go func() {
		log.Println("In setupMessageReader")
		defer c.Conn.Close()
		defer close(c.Done)
		for {
			_, message, err := c.Conn.ReadMessage()
			if err != nil {
				log.Println("read error:", err)
				return
			}
			c.processMessage(string(message))
		}
	}()

}

func (c *Client) closeConnection() {
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

func (c *Client) processMessage(message string) {
	switch message {
	case "DO-OAUTH":

	case "DONE":
		c.closeConnection()
	}
}
