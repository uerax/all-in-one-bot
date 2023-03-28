package main

import (
	"tg-aio-bot/config"
	"tg-aio-bot/tg"
)


func main() {
	config.Load()
	tg.Server()
}