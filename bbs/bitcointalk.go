package bbs

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
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
	filter map[string]struct{}
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
		filter: make(map[string]struct{}, 0),
		path: goconf.VarStringOrDefault("/usr/local/share/aio/", "bbs", "path"),
		running: false,
	}
	go b.FilterFill()
	b.old = b.Recover()
	if len(b.old) == 0 {
		b.Monitor()
	}
	b.notifi = true
	
	go b.CronDump()

	return b

}

func (b *Bitcointalk) FilterFill() {
	
	// 下载JSON文件
	url := "https://raw.githubusercontent.com/uerax/all-in-one-bot/v2/bbs/bitcointalk/filter.json"
	resp, err := http.Get(url)
	if err != nil {
		log.Println("无法下载JSON文件:", err)
		return
	}
	defer resp.Body.Close()

	// 读取JSON文件内容
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println("无法读取JSON文件:", err)
		return
	}

	// 解析JSON为Map
	filter := make(map[string]struct{})
	err = json.Unmarshal(body, &filter)
	if err != nil {
		log.Println("无法解析JSON:", err)
		return
	}

	// 使用解析后的Map进行操作
	b.filter = filter
}

func (b *Bitcointalk) CronDump() {
	ticker := time.NewTicker(1 * time.Hour)
	for range ticker.C {
		go b.Dump()
		go b.FilterFill()
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
	body, err := os.ReadFile(goconf.VarStringOrDefault("/usr/local/share/aio/", "bbs", "path") + "bitcointalk.json")
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
		topic := ""
		link, exists := td.Attr("href")
		if exists && strings.Contains(link, "topic=") {
			idx := strings.Index(link, "topic=")
			topic = link[idx+6:]
		}
		if td.Text() != "" {
			text := strings.TrimSpace(td.Text())
			if _, ok := b.old[topic]; !ok {
				b.old[topic] = struct{}{}
				for k := range b.filter {
					if strings.Contains(strings.ToLower(text), strings.ToLower(k)) {
						return
					}
				}
				if b.notifi {
					reply := strings.TrimSpace(s.Find("td").Eq(4).Text())
					views := strings.TrimSpace(s.Find("td").Eq(5).Text())
					rpy, _ := strconv.Atoi(reply)
					if rpy < 5 {
						url, exists := td.Attr("href")
						if exists {
							b.C <- "Bitcointalk 新帖推送:\n主 题: *" + text + "*\n回复: *" + reply + "*\n点击: *" + views + "*\n直达链接: " + url
						}	
					}
				}			
			}
		}
	})
}

func (b *Bitcointalk) Stop() {
	b.cancel()
}