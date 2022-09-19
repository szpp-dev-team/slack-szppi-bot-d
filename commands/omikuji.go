package commands

import (
	"math/rand"
	"time"

	"github.com/slack-go/slack"
)

type SubHandlerOmikuji struct {
	c *slack.Client
}

func NewSubHandlerOmikuji(c *slack.Client) *SubHandlerOmikuji {
	return &SubHandlerOmikuji{c}
}

func (o *SubHandlerOmikuji) Name() string {
	return "omikuji"
}

func (o *SubHandlerOmikuji) Handle(slashCmd *slack.SlashCommand) error {
	rand.Seed(time.Now().UnixMicro())
	p := rand.Float64()

	res := ""
	if p < 0.25 {
		res = "大吉！！！おめでとう！:tada::tada::tada:"
	} else if p < 0.5 {
		res = "中吉！！おめ！:tada:"
	} else if p < 0.75 {
		res = "吉！！！"
	} else {
		res = "う　し　た　ぷ　に　き　あ　く　ん　笑"
	}

	_, _, _, err := o.c.SendMessage(slashCmd.ChannelID, slack.MsgOptionText(res, false))

	return err
}
