package config

type ConfigSchema struct {
	TelegramBotToken string `env:"TELEGRAM_BOT_TOKEN"`
	WHUrl            string `env:"WH_URL" default:""`
	ProxyUrl         string `env:"PROXY_URL" default:""`
	FFMPEGBin        string `env:"FFMPEG_BIN" default:"ffmpeg"`
	StoragePath      string `env:"STORAGE_PATH" default:"storage"`
}

var Config = &ConfigSchema{}
