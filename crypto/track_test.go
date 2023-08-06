package crypto

import (
	"fmt"
	"testing"

	"github.com/uerax/goconf"
)

func TestNewTrack(t *testing.T) {
	goconf.LoadConfig("../all-in-one-bot.yml")
	t2 := NewTrack()
	fmt.Println(t2.getEthByHtml("0xac3e9ea8b22266287e76d4eb5190d321cc9dbdd5e88f7583cb20f4620b53ea5c", false))
	

}
