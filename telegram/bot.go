package telegram

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"ytubecircles/config"
)

type BotInterface interface {
	processUpdates(updates tgbotapi.UpdatesChannel) error
	Start() error
}

// Bot представляет телеграм-бот
type Bot struct {
	bot *tgbotapi.BotAPI
}

// NewBot создает новый экземпляр бота
func NewBot(token string) (BotInterface, error) {
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, fmt.Errorf("failed to create bot: %w", err)
	}
	if config.Config.WHUrl != "" {
		log.Printf("Used WH url: %s", config.Config.WHUrl)

		return &WebHookBot{Bot{bot: bot}}, nil
	}
	log.Printf("Used non WH")

	return &LongPoolBot{Bot{bot: bot}}, nil
}

func (b *Bot) sendErrorMessage(chaiID int64, err error) {
	log.Printf("process error: %s", err)

	errorMessage := fmt.Sprintf("%s", err)
	_, msg_err := b.bot.Send(tgbotapi.NewMessage(chaiID, fmt.Sprintf("Error: \n```\n%v\n```", errorMessage[len(errorMessage)-min(len(errorMessage), 4000):])))
	if msg_err != nil {
		log.Printf("Failed to send error message: %v", msg_err)
	}
}

func (b *Bot) processUpdates(updates tgbotapi.UpdatesChannel) error {
	for update := range updates {
		go func() {
			if update.Message == nil {
				return
			}

			if update.Message.Text == "/start" || update.Message.Text == "/help" {
				b.handleHelp(update)
			}

			if update.Message.Text == "" {
				log.Printf("User: %s,  Text is empty", update.Message.From.UserName)
				return
			}

			log.Printf("Username: %s, text: %s", update.Message.From.UserName, update.Message.Text)

			b.processVideo(update)
		}()

	}
	return nil
}
