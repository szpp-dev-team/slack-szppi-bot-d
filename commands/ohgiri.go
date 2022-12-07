package commands

import (
	"encoding/json"

	"github.com/slack-go/slack"
)

type Ohgiri struct {
	Odai  string `json:"odai"`
	Kotae string `json:"kotae"`
}

type SubHandlerOhgiri struct {
	c       *slack.Client
	n       int32
	ohgiris *[]Ohgiri
	cursor  int
}

func NewSubHandlerOhgiri(c *slack.Client) *SubHandlerOhgiri {
	o, err := loadOhgiris()
	if err != nil {
		// error handling
	}
	return &SubHandlerOhgiri{c, 0, o, 0}
}

func (o *SubHandlerOhgiri) Name() string {
	return "ohgiri"
}

func (o *SubHandlerOhgiri) Handle(slashCmd *slack.SlashCommand) error {
	ohgiri := (*o.ohgiris)[o.cursor]
	odaiText := "*「" + ohgiri.Odai + "」*\n答えを思いついたらスレッドに書くっぴ！"
	o.cursor++
	if o.cursor >= len(*o.ohgiris) {
		o.cursor = 0
	}
	blocks := []slack.Block{
		slack.NewSectionBlock(
			slack.NewTextBlockObject("mrkdwn", odaiText, false, false),
			nil,
			nil,
		),
		slack.NewActionBlock(
			"button_group",
			slack.NewButtonBlockElement("close_button", "", slack.NewTextBlockObject("plain_text", "回答締切", false, false)).WithStyle(slack.StylePrimary),
			slack.NewButtonBlockElement("cancel_button", "", slack.NewTextBlockObject("plain_text", "やめる", false, false)).WithStyle(slack.StyleDanger),
		),
	}
	_, _, err := o.c.PostMessage(slashCmd.ChannelID, slack.MsgOptionBlocks(blocks...), slack.MsgOptionText("お題を発表するっぴ！", false))
	return err
}

func loadOhgiris() (*[]Ohgiri, error) {
	//content, err := ioutil.ReadFile("ohgiris.json")
	//if err != nil {
	//	return nil, err
	//}

	content := []byte(`[
	{"odai": "odai1", "kotae": "kotae1"},
	{"odai": "odai2", "kotae": "kotae2"},
	{"odai": "odai3", "kotae": "kotae3"},
	{"odai": "odai4", "kotae": "kotae4"},
	{"odai": "odai5", "kotae": "kotae5"}
]`)
	o := []Ohgiri{}
	err := json.Unmarshal(content, &o)
	if err != nil {
		return nil, err
	}
	return &o, err
}
