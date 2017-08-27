package client

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"mime"
	"net/http"
	"net/url"
	"strings"
)

type S3InfoResponse struct {
	Bucket string `json:"bucket"`
	Fields *FieldsResponse
	Url    string `json:"url"`
}

type FieldsResponse struct {
	Key                 string `json:"key"`
	SuccessActionStatus string `json:"success_action_status"`
	Acl                 string `json:"acl"`
	Policy              string `json:"policy"`
	XAmzCredential      string `json:"x-amz-credential"`
	XAmzAlgorithm       string `json:"x-amz-algorithm"`
	XAmzDate            string `json:"x-amz-date"`
	XAmzSignature       string `json:"x-amz-signature"`
}

type AssetManager struct{}

func (a *AssetManager) SyncAssets(token string) {
	httpClient := NewHttpClient(token)

	for _, file := range a.readFiles() {
		fmt.Println(file)

		// call s3_info to get place to put it
		bodyBytes := httpClient.GetRequest("/api/v1/s3_upload_info")
		fmt.Println(string(bodyBytes))

		fileBody, _ := ioutil.ReadFile(file)

		var s3InfoResponse *S3InfoResponse
		_ = json.Unmarshal(bodyBytes, &s3InfoResponse)

		httpClient := http.Client{}
		fmt.Println("POSTing to url: ", s3InfoResponse.Url)

		form := url.Values{}
		form.Add("key", s3InfoResponse.Fields.Key)
		form.Add("success_action_status", s3InfoResponse.Fields.SuccessActionStatus)
		form.Add("acl", s3InfoResponse.Fields.Acl)
		form.Add("policy", s3InfoResponse.Fields.Policy)
		form.Add("content-type", a.mimeType(file))
		form.Add("x-amz-credential", s3InfoResponse.Fields.XAmzCredential)
		form.Add("x-amz-algorithm", s3InfoResponse.Fields.XAmzAlgorithm)
		form.Add("x-amz-date", s3InfoResponse.Fields.XAmzDate)
		form.Add("x-amz-signature", s3InfoResponse.Fields.XAmzSignature)
		form.Add("file", string(fileBody))

		req, _ := http.NewRequest("POST", s3InfoResponse.Url, strings.NewReader(form.Encode()))
		req.PostForm = form
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

		resp, rerr := httpClient.Do(req)
		if rerr != nil {
			fmt.Println("error doing request: ", rerr)
		}

		defer resp.Body.Close()

		fmt.Println(resp.Status)
		bodyBytes, _ = ioutil.ReadAll(resp.Body)
		fmt.Println(string(bodyBytes))

		// put it there
		// call ultradeck backend to say we put it there
	}
}

func (a *AssetManager) readFiles() []string {
	files, err := ioutil.ReadDir(".")
	if err != nil {
		log.Fatal("Error reading directory: ", err)
	}

	var ret []string

	for _, file := range files {
		// TODO: support more extension types?
		if strings.HasPrefix(a.mimeType(file.Name()), "image") {
			ret = append(ret, file.Name())
		}
	}

	return ret
}

func (a *AssetManager) mimeType(fileName string) string {
	splitted := strings.Split(fileName, ".")
	ext := splitted[len(splitted)-1]
	ext = "." + ext
	mimeType := mime.TypeByExtension(ext)
	return mimeType
}
