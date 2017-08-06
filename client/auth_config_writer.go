package client

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/user"
)

type AuthConfig struct {
	AuthJson *AuthJson
}

type AuthJson struct {
	Token string `json:"token"`
}

func NewAuthConfigWriter(token string) *AuthConfig {
	return &AuthConfig{AuthJson: &AuthJson{Token: token}}
}

func (c *AuthConfig) AuthFileExists() bool {
	if _, err := os.Stat(c.configFileLocation()); os.IsNotExist(err) {
		return false
	}
	return true
}

func (c *AuthConfig) WriteAuth() {
	log.Println("Writing auth")
	data, _ := json.Marshal(c.AuthJson)
	log.Println("Data is", data)

	if c.AuthFileExists() {
		c.RemoveAuthFile()
	}

	if err := os.MkdirAll(c.configFilePath(), os.ModePerm); err != nil {
		log.Println("Error creating directory structure", err)
		os.Exit(1)
	}

	if err := ioutil.WriteFile(c.configFileLocation(), []byte(data), 0644); err != nil {
		log.Println("Error writing json file", err)
		os.Exit(1)
	}
	log.Println("Done writing auth")
}

func (c *AuthConfig) RemoveAuthFile() {
	if !c.AuthFileExists() {
		return
	}

	if err := os.RemoveAll(c.configFilePath()); err != nil {
		log.Println("Error removing config file", err)
	}
}

func (c *AuthConfig) configFilePath() string {
	usr, _ := user.Current()
	return fmt.Sprintf("%s/.config/ultradeck/", usr.HomeDir)
}

func (c *AuthConfig) configFileLocation() string {
	return c.configFilePath() + "auth.json"
}
