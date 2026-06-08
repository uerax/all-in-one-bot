package bbs

import (
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/uerax/all-in-one-bot/common"
)

type Nodeseek struct {
	ch      chan<- common.AioEvent
	latest  int64
	keyword []string
}

func NewNodeseek(ch ...chan<- common.AioEvent) *Nodeseek {
	var eventCh chan<- common.AioEvent
	if len(ch) > 0 {
		eventCh = ch[0]
	}
	return &Nodeseek{
		ch:      eventCh,
		latest:  0,
		keyword: []string{"升级码", "抽奖"},
	}
}

func (t *Nodeseek) sendMarkdown(msg string) {
	common.Send(t.ch, common.Markdown(msg, true))
}

func (t *Nodeseek) AddKeyword(keywd string) {
	t.keyword = append(t.keyword, keywd)
}

func (t *Nodeseek) ShowKeyword() {
	keywd := fmt.Sprintf("%v", t.keyword)
	t.sendMarkdown(keywd)
	fmt.Print(keywd)
}

func (t *Nodeseek) Monitor() {
	t.sendMarkdown("已开启监控Nodeseek新帖")
	tick := time.NewTicker(time.Minute)
	for range tick.C {
		t.nodeseek()
	}
}

func (t *Nodeseek) nodeseek() {
	type Item struct {
		Title       string `xml:"title"`
		Description string `xml:"description"`
		Link        string `xml:"link"`
		PubDate     string `xml:"pubDate"`
		Guid        int64  `xml:"guid"`
	}
	type Channel struct {
		Item []*Item `xml:"item"`
	}
	type NodeseekResp struct {
		XMLName xml.Name `xml:"rss"`
		Version string   `xml:"version,attr"`
		Channel *Channel `xml:"channel"`
	}

	r, err := http.Get("https://rss.nodeseek.com")
	if err != nil {
		return
	}
	defer r.Body.Close()

	body, err := io.ReadAll(r.Body)
	if err != nil {
		return
	}

	bbs := NodeseekResp{}
	if err := xml.Unmarshal(body, &bbs); err != nil {
		return
	}
	if bbs.Channel == nil || len(bbs.Channel.Item) == 0 {
		return
	}

	msg := ""
	latest := t.latest
	for _, v := range bbs.Channel.Item {
		if t.latest >= v.Guid {
			continue
		}
		if latest < v.Guid {
			latest = v.Guid
		}
		v.Title = strings.ToLower(strings.TrimSpace(v.Title))
		for _, k := range t.keyword {
			if strings.Contains(v.Title, k) {
				msg += fmt.Sprintf("[%s](%s)\n", v.Title, v.Link)
				break
			}
		}
	}

	if latest > t.latest {
		t.latest = latest
	}
	if msg != "" {
		t.sendMarkdown("*NodeSeek新帖:*\n" + msg)
	}
}
