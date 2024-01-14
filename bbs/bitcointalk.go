package bbs

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/uerax/goconf"
)

type Bitcointalk struct {
	url string
	old map[string]struct{}
	ctx context.Context
	cancel context.CancelFunc
	C chan string
	notifi bool
	path string
	running bool
}

func NewBitcointalk() *Bitcointalk {
	b := &Bitcointalk{
		url: "https://bitcointalk.org/index.php?board=159.0",
		old: make(map[string]struct{}),
		C: make(chan string, 5),
		notifi: false,
		path: goconf.VarStringOrDefault("./", "bbs", "path"),
		running: false,
	}

	b.old = b.Recover()
	if len(b.old) == 0 {
		b.Monitor()
	}
	b.notifi = true
	go b.CronDump()

	return b

}

func (b *Bitcointalk) CronDump() {
	ticker := time.NewTicker(5 * time.Minute)
	for range ticker.C {
		b.Dump()
	}
}

func (b *Bitcointalk) Dump() {
	if len(b.old) == 0 {
		return
	}
	body, err := json.Marshal(b.old)
	if err != nil {
		log.Println("Bitcointalk备份序列化失败:", err)
		return
	}

	if _, err := os.Stat(b.path); os.IsNotExist(err) { // 检查目录是否存在
		err := os.MkdirAll(b.path, os.ModePerm) // 创建目录
		if err != nil {
			log.Println("创建本地文件夹失败")
			return
		}
	}
	err = os.WriteFile(b.path+"bitcointalk.json", body, 0644)
	if err != nil {
		log.Println("bitcointalk dump文件创建/写入失败")
		return
	}
}

func (b *Bitcointalk) Recover() map[string]struct{} {
	dump := make(map[string]struct{})
	body, err := os.ReadFile(goconf.VarStringOrDefault("", "bbs", "path") + "bitcointalk.json")
	if err != nil {
		return dump
	}

	err = json.Unmarshal(body, &dump)
	if err != nil {
		return dump
	}

	return dump
}

func (b *Bitcointalk) Start() {
	if b.running {
		b.C <- "已开启监控Bitcointalk新帖, 无需重复开启"
		return
	}
	b.running = true
	b.ctx, b.cancel = context.WithCancel(context.Background())
	ticker := time.NewTicker(1 * time.Minute)

	log.Println("开启定时监控Bitcointalk新帖")
	b.C <- "开启定时监控Bitcointalk新帖"

	for {
		select {
		case <-ticker.C:
			b.Monitor()
		case <-b.ctx.Done():
			log.Println("关闭定时监控Bitcointalk新帖")
			b.C <- "关闭定时监控Bitcointalk新帖"
			b.running = false
			return
		}
	}
}

func (b *Bitcointalk) Monitor() {
	r, err := http.Get(b.url)
	if err != nil {
		return
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Println("body解析失败")
		return
	}
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(body))
	if err != nil {
		log.Println("html解析失败")
		return
	}
	doc.Find("tr").Each(func(i int, s *goquery.Selection) {
		td := s.Find("td").Eq(2).Find("span").Find("a")
		if td.Text() != "" {
			if _, ok := b.old[td.Text()]; !ok {
				b.old[td.Text()] = struct{}{}
				reply := s.Find("td").Eq(4).Text()
				views := s.Find("td").Eq(5).Text()
				url, exists := td.Attr("href")
				if exists && b.notifi {
					b.C <- "Bitcointalk 新帖推送:\n主 题: " + td.Text() + "\n回复: " + reply + "\n点击: " + views + "\n直达链接: " + url
				}				
			}
		}
	})
}

func (b *Bitcointalk) Stop() {
	b.cancel()
}