package main

// https://github.com/gorilla/websocket/blob/master/examples/echo/client.go

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"

	"github.com/gammons/ultradeck-cli/client"
	"github.com/gammons/ultradeck-cli/ultradeck"
)

type Client struct {
	Conn *client.WebsocketConnection
}

func main() {
	c := &Client{}

	switch os.Args[1] {
	case "auth":
		c.doAuth()

	// initialize an existing markdown file to be connected with ultradeck.co
	case "init":

	// creates a new directory wioth a deck.md in it
	// also ties it to ultradeck.co with a .ud.yml file in it
	// also initializes git repo with a .gitignore?
	case "create":
		c.authorizedCommand(c.create)

	// pushes deck (and related assets) to ultradeck.co
	// ultradeck will check timestamp, and reject if timestamp on server is newer
	// can be forced with -f
	case "push":

	// pull deck (and related assets) from ultradeck.co
	// client will check timestamps and reject if client timestamp is newer
	// must be done PER FILE
	// can be forced with -f
	case "pull":

	// watch a directory and auto-make changes on ultradeck's server
	// uses websocket connection and other cool shit to pull this off
	case "watch":

	// check if logged in. internal for testing
	case "check":
		c.authorizedCommand(c.checkAuth)
	}
}

func (c *Client) doAuth() {
	c.Conn = &client.WebsocketConnection{}
	c.Conn.DoAuth(c.processAuthResponse)
}

func (c *Client) processAuthResponse(req *ultradeck.Request) {
	writer := client.NewAuthConfig(req.Data["access_token"].(string))
	writer.WriteAuth()
	c.Conn.CloseConnection()
}

func (c *Client) checkAuth(resp *client.AuthCheckResponse) {
	fmt.Printf("\nWelcome, %s! You're signed in.\n", resp.Name)
}

type Deck struct {
	Title string `json:"title"`
}

type CreateDeck struct {
	Deck *Deck `json:"deck"`
}

func (c *Client) create(resp *client.AuthCheckResponse) {
	reader := bufio.NewReader(os.Stdin)
	fmt.Println("What is the name of your deck?")
	name, _ := reader.ReadString('\n')
	fmt.Println(name)

	httpClient := client.NewHttpClient(resp.Token)
	createDeck := &CreateDeck{Deck: &Deck{Title: name}}
	j, _ := json.Marshal(&createDeck)
	bodyBytes := httpClient.PostRequest("api/v1/decks", j)
	fmt.Println(string(bodyBytes))
}

func (c *Client) authorizedCommand(cmd func(resp *client.AuthCheckResponse)) {
	authConfig := &client.AuthConfig{}
	if authConfig.AuthFileExists() {
		token := authConfig.GetToken()

		authCheck := &client.AuthCheck{}
		resp := authCheck.CheckAuth(token)
		resp.Token = token

		if resp.IsSignedIn {
			cmd(resp)
		} else {
			fmt.Println("\nIt does not look like you're signed in anymore.")
			fmt.Println("Please run 'ultradeck auth' to sign in again.")
		}
	} else {
		fmt.Println("\nNo auth config file found!")
		fmt.Println("Please run 'ultradeck auth' to log in.")
	}
}
