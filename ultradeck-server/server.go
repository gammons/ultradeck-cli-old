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

redis message:

json = {request: "auth_response", data: {intermediateToken: "e91ffb08-dbdf-4e9e-abd9-3eb59a2ca37b", token: "abcd1234"}}
r.rpush 'ultradeck', json.to_json

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

			s.writeResponse(req)
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
	switch request.Request {
	case ultradeckcli.AuthRequest:
		log.Println("Received auth request, awaiting response...")
	}
}

func (s *Server) writeResponse(request *ultradeckcli.Request) {
	log.Println("Writing response")
	conns := s.Connections[request.Data["token"].(string)]

	var keysToDelete []int
	for i, conn := range conns {

		w, err := conn.NextWriter(websocket.TextMessage)
		if err != nil {
			keysToDelete = append(keysToDelete, i)
			log.Println("Error writing auth response: ", err)
		} else {
			res := &ultradeckcli.Request{Request: ultradeckcli.AuthResponse, Data: request.Data}
			message, _ := json.Marshal(res)

			w.Write(message)
			w.Close()
		}
	}

	for _, i := range keysToDelete {
		log.Println("deleting closed connection at position ", i)
		s.Connections[request.Data["token"].(string)] = append(conns[:i], conns[i+1:]...)
	}
}

func (s *Server) deleteClosedConns() {

}
