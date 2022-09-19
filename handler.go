package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/slack-go/slack"
)

type CommandExecutor interface {
	Handle(slashCmd *slack.SlashCommand) error
	Name() string
}

type SlashHandler struct {
	client    *slack.Client
	executors []CommandExecutor
}

func NewSlashHandler(client *slack.Client) *SlashHandler {
	return &SlashHandler{client, make([]CommandExecutor, 0, 100)}
}

func (s *SlashHandler) RegisterSubHandlers(executors ...CommandExecutor) {
	s.executors = append(s.executors, executors...)
}

func (s *SlashHandler) Handle(rw http.ResponseWriter, slashCmd *slack.SlashCommand) {
	fmt.Println(slashCmd.Text)

	rw.Header().Add("Content-Type", "application/json")
	rw.WriteHeader(http.StatusOK)
	// 打ったコマンドを表示させる
	msg := &slack.Msg{ResponseType: slack.ResponseTypeInChannel}
	_ = json.NewEncoder(rw).Encode(msg)

	go func() {
		for _, executor := range s.executors {
			if executor.Name() == slashCmd.Text {
				if err := executor.Handle(slashCmd); err != nil {
					log.Println(err)
					http.Error(rw, err.Error(), http.StatusInternalServerError)
					return
				}
				break
			}
		}
	}()
}
