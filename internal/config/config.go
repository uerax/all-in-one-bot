package config

func LoadConfig() *Config {
	config := &Config{}

	telegram := &Telegram{
		Token: strOrDefault("TELEGRAM_TOKEN", ""),
		Timeout: int64OrDefault("TELEGRAM_TIMEOUT", 10),
	}

	config.Telegram = *telegram

	return config
}

type Config struct {
	Telegram Telegram
}

type Telegram struct {
	Token string
	Timeout int64
}


