package client

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
)

type Deck struct {
	Config *DeckConfig `json:"deck"`
}

type DeckConfig struct {
	ID           int      `json:"id"`
	Title        string   `json:"title"`
	Description  string   `json:"description"`
	Slug         string   `json:"slug"`
	IsPublic     bool     `json:"is_public"`
	ThemeID      int      `json:"theme_id"`
	ThemeVersion int      `json:"theme_version"`
	UpdatedAt    string   `json:"updated_at"`
	Slides       []*Slide `json:"slides_attributes"`
	Assets       []*Asset `json:"assets_attributes"`
}

type Slide struct {
	ID             int    `json:"-"`
	Position       int    `json:"position"`
	Markdown       string `json:"markdown"`
	PresenterNotes string `json:"presenter_notes"`
	ColorVariation int    `json:"color_variation"`
}

type Asset struct {
	ID       int    `json:"id"`
	Filename string `json:"filename"`
	URL      string `json:"url"`
}

type DeckConfigManager struct{}

func (d *DeckConfigManager) Write(jsonData []byte) {
	var deckConfig *DeckConfig
	if err := json.Unmarshal(jsonData, &deckConfig); err != nil {
		log.Println("Error writing deck config ", err)
	}

	marshalledData, _ := json.Marshal(deckConfig)
	if err := ioutil.WriteFile(".ud.json", marshalledData, 0644); err != nil {
		log.Println("Error writing deck config: ", err)
	}
}

func (d *DeckConfigManager) ReadFile() *DeckConfig {
	if !d.FileExists() {
		return nil
	}

	data, err := ioutil.ReadFile(".ud.json")
	if err != nil {
		log.Println("error reading deck config file: ", err)
	}

	var deckConfig *DeckConfig
	err = json.Unmarshal(data, &deckConfig)
	if err != nil {
		log.Println("error reading deck config file: ", err)
	}

	return deckConfig
}

func (d *DeckConfigManager) PrepareJSON() []byte {
	config := d.ReadFile()
	config.Slides = d.ParseMarkdown()

	deck := &Deck{Config: config}

	j, _ := json.Marshal(&deck)

	fmt.Println("json is ", string(j))

	return j
}

func (d *DeckConfigManager) GetDeckID() int {
	config := d.ReadFile()
	return config.ID
}

func (d *DeckConfigManager) ParseMarkdown() []*Slide {
	markdown, err := ioutil.ReadFile("deck.md")
	if err != nil {
		log.Println("error reading deck config file: ", err)
	}

	splitted := strings.Split(string(markdown), "---\n")
	var slides []*Slide

	for i, markdown := range splitted {
		slides = append(slides, &Slide{Position: (i + 1), Markdown: markdown})
	}
	return slides
}

func (d *DeckConfigManager) FileExists() bool {
	if _, err := os.Stat(".ud.json"); os.IsNotExist(err) {
		return false
	}
	return true
}
