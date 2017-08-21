package client

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseMarkdown(t *testing.T) {
	assert := assert.New(t)

	markdown := `
# Here is slide 1
---
# here is slide 2

* with
* a
* list
`
	manager := &DeckConfigManager{}
	slides := manager.ParseMarkdown(markdown)

	assert.Equal(2, len(slides))
	assert.Equal("\n# Here is slide 1\n", slides[0].Markdown)
	assert.Equal("# here is slide 2\n\n* with\n* a\n* list\n", slides[1].Markdown)

	assert.Equal(0, slides[0].Position)
	assert.Equal(1, slides[1].Position)

	assert.Equal(0, slides[0].ID)
}
