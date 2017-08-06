package main

import (
	"encoding/json"
	"flag"
	"log"
	"net/http"

	"gitlab.com/gammons/ultradeck-cli/ultradeck"

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
			s.removeConnection(conn)
			break
		}
		log.Printf("recv: '%s'", message)

		req := &ultradeck.Request{}
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

			req := &ultradeck.Request{}
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

func (s *Server) processRequest(request *ultradeck.Request) {
	switch request.Request {
	case ultradeck.AuthRequest:
		log.Println("Received auth request, awaiting response...")
	}
}

func (s *Server) writeResponse(request *ultradeck.Request) {
	log.Println("Writing response")
	conns := s.Connections[request.Data["token"].(string)]

	for _, conn := range conns {

		w, err := conn.NextWriter(websocket.TextMessage)
		if err != nil {
			log.Println("Error writing auth response: ", err)
			s.removeConnection(conn)
		} else {
			res := &ultradeck.Request{Request: ultradeck.AuthResponse, Data: request.Data}
			message, _ := json.Marshal(res)

			w.Write(message)
			w.Close()
		}
	}
}

func (s *Server) removeConnection(c *websocket.Conn) {
	var key string
	var indexToDelete int

	for k, connections := range s.Connections {
		for i, connection := range connections {
			if c == connection {
				log.Printf("Found bad connection at [%s][%v]", k, i)
				key = k
				indexToDelete = i
			}
		}
	}

	conns := s.Connections[key]
	s.Connections[key] = append(conns[:indexToDelete], conns[indexToDelete+1:]...)

	if len(s.Connections[key]) == 0 {
		delete(s.Connections, key)
	}
}
