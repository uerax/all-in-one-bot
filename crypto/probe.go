package crypto

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/uerax/goconf"
)

const ()

type line struct {
	LowPrice   string
	HighPrice  string
	LowNotify  int64
	HighNotify int64
}

type Probe struct {
	HighLine      map[string]*line
	LowLine       map[string]*line
	C             chan map[string]string
	Kline         chan string
	Meme          chan string
	frequency     int64
	api           *Crypto
	task          map[string]context.CancelFunc
	memeHighTask  map[string]context.CancelFunc
	memeLowTask   map[string]context.CancelFunc
	smartAddr     map[string]context.CancelFunc
	smartBuys     map[string]map[string]struct{}
	smartItv      int
	smartDumpPath string
	Keys		  *PollingKey
}

func NewProbe() *Probe {
	p := &Probe{
		HighLine:      make(map[string]*line),
		LowLine:       make(map[string]*line),
		C:             make(chan map[string]string),
		Kline:         make(chan string, 5),
		Meme:          make(chan string, 5),
		frequency:     goconf.VarInt64OrDefault(600, "crypto", "monitor", "frequency"),
		api:           NewCrypto(goconf.VarStringOrDefault("", "crypto", "binance", "apiKey"), goconf.VarStringOrDefault("", "crypto", "binance", "secretKey")),
		task:          make(map[string]context.CancelFunc),
		memeLowTask:   make(map[string]context.CancelFunc),
		memeHighTask:  make(map[string]context.CancelFunc),
		smartAddr:     make(map[string]context.CancelFunc),
		smartBuys:     recoverSmartAddrList(),
		smartItv:      goconf.VarIntOrDefault(30, "crypto", "etherscan", "interval"),
		smartDumpPath: goconf.VarStringOrDefault("/usr/local/share/aio/", "crypto", "etherscan", "path"),
		Keys: NewPollingKey(),
	}

	go p.DumpCron()

	return p
}

func (t *Probe) ListKLineProbe() string {
	b := strings.Builder{}
	b.WriteString("当前正在探测的加密货币有:")
	for k := range t.task {
		b.WriteString("\n`")
		b.WriteString(k)
		b.WriteString("`")
	}
	return b.String()
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
		t.Kline <- fmt.Sprintf("永续合约: %s 监控已关闭", crypto)
	}
}

func (t *Probe) KLineProbe(crypto string, ctx context.Context) {
	now := time.Now()
	p := now.Minute() / 15
	instance := 15*(p+1)*60 - now.Minute()*60 - now.Second() - 60
	fmt.Printf("距离15分钟系数还有%d", now.Add(time.Duration(instance*int(time.Second))).Unix())
	time.Sleep(time.Duration(instance) * time.Second)

	do := func() {
		kline := t.api.UFutureKline("15m", 5, crypto)

		if len(kline) == 5 && ((kline[4]+kline[3]+kline[2] == -3 && kline[0]+kline[1] != -2) || (kline[4]+kline[3]+kline[2] == 3 && kline[0]+kline[1] != 2)) {
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

