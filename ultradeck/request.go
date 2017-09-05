package ultradeck

const AuthRequest = "auth"
const AuthResponse = "auth_response"

type Request struct {
	Request string                 `json:"request"`
	Data    map[string]interface{} `json:"data"`
}
