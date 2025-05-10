package telegram

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
)

func (b *Bot) handleHelp(update tgbotapi.Update) {
	msgConfig := tgbotapi.MessageConfig{
		BaseChat: tgbotapi.BaseChat{
			ChatID:           update.Message.Chat.ID,
			ReplyToMessageID: 0,
		},
		Text: "This bot convert YouTube video to Video\\-Message circle by link to video and time " +
			"start and time end markers\n\n" +
			"```\nYT-link [time-start [duration]]\n```\n*Examples:*\n" +
			"```\nhttps://youtu.be/dQw4w9WgXcQ 00:00:43.5 00:00:10\n```For rick roll from 43.5 to 53 seconds\n" +
			"```\nhttps://www.youtube.com/watch?v=dQw4w9WgXcQ 00:00:43 10.5\n```For rick roll from 43 to 53.5 seconds\n" +
			"```\nhttps://www.youtube.com/watch?v=dQw4w9WgXcQ 43.5\n```For rick roll from 43.5 to 103.5 seconds (60 seconds by default and max)\n" +
			"```\nhttps://www.youtube.com/watch?v=dQw4w9WgXcQ\n```For rick roll from 0 to 60 seconds",
		ParseMode:             "Markdown",
		DisableWebPagePreview: false,
	}
	_, err := b.bot.Send(
		msgConfig,
	)
	if err != nil {
		log.Printf("Error while sending message: %v", err)
	}
	return
}
