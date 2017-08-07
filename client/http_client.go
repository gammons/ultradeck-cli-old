package client

import (
	"fmt"
	"io/ioutil"
	"net/http"
)

const HttpUrl = "http://localhost:3000/"

type HttpClient struct {
	Token string
}

func NewHttpClient(token string) *HttpClient {
	return &HttpClient{Token: token}
}

func (h *HttpClient) PerformRequest(path string) []byte {

	url := HttpUrl + path
	client := &http.Client{}
	req, _ := http.NewRequest("GET", url, nil)
	authHeader := fmt.Sprintf("Bearer %s", h.Token)

	req.Header.Set("Authorization", authHeader)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	res, _ := client.Do(req)

	bodyBytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		panic(err)
	}
	defer res.Body.Close()

	return bodyBytes
}
