package config

func LoadConfig() *Config {
	config := &Config{}

	telegram := &Telegram{
		Token: strOrDefault("TELEGRAM_TOKEN", ""),
		Timeout: int64OrDefault("TELEGRAM_TIMEOUT", 10),
	}

	config.Telegram = *telegram

	bitcointalk := &Bitcointalk{
		Limit: intOrDefault("BITCOINTALK_LIMIT", 60),
		Url: strOrDefault("BITCOINTALK_URL", "https://bitcointalk.org/index.php?board=159.0"),
		Interval: intOrDefault("BITCOINTALK_INTERVAL", 60),
	}

	config.Bitcointalk = *bitcointalk

	return config
}

type Config struct {
	Telegram Telegram
	Bitcointalk Bitcointalk
}

type Bitcointalk struct {
	Limit int
	Url string
	Interval int
}

type Telegram struct {
	Token string
	Timeout int64
}


