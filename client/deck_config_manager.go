package client

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"strings"
)

type DeckConfig struct {
	ID           int      `json:"id"`
	Title        string   `json:"title"`
	Description  string   `json:"description"`
	Slug         string   `json:"slug"`
	IsPublic     bool     `json:"is_public"`
	Theme        string   `json:"theme"`
	ThemeVersion string   `json:"themeVersion"`
	UpdatedAt    string   `json:"updatedAt"`
	Slides       []*Slide `json:"slides"`
	Assets       []*Asset `json:"assets"`
}

type Slide struct {
	ID             int    `json:"id"`
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

// func (d *DeckConfigManager) read() *DeckConfig {
// }

func (d *DeckConfigManager) ParseMarkdown(markdown string) []*Slide {
	splitted := strings.Split(markdown, "---\n")
	var slides []*Slide

	for i, markdown := range splitted {
		slides = append(slides, &Slide{Position: i, Markdown: markdown})
	}
	return slides
}
