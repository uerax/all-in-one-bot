package main

import (
	"github.com/uerax/all-in-one-bot/config"
	"github.com/uerax/all-in-one-bot/tg"
)


func main() {
	config.Load()
	tg.Server()
}