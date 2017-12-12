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
	ID          string   `json:"id"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Slug        string   `json:"slug"`
	IsPublic    bool     `json:"is_public"`
	ThemeID     string   `json:"theme_id"`
	UpdatedAt   string   `json:"updated_at"`
	Slides      []*Slide `json:"slides_attributes"`
	Assets      []*Asset `json:"assets_attributes"`
}

type Slide struct {
	ID             string `json:"-"`
	Markdown       string `json:"markdown"`
	PresenterNotes string `json:"presenter_notes"`
	ColorVariation int    `json:"color_variation"`
}

type Asset struct {
	ID        string `json:"-"`
	Filename  string `json:"filename"`
	URL       string `json:"url"`
	UpdatedAt string `json:"updated_at"`
}

type DeckConfigManager struct{}

func (d *DeckConfigManager) Write(jsonData []byte) {
	var deckConfig *DeckConfig
	fmt.Println("About to show jsonData")
	fmt.Println(string(jsonData[:]))
	if err := json.Unmarshal(jsonData, &deckConfig); err != nil {
		log.Println("Error writing deck", err)
	}

	d.WriteConfig(deckConfig)
}

func (d *DeckConfigManager) WriteConfig(deckConfig *DeckConfig) {
	log.Println(deckConfig)
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

func (d *DeckConfigManager) PrepareJSON(deckConfig *DeckConfig) []byte {
	deckConfig.Slides = d.ParseMarkdown(deckConfig)

	deck := &Deck{Config: deckConfig}

	j, _ := json.Marshal(&deck)

	return j
}

func (d *DeckConfigManager) GetDeckID() string {
	config := d.ReadFile()
	return config.ID
}

func (d *DeckConfigManager) ParseMarkdown(deckConfig *DeckConfig) []*Slide {
	markdown, err := ioutil.ReadFile("deck.md")
	if err != nil {
		log.Println("error reading deck config file: ", err)
	}

	splitted := strings.Split(string(markdown), "---\n")
	var slides []*Slide

	for _, markdown := range splitted {
		// attempt to find the previous slide from the deckConfig
		var previousSlide *Slide

		for i := range deckConfig.Slides {
			if deckConfig.Slides[i].Markdown == markdown {
				previousSlide = deckConfig.Slides[i]
			}
		}

		newSlide := &Slide{Markdown: markdown}

		if previousSlide != nil {
			newSlide.PresenterNotes = previousSlide.PresenterNotes
			newSlide.ColorVariation = previousSlide.ColorVariation
		}

		slides = append(slides, newSlide)
	}
	return slides
}

func (d *DeckConfigManager) FileExists() bool {
	if _, err := os.Stat(".ud.json"); os.IsNotExist(err) {
		return false
	}
	return true
}
