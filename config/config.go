package config

import (
	"flag"

	"github.com/uerax/goconf"
)

var (
	path = flag.String("c", "all-in-one-bot.yml", "项目的配置文件地址(使用绝对路径) 例: -c /etc/all-in-one-bot.yml")
)

func Load() {
	flag.Parse()
	goconf.LoadConfig(*path)
}