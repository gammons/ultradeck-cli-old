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
	Connections map[string][]*websocket.Conn
	RedisConn   *redis.Client
}

/*

rails only has a connection to redis.  goals is not to have rails need websockets.

assuming websocket server shares redis with rails

Auth flow:

1. client calls auth request to server with intermediate token
2. websocket server stores connection in connections map with intermediate token, for later use
2. client will call UD backend auth with the intermediate token from the client
3. rails stores intermediate token in session
4. oauth is performed
5. on oauth callback, rails pushes a message for the intermediate key, and gives word of the good news, sending back access token
6. websocket server forwards access token to client and closes connection
7. client writes ~/.config/ultradeck.json

*/

/*

result from rails app:

{ response: "auth_response", intermediate_token: "abcd1234", token: "<jwt token>" }


*/

func main() {
	log.SetFlags(0)

	server := &Server{}
	server.Connections = make(map[string][]*websocket.Conn)
	server.SetupRedisListener()

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

		connectionsForToken := s.Connections[req.Data["token"].(string)]
		connectionsForToken = append(connectionsForToken, conn)
		s.Connections[req.Data["token"].(string)] = connectionsForToken

		s.processRequest(req)
	}
}

func (s *Server) SetupRedisListener() {
	log.Println("Connecting to redis...")
	s.RedisConn = redis.NewClient(&redis.Options{
		Addr:     *redisAddr,
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	_, err := s.RedisConn.Ping().Result()
	if err != nil {
		log.Println("Error connecting to redis: ", err)
		panic(err)
	}

	go func() {
		for {
			result, err := s.RedisConn.BLPop(0, "ultradeck").Result()
			if err != nil {
				panic(err)
			}

			log.Println("redis result is ", result[1])

			req := &ultradeckcli.Request{}
			err = json.Unmarshal([]byte(result[1]), req)
			if err != nil {
				panic(err)
			}

			s.processRequest(req)
		}
	}()
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

func (s *Server) processRequest(request *ultradeckcli.Request) {
	log.Println("Processing request ", request)
	switch request.Request {
	// case ultradeckcli.AuthRequest: //server does not need to explicitly do anything while waiting for oauth
	// 	s.performAuth(request)
	case ultradeckcli.AuthResponse:
		s.performAuthResponse(request)
	}
}

func (s *Server) performAuthResponse(request *ultradeckcli.Request) {

	conns := s.Connections[request.Data["intermediateToken"].(string)]

	for _, conn := range conns {

		w, err := conn.NextWriter(websocket.TextMessage)
		if err != nil {
			// TODO: when we error here, that means the client connection has gone away.
			// we will need to remove this connection from the Connections list.
			log.Println("Error writing auth response: ", err)
			return
		}

		res := &ultradeckcli.Request{Request: ultradeckcli.AuthResponse, Data: request.Data}
		message, _ := json.Marshal(res)

		w.Write(message)
		w.Close()
	}
}
