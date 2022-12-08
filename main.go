package main

import (
	//	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/slack-go/slack"
	"github.com/szpp-dev-team/szpp-slack-bot/commands"
	// "google.golang.org/api/customsearch/v1"
	// "google.golang.org/api/option"
)

func main() {
	botUserOauthToken := os.Getenv("BOT_USER_OAUTH_TOKEN")
	signingSecret := os.Getenv("SIGNING_SECRET")
	//	customsearchApiKey := os.Getenv("CUSTOM_SEARCH_API_KEY")

	port := getenvOr("PORT", "8080")

	client := slack.New(botUserOauthToken)
	//	customsearchService, err := customsearch.NewService(context.Background(), option.WithAPIKey(customsearchApiKey))
	//	if err != nil {
	//		log.Fatal(err)
	//	}

	slashHandler := NewSlashHandler(client)
	// カスタムコマンドの追加はここで行う(インスタンスを引数に渡せばおk)
	slashHandler.RegisterSubHandlers(
		commands.NewSubHandlerOmikuji(client),
		//		commands.NewSubHandlerImage(client, customsearchService),
		commands.NewSubHandlerOhgiri(client),
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

		slashHandler.Handle(w, &slashCmd)
	})

	http.HandleFunc("/slack/actions", func(w http.ResponseWriter, r *http.Request) {
		var payload slack.InteractionCallback
		err := json.Unmarshal([]byte(r.FormValue("payload")), &payload)
		if err != nil {
			log.Println("Could not parse action response JSON:", err)
			http.Error(w, "Could not parse action responseJSON", http.StatusBadRequest)
			return
		}
		switch payload.Type {
		case slack.InteractionTypeBlockActions:
			for _, v := range payload.ActionCallback.BlockActions {
				log.Println("ActionID:", v.ActionID)
				// TODO: このあと ActionID ごとにハンドリングする close_button, cancel_button
			}
			return
		default:
			w.WriteHeader(http.StatusOK)
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
