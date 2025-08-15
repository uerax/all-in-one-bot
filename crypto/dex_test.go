package crypto

import (
	"testing"

	"github.com/uerax/all-in-one-bot/config"
)


func TestNewScan(t *testing.T) {
	config.Load("../all-in-one-bot.yml")
	c := NewScan()
	//c.Tokentx("0x0C03Ce270B4826Ec62e7DD007f0B716068639F7B", "0x98F3706C10f91bA060348564D78c887011C36B4C", "base")
	c.GetTransactionReceipt("0x3bc00a4a5bfcdc3cf53be00d666e0032d2032417340fa2171dfc15d001876509", "base")
}
