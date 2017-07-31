package main

import (
	"flag"
	"log"
	"net/http"
	"strings"

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

	http.HandleFunc("/", server.HandleNewRequest)
	log.Println("Listening ln localhost:8080")
	log.Fatal(http.ListenAndServe(*addr, nil))
}

func (s *Server) HandleNewRequest(w http.ResponseWriter, r *http.Request) {
	s.serve(w, r)
}

func (s *Server) upgradeConnection(w http.ResponseWriter, r *http.Request) *websocket.Conn {
	var upgrader = websocket.Upgrader{}
	Conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return nil
	}
	defer Conn.Close()

	return Conn
}

func (s *Server) serve(w http.ResponseWriter, r *http.Request) {
	conn := s.upgradeConnection(w, r)
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Println("Read error: ", err)
			break
		}
		log.Printf("recv: '%s'", message)

		splitted := strings.Fields(string(message))
		switch splitted[0] {
		case "AUTH":
			s.performAuth(splitted)
		}
	}
}

func (s *Server) performAuth(message []string) {
	//s.Conn.WriteMessage(websocket.TextMessage)

}

// 	err = c.WriteMessage(mt, []byte("OK dingleberries"))
// 	if err != nil {
// 		log.Println("Write error: ", err)
// 	}
// }
