package telegram

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
	"yttgmem/config"
	"yttgmem/ytVideoMaker"
)

type UserLinkRequest struct {
	videoUrl string
	start    float64
	duration float64
}

func parseTimeString(input string) (float64, error) {
	if strings.Contains(input, ":") {
		parts := strings.Split(input, ".")
		timePart := parts[0]

		t, err := time.Parse("15:04:05", timePart)
		if err != nil {
			return 0, fmt.Errorf("invalid time format: %v", err)
		}

		seconds := float64(t.Hour()*3600 + t.Minute()*60 + t.Second())

		// Handle milliseconds if present
		if len(parts) > 1 {
			// Pad or trim milliseconds to 3 digits
			msStr := parts[1]
			if len(msStr) > 3 {
				msStr = msStr[:3]
			} else {
				msStr = msStr + strings.Repeat("0", 3-len(msStr))
			}

			ms, err := strconv.Atoi(msStr)
			if err != nil {
				return 0, fmt.Errorf("invalid milliseconds: %v", err)
			}
			seconds += float64(ms) / 1000.0
		}

		return seconds, nil
	}

	// Handle plain number (seconds or seconds.milliseconds)
	num, err := strconv.ParseFloat(input, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid number format: %v", err)
	}

	return num, nil
}

func parseUserLink(text string) (*UserLinkRequest, error) {
	re, _ := regexp.Compile(`(https://[\w.]*youtu[\w./?=_&-]+)(?: ([\d.:]+))?(?: ([\d.:]+))?`)
	res := re.FindAllStringSubmatch(text, -1)

	log.Printf("text matched: %v", res)

	if len(res) == 0 {
		return nil, fmt.Errorf("text not matched: %v", text)
	}

	parsed := res[0]

	log.Printf("parsed: %v", parsed[1:])

	videoUrl := parsed[1]
	timeStart := 0.0
	duration := 60.0

	if parsed[2] != "" {
		newStart, err := parseTimeString(parsed[2])
		if err != nil {
			log.Printf("Failed to parse time: %v", err)
			newStart = 0
		}
		timeStart = newStart
	}

	if parsed[3] != "" {
		newDuration, err := parseTimeString(parsed[3])
		if err != nil {
			log.Printf("Failed to parse duration: %v", err)
			newDuration = 60
		}
		duration = newDuration
	}
	log.Printf("Parsed: url %s, start %d, end %d", videoUrl, timeStart, duration)
	return &UserLinkRequest{videoUrl, timeStart, duration}, nil
}

func (b *Bot) processVideo(update tgbotapi.Update) {
	videoParams, err := parseUserLink(update.Message.Text)
	if err != nil {
		log.Printf("Failed to parse user link: %v", err)
		b.bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Unable to recognize link"))
		return
	}

	title, destFolder, err := ytVideoMaker.DownloadVideoAndAudio(videoParams.videoUrl, config.Config.StoragePath, videoParams.start, videoParams.duration)
	if err != nil {
		b.sendErrorMessage(update.Message.Chat.ID, err)
		return
	}
	defer os.RemoveAll(destFolder)

	videoConfig := tgbotapi.NewVideoNote(update.Message.Chat.ID, 7, tgbotapi.FilePath(destFolder+"/output.mp4"))
	msg, err := b.bot.Send(videoConfig)

	if err != nil {
		log.Printf("Failed to send video: %v", err)
		b.sendErrorMessage(update.Message.Chat.ID, err)
	} else {
		b.bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, title))
	}

	log.Printf("Video sent: %v", msg)

}
