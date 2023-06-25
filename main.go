package main

import (
	"flag"
	"fmt"

	"github.com/uerax/all-in-one-bot/config"
	"github.com/uerax/all-in-one-bot/tg"
)

var (
	version = "aio version: aio/0.0.33"
)

func main() {
	path := flag.String("c", "all-in-one-bot.yml", "项目的配置文件地址(使用绝对路径) 例: -c /etc/all-in-one-bot.yml")
	v := flag.Bool("v", false, "返回当前版本")
	flag.Parse()
	if *v {
		fmt.Println(version)
		return
	}
	defer func ()  {
		if err := recover(); err != nil {
			fmt.Printf("Runtime panic caught: %v\n", err)
		}
	}()
	config.Load(*path)
	tg.Server()
}