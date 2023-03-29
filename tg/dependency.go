package tg

import (
	"tg-aio-bot/crypto"
)

var api = &aio{}

type aio struct {
	CryptoApi *crypto.Monitor
}

func (t *aio) New() {
	t.CryptoApi = crypto.NewCryptoMonitor()
}