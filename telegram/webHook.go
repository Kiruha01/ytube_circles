package telegram

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"net/http"
	"ytubecircles/config"
)

type WebHookBot struct {
	Bot
}

// Start запускает бота
func (b *WebHookBot) Start() error {
	wh, _ := tgbotapi.NewWebhook(config.Config.WHUrl + "/updates")

	log.Println("Starting webhook...")

	updates := b.bot.ListenForWebhook("/updates")
	go http.ListenAndServe("0.0.0.0:8443", nil)

	_, err := b.bot.Request(wh)
	if err != nil {
		log.Fatal(err)
	}

	info, err := b.bot.GetWebhookInfo()
	if err != nil {
		log.Fatal(err)
	}

	if info.LastErrorDate != 0 {
		log.Printf("Telegram callback failed: %s", info.LastErrorMessage)
	}

	log.Println("Listening for webhook...")

	return b.processUpdates(updates)
}
