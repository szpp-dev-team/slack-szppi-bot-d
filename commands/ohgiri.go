package commands

import (
	"encoding/json"
	"log"
	"os"
	"strings"
	"time"

	"github.com/slack-go/slack"
)

type Winner struct {
	Kaitou string
	UserName string
	Reactions int
}

type Ohgiri struct {
	Odai  string `json:"odai"`
	Kotae string `json:"kotae"`
}

type SubHandlerOhgiri struct {
	c       *slack.Client
	ohgiris *[]Ohgiri
	cursor  int
}

func NewSubHandlerOhgiri(c *slack.Client) *SubHandlerOhgiri {
	o, err := loadOhgiris()
	if err != nil {
		log.Fatalln("failed to load JSON datbase:", err)
	}
	return &SubHandlerOhgiri{c, o, 0}
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
	//_, _, err := o.c.PostMessage(slashCmd.ChannelID, slack.MsgOptionBlocks(blocks...), slack.MsgOptionText("お題を発表するっぴ！", false))
	
	messageChannelID, messageTimestamp, err := o.c.PostMessage(slashCmd.ChannelID, slack.MsgOptionBlocks(blocks...), slack.MsgOptionText("お題を発表するっぴ！", false))
	if err != nil {
		log.Fatalln("failed to post message.", err)
	}

	time.Sleep(time.Second * 30)
	threadMessage, _, _, err := o.c.GetConversationReplies(&slack.GetConversationRepliesParameters{ChannelID: messageChannelID, Timestamp: messageTimestamp})
	if err != nil{
		log.Fatalln("faild", err)
	}

	winners := chooseWinner(threadMessage)
	
	if isWinnerEmpty(winners){ //投票がなかった場合
		KotaeText := "投票がなかったっぴ！ちなみに模範解答はこれっぴ!\n*「" + ohgiri.Kotae + "」*\n"
		Kotaeblocks := []slack.Block{
			slack.NewSectionBlock(
				slack.NewTextBlockObject("mrkdwn", KotaeText, false, false),
				nil,
				nil,
			),
		}
		_, _, err := o.c.PostMessage(slashCmd.ChannelID, slack.MsgOptionBlocks(Kotaeblocks...), slack.MsgOptionTS(messageTimestamp), slack.MsgOptionText("結果発表〜〜〜！！", false))
		if err != nil{
			log.Fatalln("errdesu")
		}
	} else {
		if len(winners) == 1 { //優勝者が一人だった場合
			user, err := o.c.GetUserInfo(winners[0].UserName)
			if err != nil{
				log.Fatalln("error", err)
			}
			KotaeText := "この回答が一番おもしろかったっぴ!\n"+ user.Profile.DisplayName +" 作 *「" + winners[0].Kaitou + "」*"
			Kotaeblocks := []slack.Block{
				slack.NewSectionBlock(
					slack.NewTextBlockObject("mrkdwn", KotaeText, false, false),
					nil,
					nil,
				),
			}
			_, _, err = o.c.PostMessage(slashCmd.ChannelID, slack.MsgOptionBlocks(Kotaeblocks...), slack.MsgOptionTS(messageTimestamp), slack.MsgOptionText("結果発表〜〜〜！！", false))
			if err != nil{
				log.Fatalln("errdesu")
			}
		} else { //優勝者が複数いた場合
			KotaeText := []string{}
			for _, winner := range winners{
				user, err := o.c.GetUserInfo(winner.UserName)
				if err != nil {
					log.Fatalln("erreesu", err)
				}
				KotaeText = append(KotaeText, user.Profile.DisplayName + " 作 *「" + winner.Kaitou + "」*")
			}

			Kotaeblocks := []slack.Block{slack.NewSectionBlock(slack.NewTextBlockObject("mrkdwn", "面白すぎて一つに決められなかったっぴ！！", false, false), nil, nil)}
			Kotaeblocks = append(Kotaeblocks, slack.NewSectionBlock(slack.NewTextBlockObject("mrkdwn", strings.Join(KotaeText, "\n"), false, false), nil, nil))

			_, _, err := o.c.PostMessage(slashCmd.ChannelID, slack.MsgOptionBlocks(Kotaeblocks...), slack.MsgOptionTS(messageTimestamp), slack.MsgOptionText("結果発表〜〜〜！！", false))
			if err != nil{
				log.Fatalln("errdesu")
			}
		}
	}
	return err
}

func loadOhgiris() (*[]Ohgiri, error) {
	content, err := os.ReadFile("data/ohgiris.json")
	if err != nil {
		return nil, err
	}

	//	content := []byte(`[
	//	{"odai": "odai1", "kotae": "kotae1"},
	//	{"odai": "odai2", "kotae": "kotae2"},
	//	{"odai": "odai3", "kotae": "kotae3"},
	//	{"odai": "odai4", "kotae": "kotae4"},
	//	{"odai": "odai5", "kotae": "kotae5"}
	//]`)
	o := []Ohgiri{}
	err = json.Unmarshal(content, &o)
	if err != nil {
		return nil, err
	}
	return &o, err
}

func isWinnerEmpty(winner []Winner) (bool){
	frag := false
	if len(winner) == 0{
		frag = true
	}
	return frag
}

func chooseWinner(threadMessage []slack.Message)([]Winner){
	winners := []Winner{}
	reactionCount := 0
	winnerindex := 0

	for _, msg := range threadMessage{
		if msg.Reactions != nil{
			for _, reaction := range msg.Reactions{
				reactionCount += reaction.Count
			}
			if !isWinnerEmpty(winners){
				if winners[0].Reactions < reactionCount{ //もしカウントが多い人がいた場合 -> 配列をクリアして新しく追加
					winners = []Winner{}
					winners = append(winners, Winner{
						Kaitou: msg.Text,
						UserName: msg.User,
						Reactions: reactionCount,
					})
					winnerindex = 1
				} else if winners[0].Reactions == reactionCount { //カウントが一緒なら -> 配列に追加
					winners = append(winners, Winner{
						Kaitou: msg.Text,
						UserName: msg.User,
						Reactions: reactionCount,
					})
					winnerindex += 1
				}
			} else { //最初は無条件追加
				winners = append(winners, Winner{
					Kaitou: msg.Text,
					UserName: msg.User,
					Reactions: reactionCount,
				})
				winnerindex = 1
			}
		}
		reactionCount = 0
	}
	return winners
}