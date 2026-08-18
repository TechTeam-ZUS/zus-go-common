package notify

import (
	"fmt"
	"time"
)

// Card mirrors a single-block Lark/Feishu interactive card payload.
type Card struct {
	MsgType string   `json:"msg_type"`
	Card    cardBody `json:"card"`
}

type cardBody struct {
	Config   cardConfig `json:"config"`
	Header   cardHeader `json:"header"`
	Elements []element  `json:"elements"`
}

type cardConfig struct {
	WideScreenMode bool `json:"wide_screen_mode"`
}

type cardHeader struct {
	Title    cardTitle `json:"title"`
	Template string    `json:"template"`
}

type cardTitle struct {
	Tag     string `json:"tag"`
	Content string `json:"content"`
}

type element struct {
	Tag  string      `json:"tag"`
	Text elementText `json:"text"`
}

type elementText struct {
	Tag     string `json:"tag"`
	Content string `json:"content"`
}

// NewCard builds a single-block interactive card
func NewCard(title, template, content string) Card {
	return Card{
		MsgType: "interactive",
		Card: cardBody{
			Config: cardConfig{WideScreenMode: true},
			Header: cardHeader{
				Title:    cardTitle{Tag: "plain_text", Content: title},
				Template: template,
			},
			Elements: []element{
				{Tag: "div", Text: elementText{Tag: "lark_md", Content: wrap(content)}},
			},
		},
	}
}

func wrap(content string) string {
	return fmt.Sprintf("Time: %s\n%s", time.Now().Format(timeLayout), content)
}
