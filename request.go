package ultradeckcli

import "github.com/twinj/uuid"

const AUTH_REQUEST = "auth"

type Request struct {
	Request string                 `json:"request"`
	Data    map[string]interface{} `json:"data"`
}

type AuthRequest struct {
	Token     uuid.Uuid `json:"token"`
	TokenType string    `json:"tokenType"`
}
