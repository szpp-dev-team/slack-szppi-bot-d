package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/slack-go/slack"
	"github.com/szpp-dev-team/szpp-slack-bot/commands"
)

func main() {
	botUserOauthToken := os.Getenv("BOT_USER_OAUTH_TOKEN")
	signingSecret := os.Getenv("SIGNING_SECRET")
	port := getenvOr("PORT", "8080")

	client := slack.New(botUserOauthToken)

	slashHandler := NewSlashHandler(client)
	// カスタムコマンドの追加はここで行う(インスタンスを引数に渡せばおk)
	slashHandler.RegisterSubHandlers(
		commands.NewSubHandlerOmikuji(client),
	)

	http.HandleFunc("/slack/events", func(w http.ResponseWriter, r *http.Request) {
		verifier, err := slack.NewSecretsVerifier(r.Header, signingSecret)
		if err != nil {
			log.Println("failed to verify secrets:", err)
			http.Error(w, "failed to verify secrets", http.StatusInternalServerError)
			return
		}

		r.Body = io.NopCloser(io.TeeReader(r.Body, &verifier))
		slashCmd, err := slack.SlashCommandParse(r)
		if err != nil {
			log.Println("failed to parse slash command:", err)
			http.Error(w, "failed to parse slash command", http.StatusInternalServerError)
			return
		}

		if err := verifier.Ensure(); err != nil {
			log.Println("failed to ensure compares the signature:", err)
			http.Error(w, "failed to ensure compares the signature", http.StatusUnauthorized)
			return
		}

		switch slashCmd.Command {
		case "/szppi":
			slashHandler.Handle(w, &slashCmd)
		default:
			log.Println("no such slash command", slashCmd.Command)
			http.Error(w, "no such slash command "+slashCmd.Command, http.StatusBadRequest)
			return
		}
	})

	if err := http.ListenAndServe(fmt.Sprintf(":%v", port), nil); err != nil {
		log.Fatal(err)
	}
}

func getenvOr(key, altValue string) string {
	value := os.Getenv(key)
	if value == "" {
		value = altValue
	}
	return value
}
