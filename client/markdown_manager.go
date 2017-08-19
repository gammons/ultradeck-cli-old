package client

import (
	"io/ioutil"
	"log"
)

type MarkdownManager struct{}

func (m *MarkdownManager) WriteFile() {
	var deckContents = "# Here's a great deck\n---\n"
	if err := ioutil.WriteFile("deck.md", []byte(deckContents), 0644); err != nil {
		log.Println("Error writing deck.md: ", err)
	}
}
