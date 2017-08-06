package ultradeck

const AuthRequest = "auth"
const AuthResponse = "auth_response"
const PushRequest = "push"
const PullRequest = "pull"

type Request struct {
	Request string                 `json:"request"`
	Data    map[string]interface{} `json:"data"`
}
