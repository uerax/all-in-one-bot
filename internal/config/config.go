package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

func LoadConfig(configPath string) *Config {
	// 从 YAML 加载基础配置，随后用环境变量覆盖（env > YAML > 内置默认）。
	var base Config
	if configPath != "" {
		data, err := os.ReadFile(configPath)
		if err != nil {
			fmt.Printf("FATAL: 读取配置文件失败: %v\n", err)
			os.Exit(1)
		}
		if err := yaml.Unmarshal(data, &base); err != nil {
			fmt.Printf("FATAL: 解析配置文件失败: %v\n", err)
			os.Exit(1)
		}
	}

	config := &Config{}

	config.Telegram = Telegram{
		Token:    strOrDefaultBase("TELEGRAM_TOKEN", base.Telegram.Token, ""),
		Timeout:  int64OrDefaultBase("TELEGRAM_TIMEOUT", base.Telegram.Timeout, 10),
		AdminIDs: telegramAdminIDs("TELEGRAM_ADMIN_IDS", base.Telegram.AdminIDs),
	}

	config.Polymarket = Polymarket{
		GammaBaseURL:  strOrDefaultBase("POLYMARKET_GAMMA_BASE_URL", base.Polymarket.GammaBaseURL, "https://gamma-api.polymarket.com"),
		ClobBaseURL:   strOrDefaultBase("POLYMARKET_CLOB_BASE_URL", base.Polymarket.ClobBaseURL, "https://clob.polymarket.com"),
		Timeout:       intOrDefaultBase("POLYMARKET_TIMEOUT", base.Polymarket.Timeout, 10),
		DefaultLimit:  intOrDefaultBase("POLYMARKET_DEFAULT_LIMIT", base.Polymarket.DefaultLimit, 10),
		Address:       strOrDefaultBase("POLYMARKET_ADDRESS", base.Polymarket.Address, ""),
		ApiKey:        strOrDefaultBase("POLYMARKET_API_KEY", base.Polymarket.ApiKey, ""),
		ApiSecret:     strOrDefaultBase("POLYMARKET_API_SECRET", base.Polymarket.ApiSecret, ""),
		ApiPassphrase: strOrDefaultBase("POLYMARKET_API_PASSPHRASE", base.Polymarket.ApiPassphrase, ""),
	}

	config.Bitcointalk = Bitcointalk{
		Limit:    intOrDefaultBase("BITCOINTALK_LIMIT", base.Bitcointalk.Limit, 60),
		Url:      strOrDefaultBase("BITCOINTALK_URL", base.Bitcointalk.Url, "https://bitcointalk.org/index.php?board=159.0"),
		Interval: intOrDefaultBase("BITCOINTALK_INTERVAL", base.Bitcointalk.Interval, 60),
	}

	config.Database = Database{
		Type:     strOrDefaultBase("DATABASE_TYPE", base.Database.Type, "file"),
		FilePath: strOrDefaultBase("DATABASE_FILE_PATH", base.Database.FilePath, "https://raw.githubusercontent.com/uerax/all-in-one-bot/refs/heads/lite/data"),
	}

	config.Nodeseek = Nodeseek{
		Limit:    intOrDefaultBase("NODESEEK_LIMIT", base.Nodeseek.Limit, 60),
		Url:      strOrDefaultBase("NODESEEK_URL", base.Nodeseek.Url, "https://rss.nodeseek.com/"),
		Interval: intOrDefaultBase("NODESEEK_INTERVAL", base.Nodeseek.Interval, 60),
	}

	config.Coingecko = Coingecko{
		Keys: strOrDefaultBase("COINGECKO_KEYS", base.Coingecko.Keys, ""),
	}

	config.GeckoTerminal = GeckoTerminal{
		BaseURL: strOrDefaultBase("GECKOTERMINAL_BASE_URL", base.GeckoTerminal.BaseURL, "https://api.geckoterminal.com/api/v2"),
		Timeout: intOrDefaultBase("GECKOTERMINAL_TIMEOUT", base.GeckoTerminal.Timeout, 15),
	}

	config.Crocodile = Crocodile{
		Lookback:          intOrDefaultBase("CROCODILE_LOOKBACK", base.Crocodile.Lookback, 5),
		YesterdayMultiple: float64OrDefaultBase("CROCODILE_YESTERDAY_MULTIPLE", base.Crocodile.YesterdayMultiple, 3.0),
		AverageMultiple:   float64OrDefaultBase("CROCODILE_AVERAGE_MULTIPLE", base.Crocodile.AverageMultiple, 2.0),
		Interval:          intOrDefaultBase("CROCODILE_INTERVAL", base.Crocodile.Interval, 86400),
	}

	return config
}

type Config struct {
	Telegram      Telegram      `yaml:"telegram"`
	Polymarket    Polymarket    `yaml:"polymarket"`
	Bitcointalk   Bitcointalk   `yaml:"bitcointalk"`
	Database      Database      `yaml:"database"`
	Nodeseek      Nodeseek      `yaml:"nodeseek"`
	Coingecko     Coingecko     `yaml:"coingecko"`
	GeckoTerminal GeckoTerminal `yaml:"geckoterminal"`
	Crocodile     Crocodile     `yaml:"crocodile"`
}

type Database struct {
	Type     string `yaml:"type"`
	FilePath string `yaml:"file_path"`
}

type Polymarket struct {
	GammaBaseURL  string `yaml:"gamma_base_url"`
	ClobBaseURL   string `yaml:"clob_base_url"`
	Timeout       int    `yaml:"timeout"`
	DefaultLimit  int    `yaml:"default_limit"`
	Address       string `yaml:"address"`
	ApiKey        string `yaml:"api_key"`
	ApiSecret     string `yaml:"api_secret"`
	ApiPassphrase string `yaml:"api_passphrase"`
}

type Bitcointalk struct {
	Limit    int    `yaml:"limit"`
	Url      string `yaml:"url"`
	Interval int    `yaml:"interval"`
}

type Telegram struct {
	Token    string  `yaml:"token"`
	Timeout  int64   `yaml:"timeout"`
	AdminIDs []int64 `yaml:"admin_ids"` // 管理员 Telegram ID 白名单（逗号分隔/env，列表/YAML）；空则全部放行
}

type Nodeseek struct {
	Limit    int    `yaml:"limit"`
	Url      string `yaml:"url"`
	Interval int    `yaml:"interval"`
}

type Coingecko struct {
	Keys string `yaml:"keys"` // 逗号分隔的 demo API keys
}

type GeckoTerminal struct {
	BaseURL string `yaml:"base_url"`
	Timeout int    `yaml:"timeout"`
}

type Crocodile struct {
	Lookback          int     `yaml:"lookback"`
	YesterdayMultiple float64 `yaml:"yesterday_multiple"`
	AverageMultiple   float64 `yaml:"average_multiple"`
	Interval          int     `yaml:"interval"` // 扫描间隔秒数，默认 86400（每天）
}
