package telegram

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"os"
	"regexp"
	"strconv"
	"time"
	"yttgmem/config"
	"yttgmem/ytVideoMaker"
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

func (b *Bot) processUpdates(updates tgbotapi.UpdatesChannel) error {
	for update := range updates {
		go func() {
			if update.Message == nil {
				return
			}

			if update.Message.Text == "/start" || update.Message.Text == "/help" {

				msgConfig := tgbotapi.MessageConfig{
					BaseChat: tgbotapi.BaseChat{
						ChatID:           update.Message.Chat.ID,
						ReplyToMessageID: 0,
					},
					Text: "This bot convert YouTube video to Video\\-Message circle by link to video and time " +
						"start and time end markers\n\n" +
						"```\nYT-link [time-start [duration]]\n```\n*Examples:*\n" +
						"```\nhttps://youtu.be/dQw4w9WgXcQ 00:00:43 00:00:10\n```For rick roll from 43 to 53 seconds\n" +
						"```\nhttps://www.youtube.com/watch?v=dQw4w9WgXcQ 00:00:43 10\n```For rick roll from 43 to 53 seconds\n" +
						"```\nhttps://www.youtube.com/watch?v=dQw4w9WgXcQ 43\n```For rick roll from 43 to 103 seconds (60 seconds by default and max)\n" +
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

			if update.Message.Text == "" {
				log.Println("Text is empty")
				return
			}

			log.Printf("Username: %s, text: %s", update.Message.From.UserName, update.Message.Text)

			re, _ := regexp.Compile(`(https://[\w.]*youtu[\w./?=_&-]+)(?: ([\d:]+))?(?: ([\d:]+))?`)
			res := re.FindAllStringSubmatch(update.Message.Text, -1)

			log.Printf("Text matched: %v", res)

			if len(res) == 0 {
				log.Printf("Text not matched: %v", update.Message.Text)
				return
			}

			parsed := res[0]

			log.Printf("Parsed: %v", parsed[1:])

			videoUrl := parsed[1]
			timeStart := 0
			timeEnd := 60

			if parsed[2] != "" {
				duration, err := time.Parse(time.TimeOnly, parsed[2])
				if err != nil {
					log.Printf("Failed to parse time: %v", err)
					intParsed, err := strconv.Atoi(parsed[2])
					if err != nil {
						log.Printf("Failed to parse int: %v", err)
						intParsed = 0
					}
					duration = time.Date(0, 0, 0, 0, 0, intParsed, 0, time.UTC)
				}
				timeStart = duration.Second()
			}

			if parsed[3] != "" {
				duration, err := time.Parse(time.TimeOnly, parsed[3])
				if err != nil {
					log.Printf("Failed to parse time: %v", err)
					intParsed, err := strconv.Atoi(parsed[3])
					if err != nil {
						log.Printf("Failed to parse int: %v", err)
						intParsed = 60
					}
					duration = time.Date(0, 0, 0, 0, 0, intParsed, 0, time.UTC)
				}
				timeEnd = duration.Second()
			}
			log.Printf("Parsed: url %s, start %d, end %d", videoUrl, timeStart, timeEnd)

			title, destFolder, err := ytVideoMaker.DownloadVideoAndAudio(videoUrl, config.Config.StoragePath, timeStart, timeEnd)
			if err != nil {
				log.Printf("Failed to download video: %v", err)
				errorMessage := fmt.Sprintf("%s", err)
				_, err = b.bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf("Error: \n```\n%v\n```", errorMessage[len(errorMessage)-4000:])))
				if err != nil {
					log.Printf("Failed to send error message: %v", err)
				}
				return
			}
			videoCoinfig := tgbotapi.NewVideoNote(update.Message.Chat.ID, 7, tgbotapi.FilePath(destFolder+"/output.mp4"))
			msg, err := b.bot.Send(videoCoinfig)
			if err != nil {
				log.Printf("Failed to send video: %v", err)
			} else {
				b.bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, title))
			}
			log.Printf("Video sent: %v", msg)

			err = os.RemoveAll(destFolder)
			if err != nil {
				log.Printf("Failed to remove file: %v", err)
			}
		}()

	}
	return nil
}
