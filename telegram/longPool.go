package telegram

import tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

type LongPoolBot struct {
	Bot
}

func (b *LongPoolBot) Start() error {
	b.bot.Send(tgbotapi.DeleteWebhookConfig{})
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := b.bot.GetUpdatesChan(u)
	return b.processUpdates(updates)
}
