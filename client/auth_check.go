package client

import "encoding/json"

type AuthCheck struct{}

type AuthCheckResponse struct {
	IsSignedIn bool   `json:"is_signed_in"`
	Name       string `json:"username"`
	Token      string
}

func (a *AuthCheck) CheckAuth(token string) *AuthCheckResponse {
	httpClient := NewHttpClient(token)
	bodyBytes := httpClient.GetRequest("api/v1/auth/me")

	resp := &AuthCheckResponse{}
	if err := json.Unmarshal(bodyBytes, &resp); err != nil {
		panic(err)
	}
	return resp
}
