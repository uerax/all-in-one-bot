package crypto

import (
	"context"
	"fmt"
	"time"

	"github.com/uerax/goconf"
)

type line struct {
	LowPrice   string
	HighPrice  string
	LowNotify  int64
	HighNotify int64
}

type Probe struct {
	HighLine  map[string]*line
	LowLine   map[string]*line
	C         chan map[string]string
	Kline     chan string
	frequency int64
	api       *Crypto
	task      map[string]context.CancelFunc
}

func NewProbe() *Probe {
	return &Probe{
		HighLine:  make(map[string]*line),
		LowLine:   make(map[string]*line),
		C:         make(chan map[string]string),
		Kline:     make(chan string, 5),
		frequency: goconf.VarInt64OrDefault(600, "crypto", "monitor", "frequency"),
		api:       NewCrypto(goconf.VarStringOrDefault("", "crypto", "binance", "apiKey"), goconf.VarStringOrDefault("", "crypto", "binance", "secretKey")),
		task:      make(map[string]context.CancelFunc),
	}
}

func (t *Probe) AddKLineProbe(crypto string) {
	if _, ok := t.task[crypto]; !ok {
		ctx, cf := context.WithCancel(context.Background())
		t.task[crypto] = cf
		go t.KLineProbe(crypto, ctx)
		t.Kline <- fmt.Sprintf("永续合约: %s 监控已启动", crypto)
	}
}

func (t *Probe) StopKLineProbe(crypto string) {
	if v, ok := t.task[crypto]; ok {
		v()
		delete(t.task, crypto)
	}
}

func (t *Probe) KLineProbe(crypto string, ctx context.Context) {
	now := time.Now()
	p := now.Minute() / 15
	instance := 15*(p+1)*60 - now.Minute()*60 - now.Second() - 30
	fmt.Printf("距离15分钟系数还有%d", now.Add(time.Duration(instance*int(time.Second))).Unix())
	time.Sleep(time.Duration(instance) * time.Second)

	do := func() {
		kline := t.api.UFutureKline("15m", 3, crypto)
		if kline == 3 || kline == -3 {
			t.Kline <- fmt.Sprintf("永续合约: %s 连续三根15m的K线走势一致", crypto)
		}
	}
	do()

	ticker := time.NewTicker(15 * time.Minute)

	for {
		select {
		case <-ctx.Done():
			t.Kline <- fmt.Sprintf("永续合约: %s 监控已关闭", crypto)
			return
		case <-ticker.C:
			do()
		}
	}
}

// 以下为改造功能,由于目前已经不需要所以停止改造
func (t *Probe) AddHighLine(crypto string, price string) {
	if _, ok := t.HighLine[crypto]; !ok {
		t.HighLine[crypto] = &line{}
	}
	t.HighLine[crypto].HighPrice = price
	t.HighLine[crypto].HighNotify = 0
}

func (t *Probe) AddLowLine(crypto string, price string) {
	if _, ok := t.LowLine[crypto]; !ok {
		t.LowLine[crypto] = &line{}
	}
	t.LowLine[crypto].LowPrice = price
	t.LowLine[crypto].LowNotify = 0
}

func (t *Probe) RmAllLine() {
	t.HighLine = make(map[string]*line)
	t.LowLine = make(map[string]*line)
}

func (t *Probe) RmLowLine(crypto string) {
	delete(t.LowLine, crypto)
}

func (t *Probe) RmHighLine(crypto string) {
	delete(t.HighLine, crypto)
}

func (t *Probe) RmLine(crypto string) {
	t.RmHighLine(crypto)
	t.RmLowLine(crypto)
}
