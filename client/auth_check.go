package client

import "encoding/json"

type AuthCheck struct{}

type AuthCheckResponse struct {
	IsSignedIn bool   `json:"is_signed_in"`
	Name       string `json:"username"`
}

func (a *AuthCheck) CheckAuth(token string) *AuthCheckResponse {
	httpClient := NewHttpClient(token)
	bodyBytes := httpClient.PerformRequest("api/v1/auth/me")

	resp := &AuthCheckResponse{}
	if err := json.Unmarshal(bodyBytes, &resp); err != nil {
		panic(err)
	}
	return resp
}
