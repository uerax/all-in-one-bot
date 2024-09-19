package bbs

import (
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type Nodeseek struct {
	C chan string
	latest time.Time
	keyword []string
}

func NewNodeseek() *Nodeseek {
	k := []string{"甲骨文", "升级号", "oracle"}
	n := &Nodeseek{
		C: make(chan string),
		latest: time.Now(),
		keyword: k,
	}	
	return n
}

func (t *Nodeseek) Monitor() {
	t.C <- "已开启监控Nodeseek新帖"
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
	}
	type Channel struct {
		Item []*Item `xml:"item"`
	}
	type NodeseekResp struct {
		XMLName xml.Name `xml:"rss"`
		Version string   `xml:"version,attr"`
		Channel *Channel `xml:"channel"`
	}
	link := "https://rss.nodeseek.com"
	r, err := http.Get(link)
	bbs := NodeseekResp{}
	if err != nil {
		return 
	}
	b, _ := io.ReadAll(r.Body)
	defer r.Body.Close()
	err = xml.Unmarshal(b, &bbs)
	if err != nil {
		return 
	}
	if bbs.Channel == nil || len(bbs.Channel.Item) == 0 {
		return
	}
	msg := ""
	for _, v := range bbs.Channel.Item {
		time, err := time.Parse(time.RFC1123, v.PubDate)
		v.Title = strings.ToLower(strings.TrimSpace(v.Title))
		if err == nil {
			if t.latest.Before(time) {
				for _, k := range t.keyword {
					if strings.Contains(v.Title, k) {
						msg += fmt.Sprintf("[%s](%s)\n", v.Title, v.Link)
						break
					}
				}
			} else {
				break
			}
		}
	}
	
	if msg != "" {
		newdate, err := time.Parse(time.RFC1123, bbs.Channel.Item[0].PubDate)
		if err != nil {
			t.latest = newdate
		}
		t.C <- "*NodeSeek新帖:*\n" + msg
	}

}
