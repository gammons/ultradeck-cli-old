package main

import (
	"encoding/json"
	"flag"
	"log"
	"net/http"

	ultradeckcli "gitlab.com/gammons/ultradeck-cli"

	"github.com/go-redis/redis"
	"github.com/gorilla/websocket"
)

var addr = flag.String("addr", "localhost:8080", "http service address")
var redisAddr = flag.String("redisAddr", "localhost:6379", "redis address")
var ultradeckBackendAddr = flag.String("ultradeckBackendAddr", "localhost:3000", "ultradeck backend address")

// Server is server stuff
type Server struct {
	Connections map[string]*websocket.Conn
	RedisConn   *redis.Client
}

/*

rails only has a connection to redis.  goals is not to have rails need websockets.

1. websocket server shares redis with rails
2. auth request will call UD backend auth with the intermediate key from the client
3. rails stores intermediate key in session
4. oauth is performed
5. on oauth callback, rails pushes a message for the intermediate key, and gives word of the good news, sending back access token
6. websocket server forwards access token to client and closes connection
7. websocket server also stores connection in a set for future use
7. client writes ~/.config/ultradeck.json

*/

func main() {
	log.SetFlags(0)

	server := &Server{}
	server.RedisConn = redis.NewClient(&redis.Options{
		Addr:     *redisAddr,
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	http.HandleFunc("/", server.Serve)
	log.Println("Listening ln localhost:8080")
	log.Fatal(http.ListenAndServe(*addr, nil))
}

func (s *Server) Serve(w http.ResponseWriter, r *http.Request) {
	conn := s.upgradeConnection(w, r)
	defer conn.Close()

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Println("Read error: ", err)
			break
		}
		log.Printf("recv: '%s'", message)

		req := &ultradeckcli.Request{}
		json.Unmarshal(message, req)

		switch req.Request {
		case ultradeckcli.AUTH_REQUEST:
			s.performAuth(*req)
		}
	}
}

func (s *Server) upgradeConnection(w http.ResponseWriter, r *http.Request) *websocket.Conn {
	var upgrader = websocket.Upgrader{}
	Conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade error:", err)
		return nil
	}

	return Conn
}

func (s *Server) performAuth(request ultradeckcli.Request) {
	log.Println("in performAuth", request.Data["tokenType"])
	//s.Conn.WriteMessage(websocket.TextMessage)
}
