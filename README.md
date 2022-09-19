# szpp-slack-bot

`/szppi` スラッシュコマンド

## ローカルでの実行

1. `.env` を置く
2. `$ docker compose up --build`
3. `ngrok http 8080`
4. `https://******.jp.ngrok.io/slack/events` を slash command の編集ページに貼り付ける

## 拡張について

コマンドは以下の interface を実装した構造体を定義し、`slashHandler.RegisterSubHandlers()` の引数に与えることで実装することが可能です。
例: `commands/omikuji.go`

```go
type CommandExecutor interface {
	Handle(slashCmd *slack.SlashCommand) error  // コマンドの振る舞い
	Name() string                               // コマンドの名前
}
```