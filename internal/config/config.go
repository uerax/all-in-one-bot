package config

func LoadConfig() *Config {
	config := &Config{}

	telegram := &Telegram{
		Token:   strOrDefault("TELEGRAM_TOKEN", ""),
		Timeout: int64OrDefault("TELEGRAM_TIMEOUT", 10),
	}

	config.Telegram = *telegram

	bitcointalk := &Bitcointalk{
		Limit:    intOrDefault("BITCOINTALK_LIMIT", 60),
		Url:      strOrDefault("BITCOINTALK_URL", "https://bitcointalk.org/index.php?board=159.0"),
		Interval: intOrDefault("BITCOINTALK_INTERVAL", 60),
	}

	config.Bitcointalk = *bitcointalk

	database := &Database{
		Type:     strOrDefault("DATABASE_TYPE", "file"),
		FilePath: strOrDefault("DATABASE_FILE_PATH", "https://raw.githubusercontent.com/uerax/all-in-one-bot/refs/heads/lite/data"),
	}

	config.Database = *database

	nodeseek := &Nodeseek{
		Url:      strOrDefault("NODESEEK_URL", "https://rss.nodeseek.com/"),
		Interval: intOrDefault("NODESEEK_INTERVAL", 60),
	}

	config.Nodeseek = *nodeseek

	return config
}

type Config struct {
	Telegram    Telegram
	Bitcointalk Bitcointalk
	Database    Database
	Nodeseek    Nodeseek
}

type Database struct {
	Type     string
	FilePath string
}

type Bitcointalk struct {
	Limit    int
	Url      string
	Interval int
}

type Telegram struct {
	Token   string
	Timeout int64
}

type Nodeseek struct {
	Limit    int
	Url      string
	Interval int
}
