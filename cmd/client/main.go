package main

// https://github.com/gorilla/websocket/blob/master/examples/echo/client.go

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/gammons/ultradeck-cli/client"
	"github.com/gammons/ultradeck-cli/ultradeck"
	"github.com/skratchdot/open-golang/open"
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
	// do I need this, with create?  not sure I do
	//case "init":

	// creates a new directory wioth a deck.md in it
	// also ties it to ultradeck.co with a .ud.yml file in it
	// also initializes git repo with a .gitignore?
	case "create":
		c.authorizedCommand(c.create)

	// pushes deck (and related assets) to ultradeck.co
	// ultradeck will check timestamp, and reject if timestamp on server is newer
	// can be forced with -f
	case "push":
		c.authorizedCommand(c.push)

	// pull deck (and related assets) from ultradeck.co
	// client will check timestamps and reject if client timestamp is newer
	// must be done PER FILE
	// can be forced with -f
	case "pull":
		c.authorizedCommand(c.pull)

	// watch a directory and auto-make changes on ultradeck's server
	// uses websocket connection and other cool shit to pull this off
	case "watch":
		c.authorizedCommand(c.watch)

	// check if logged in. internal for testing
	case "check":
		c.authorizedCommand(c.checkAuth)

	// check if logged in. internal for testing
	case "upgrade":
		c.authorizedCommand(c.upgradeToPaid)
	}
}

func (c *Client) doAuth() {
	c.Conn = &client.WebsocketConnection{}
	c.Conn.DoAuth(c.processAuthResponse)
}

func (c *Client) processAuthResponse(req *ultradeck.Request) {
	writer := client.NewAuthConfig(req.Data)
	writer.WriteAuth()
	c.Conn.CloseConnection()
}

func (c *Client) checkAuth(resp *client.AuthCheckResponse) {
	fmt.Printf("\nWelcome, %s! You're signed in.\n", resp.Name)
}

func (c *Client) upgradeToPaid(resp *client.AuthCheckResponse) {
	req, _ := http.NewRequest("GET", "http://localhost:3001/auth", nil)
	q := req.URL.Query()
	q.Add("username", resp.Username)
	q.Add("name", resp.Name)
	q.Add("token", resp.Token)
	q.Add("image_url", resp.ImageUrl)
	q.Add("email", resp.Email)
	q.Add("subscription_name", resp.SubscriptionName)
	q.Add("redirect", "/account")
	req.URL.RawQuery = q.Encode()

	fmt.Printf("\nSending you to the pricing page...")
	open.Start(req.URL.String())
}

type Deck struct {
	Title string `json:"title"`
}

type CreateDeck struct {
	Deck *Deck `json:"deck"`
}

func (c *Client) create(resp *client.AuthCheckResponse) {
	fmt.Println("What is the name of your deck?")
	reader := bufio.NewReader(os.Stdin)
	name, _ := reader.ReadString('\n')

	httpClient := client.NewHttpClient(resp.Token)
	createDeck := &CreateDeck{Deck: &Deck{Title: name}}
	j, _ := json.Marshal(&createDeck)
	jsonData := httpClient.PostRequest("api/v1/decks", j)

	if httpClient.Response.StatusCode == 200 {
		fmt.Println("Writing .ud.jsonNNNN")
		fmt.Println(string(jsonData[:]))

		deckConfigManager := &client.DeckConfigManager{}
		deckConfigManager.Write(jsonData)

		fmt.Println("Creating deck.md")

		// TODO create a simple deck.md file
		// markdownManager := &client.MarkdownManager{}
		// markdownManager.WriteFile()

	} else {
		fmt.Println("Something went wrong with the request:")
		if strings.Contains(string(jsonData), "There is a limit") {
			fmt.Println("You can only create 1 deck with a free account.")
			fmt.Println("Run `ultradeck upgrade` to upgrade your account!")
		} else {
			fmt.Println(string(jsonData))
		}
	}
}

func (c *Client) pull(resp *client.AuthCheckResponse) {
	deckConfigManager := &client.DeckConfigManager{}
	deckConfigManager.ReadConfig()

	if !deckConfigManager.FileExists() {
		fmt.Println("Could not find deck config!")
		fmt.Println("Did you run 'ultradeck create' yet?")
		return
	}

	httpClient := client.NewHttpClient(resp.Token)

	url := fmt.Sprintf("api/v1/decks/%d", deckConfigManager.GetDeckID())
	jsonData := httpClient.GetRequest(url)

	if httpClient.Response.StatusCode == 200 {

		var serverDeckConfig *client.DeckConfig
		_ = json.Unmarshal(jsonData, &serverDeckConfig)

		// date on server must be equal to or greater than date on client
		if c.dateCompare(serverDeckConfig.UpdatedAt, deckConfigManager.DeckConfig.UpdatedAt) >= 0 {
			fmt.Println("Pulling changes from ultradeck.co...")
			deckConfigManager.Write(jsonData)

			// pull remote assets as well
			fmt.Println("Syncing assets...")
			assetManager := client.AssetManager{}
			assetManager.PullRemoteAssets(serverDeckConfig)
			fmt.Println("Done!")
		} else {
			fmt.Println("It looks like you might have local changes that are not on the server!")
			fmt.Println("Did you make changes to your deck elsewhere, or on ultradeck.co?")
			fmt.Println("You can force by running 'ultradeck pull -f'.")
		}
	} else {
		fmt.Println("Something went wrong with the request:")
		fmt.Println(string(jsonData))
	}
}

func (c *Client) push(resp *client.AuthCheckResponse) {
	deckConfigManager := &client.DeckConfigManager{}
	deckConfigManager.ReadConfig()

	if !deckConfigManager.FileExists() {
		fmt.Println("Could not find deck config!")
		fmt.Println("Did you run 'ultradeck create' yet?")
		return
	}

	fmt.Println("Pushing local changes to ultradeck.co...")

	httpClient := client.NewHttpClient(resp.Token)

	// push local assets
	assetManager := client.AssetManager{}

	// TODO:  really not sure I like this type of decorator pattern
	// can I make it cleaner?
	deckConfigManager.DeckConfig = assetManager.PushLocalAssets(resp.Token, deckConfigManager.DeckConfig)

	url := fmt.Sprintf("api/v1/decks/%d", deckConfigManager.GetDeckID())
	jsonData := httpClient.PutRequest(url, deckConfigManager.PrepareJSONForUpload())

	if httpClient.Response.StatusCode == 200 {
		deckConfigManager.Write(jsonData)
		fmt.Println("Done!")
	} else {
		fmt.Println("Something went wrong with the request:")
		fmt.Println(string(jsonData))
	}
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

func (c *Client) watch(resp *client.AuthCheckResponse) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	fmt.Println("Watching directory for changes...")

	done := make(chan bool)
	go func() {
		for {
			select {
			case event := <-watcher.Events:
				if event.Name == ".ud.json" {
					continue
				}
				if event.Op == fsnotify.Write || event.Op == fsnotify.Create || event.Op == fsnotify.Remove {
					c.push(resp)
				}

			case err := <-watcher.Errors:
				log.Println("error:", err)
			}
		}
	}()

	err = watcher.Add(".")
	if err != nil {
		log.Fatal(err)
	}
	<-done
}

func (c *Client) dateCompare(d1 string, d2 string) int {
	t1, _ := time.Parse("2006-01-02T15:04:05.000Z", d1)
	t2, _ := time.Parse("2006-01-02T15:04:05.000Z", d2)

	if t1.Before(t2) {
		return -1
	}

	if t1.Equal(t2) {
		return 0
	}

	return 1
}
