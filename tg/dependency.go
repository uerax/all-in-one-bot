package tg

import (
	"tg-aio-bot/crypto"

	"github.com/uerax/goconf"
)

var api = &aio{}

type aio struct {
	cryptoApi *crypto.Crypto
}

func (t *aio) New() {
	t.cryptoApi = crypto.NewCrypto(goconf.VarStringOrDefault("", "crypto", "binance", "apiKey"),goconf.VarStringOrDefault("", "crypto", "binance", "secretKey"))
}