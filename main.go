package main

import (
	"errors"
	"go-simpler.org/env"
	"log"
	"os"
	"ytubecircles/config"
	"ytubecircles/telegram"
	"ytubecircles/ytVideoMaker"
)

func main() {
	if err := env.Load(config.Config, nil); err != nil {
		log.Fatal(err)
	}
	err := os.RemoveAll(config.Config.StoragePath)

	if err != nil {
		panic(err)
	}
	if _, err = os.Stat(config.Config.StoragePath); errors.Is(err, os.ErrNotExist) {
		err = os.Mkdir(config.Config.StoragePath, os.ModePerm)

		if err != nil {
			panic(err)
		}
	}

	err = ytVideoMaker.InitClient(config.Config.ProxyUrl)

	if err != nil {
		panic(err)
	}

	log.Println("client init")

	bot, err := telegram.NewBot(config.Config.TelegramBotToken)

	if err != nil {
		panic(err)
	}

	err = bot.Start()
	if err != nil {
		panic(err)
	}

}
