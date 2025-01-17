package config

import (
	"github.com/uerax/goconf"
)

type Config struct {
	Telegram struct{
		Token string `yaml:"token"`
		ChatId int64 `yaml:"chatId"`
	} `yaml:"telegram"`

	Etherscan struct{
		ApiKey string `yaml:"apiKey"`
	} `yaml:"etherscan"`
}

func Load(path string) {
	goconf.LoadConfig(path)
}