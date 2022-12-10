package commands

import (
	"encoding/json"
	"log"
	"os"
	"strings"
	"time"

	"github.com/emirpasic/gods/sets/hashset"
	"github.com/slack-go/slack"
)

type Winner struct {
	Kaitou    string
	Name      string
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
		log.Print(err)
	}
	return &SubHandlerOhgiri{c, o, 0}
}

func (o *SubHandlerOhgiri) Name() string {
	return "ohgiri"
}

func (o *SubHandlerOhgiri) Handle(slashCmd *slack.SlashCommand) error {
	//お題生成部
	var odaiText string
	var isUsersOdai bool
	ohgiri := (*o.ohgiris)[o.cursor]
	if len(slashCmd.Text) > len("ohgiri") {
		isUsersOdai = true
	} else {
		isUsersOdai = false
	}
	if isUsersOdai {
		odaiText = "*「" + strings.Join(strings.Fields(slashCmd.Text)[1:], "") + "」*\n答えを思いついたらスレッドに書くっぴ！"
	} else {
		odaiText = "*「" + ohgiri.Odai + "」*\n答えを思いついたらスレッドに書くっぴ！"
	}

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
		/*slack.NewActionBlock(
			"button_group",
			slack.NewButtonBlockElement("close_button", "", slack.NewTextBlockObject("plain_text", "回答締切", false, false)).WithStyle(slack.StylePrimary),
			slack.NewButtonBlockElement("cancel_button", "", slack.NewTextBlockObject("plain_text", "やめる", false, false)).WithStyle(slack.StyleDanger),
		),*/
	}
	messageChannelID, messageTimestamp, err := o.c.PostMessage(slashCmd.ChannelID, slack.MsgOptionBlocks(blocks...), slack.MsgOptionText("お題を発表するっぴ！", false))
	if err != nil {
		log.Print(err)
		return err
	}

	time.Sleep(time.Second * 10) //今は仮で時間制にしている

	//回答集計部 + 出力部
	threadMessage, _, _, err := o.c.GetConversationReplies(&slack.GetConversationRepliesParameters{ChannelID: messageChannelID, Timestamp: messageTimestamp})
	if err != nil {
		log.Print(err)
		return err
	}

	if len(threadMessage) <= 1 { //回答した人がいなかった場合(botからの初期メッセージがあるのでパラメータは1以下の時)
		var sendMessageText string
		if isUsersOdai {
			sendMessageText = "誰も回答してくれなかったっぴ\n｡ﾟ(ﾟ＾ω＾ﾟ)ﾟ｡\n"
		} else {
			sendMessageText = "誰も回答してくれなかったっぴ\n｡ﾟ(ﾟ＾ω＾ﾟ)ﾟ｡\nちなみに模範解答はこれっぴ！\n*「" + ohgiri.Kotae + "」*\n"
		}
		msgBlock := []slack.Block{
			slack.NewSectionBlock(
				slack.NewTextBlockObject("mrkdwn", sendMessageText, false, false),
				nil,
				nil,
			),
		}
		_, _, err := o.c.PostMessage(slashCmd.ChannelID, slack.MsgOptionBlocks(msgBlock...), slack.MsgOptionTS(messageTimestamp), slack.MsgOptionText("結果発表〜〜〜！！", false))
		if err != nil {
			log.Print(err)
			return err
		}

	} else { //回答が存在する場合
		winners := chooseWinner(threadMessage) //優勝者のリスト(Winnerの構造体のスライス)
		if isWinnerEmpty(winners) {            //投票がなかった場合
			var sendMessageText string
			if isUsersOdai {
				sendMessageText = "投票がなかったっぴ！\n｡ﾟ(ﾟஇωஇﾟ)ﾟ｡\n"
			} else {
				sendMessageText = "投票がなかったっぴ！\n｡ﾟ(ﾟஇωஇﾟ)ﾟ｡\nちなみに模範解答はこれっぴ！\n*「" + ohgiri.Kotae + "」*\n"
			}
			msgBlock := []slack.Block{
				slack.NewSectionBlock(
					slack.NewTextBlockObject("mrkdwn", sendMessageText, false, false),
					nil,
					nil,
				),
			}
			_, _, err := o.c.PostMessage(slashCmd.ChannelID, slack.MsgOptionBlocks(msgBlock...), slack.MsgOptionTS(messageTimestamp), slack.MsgOptionText("結果発表〜〜〜！！", false))
			if err != nil {
				log.Print(err)
				return err
			}
		} else {
			if len(winners) == 1 { //優勝者が一人だった場合
				userInfo, err := o.c.GetUserInfo(winners[0].Name)
				if err != nil {
					log.Print(err)
					return err
				}
				sendMessageText := "この回答が一番おもしろかったっぴ！\n( ･`ω･´)\n" + userInfo.Profile.DisplayName + " 作 *「" + winners[0].Kaitou + "」*"
				msgBlock := []slack.Block{
					slack.NewSectionBlock(
						slack.NewTextBlockObject("mrkdwn", sendMessageText, false, false),
						nil,
						nil,
					),
				}
				_, _, err = o.c.PostMessage(slashCmd.ChannelID, slack.MsgOptionBlocks(msgBlock...), slack.MsgOptionTS(messageTimestamp), slack.MsgOptionText("結果発表〜〜〜！！", false))
				if err != nil {
					log.Print(err)
					return err
				}
			} else { //優勝者が複数いた場合
				sendMessageText := []string{}
				for _, winner := range winners {
					userInfo, err := o.c.GetUserInfo(winner.Name)
					if err != nil {
						log.Print(err)
						return err
					}
					sendMessageText = append(sendMessageText, userInfo.Profile.DisplayName+" 作 *「"+winner.Kaitou+"」*")
				}

				msgBlock := []slack.Block{slack.NewSectionBlock(slack.NewTextBlockObject("mrkdwn", "面白すぎて一つに決められなかったっぴ！\nΣ( ˙꒳˙ ;)", false, false), nil, nil)}
				msgBlock = append(msgBlock, slack.NewSectionBlock(slack.NewTextBlockObject("mrkdwn", strings.Join(sendMessageText, "\n"), false, false), nil, nil))

				_, _, err := o.c.PostMessage(slashCmd.ChannelID, slack.MsgOptionBlocks(msgBlock...), slack.MsgOptionTS(messageTimestamp), slack.MsgOptionText("結果発表〜〜〜！！", false))
				if err != nil {
					log.Print(err)
					return err
				}
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

func isWinnerEmpty(winner []Winner) bool {
	frag := false
	if len(winner) == 0 {
		frag = true
	}
	return frag
}

func chooseWinner(threadMessage []slack.Message) []Winner {
	winners := []Winner{}            //優勝者のリスト
	reactionCount := 0               //ループ単位(一つの大喜利回答)あたりのリアクション人数
	winnerIndex := 0                 //優勝者の数(ループで変動あり)
	reactionUserSet := hashset.New() //リアクション数によらずリアクションした人で判別するためのセット　一ループごとclear

	for _, msg := range threadMessage {
		if msg.Reactions != nil {
			for _, reaction := range msg.Reactions {
				for _, reactionUser := range reaction.Users {
					if !reactionUserSet.Contains(reactionUser) {
						reactionUserSet.Add(reactionUser)
						reactionCount++
					}
				}
			}
			if !isWinnerEmpty(winners) {
				if winners[0].Reactions < reactionCount { //もしカウントが多い人がいた場合 -> 配列をクリアして新しく追加
					winners = []Winner{}
					winners = append(winners, Winner{
						Kaitou:    msg.Text,
						Name:      msg.User,
						Reactions: reactionCount,
					})
					winnerIndex = 1
				} else if winners[0].Reactions == reactionCount { //カウントが一緒なら -> 配列に追加
					winners = append(winners, Winner{
						Kaitou:    msg.Text,
						Name:      msg.User,
						Reactions: reactionCount,
					})
					winnerIndex++
				}
			} else { //最初は無条件追加
				winners = append(winners, Winner{
					Kaitou:    msg.Text,
					Name:      msg.User,
					Reactions: reactionCount,
				})
				winnerIndex = 1
			}
		}
		reactionUserSet.Clear()
		reactionCount = 0
	}
	return winners
}
