package vps

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/uerax/goconf"
)

type VpsMonitor struct {
	valid  map[string]*Vps            // vps -> keyword
	VTU    sync.Map                   // map[string]map[int64]struct{} vps -> user
	C      chan map[int64]string      // user -> msg
	notify map[string]map[int64]int64 // url ->  user -> date
	ctx    context.Context
	cancel context.CancelFunc
}

type Vps struct {
	Url     string
	Keyword string
	Desc    string
	Name    string
}

func NewVpsMonitor() *VpsMonitor {
	vpsList, _ := goconf.VarArray("vps", "list")
	vps := make(map[string]*Vps)
	for _, v := range vpsList {
		vps[v.(map[string]interface{})["url"].(string)] = &Vps{
			Url:     v.(map[string]interface{})["url"].(string),
			Desc:    v.(map[string]interface{})["desc"].(string),
			Keyword: v.(map[string]interface{})["keyword"].(string),
			Name:    v.(map[string]interface{})["name"].(string),
		}
	}
	ctx, cancel := context.WithCancel(context.Background())
	return &VpsMonitor{
		valid:  vps,
		VTU:    sync.Map{},
		C:      make(chan map[int64]string, 1),
		notify: make(map[string]map[int64]int64),
		ctx:    ctx,
		cancel: cancel,
	}
}

func (t *VpsMonitor) AddMonitor(id int64, url string) {
	if _, ok := t.valid[url]; !ok {
		fmt.Println("添加监控的VPS链接不支持")
		return
	}
	vtu, _ := t.VTU.LoadOrStore(url, make(map[int64]struct{}))
	vtu.(map[int64]struct{})[id] = struct{}{}
	t.CancelAndStart()
}

func (t *VpsMonitor) CancelAndStart() {
	t.cancel()
	t.ctx, t.cancel = context.WithCancel(context.Background())
	t.Start()
}

func (t *VpsMonitor) Start() {
	ticker := time.NewTicker(5 * time.Second)

	for {
		select {
		case <-ticker.C:
			for k, v := range t.valid {
				go t.probe(k, v.Keyword)
			}
		case <-t.ctx.Done():
			return
		}
	}
}

func (t *VpsMonitor) probe(url, keyword string) {

	resp, err := http.Get(url)
	if err != nil {
		fmt.Printf("%s 访问失败\n", url)
		return
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("body解析失败")
		return
	}
	now := time.Now().Unix()
	frequency := goconf.VarIntOrDefault(600, "vps", "frequency")
	if !strings.Contains(string(b), keyword) {
		fmt.Println(string(b))
		s := make(map[int64]string)
		if value, ok := t.VTU.Load(url); ok {
			for k := range value.(map[int64]struct{}) {

				if _, ok := t.notify[url]; !ok {
					t.notify[url] = make(map[int64]int64)
				}

				if v, exist := t.notify[url][k]; exist || now-v >= int64(frequency) {
					s[k] = fmt.Sprintf("主机补货通知:\n商 品：%s \n详 情：%s \n链 接: %s", t.valid[url].Name, t.valid[url].Desc, url)
					t.notify[url][k] = now
				}
			}
			if len(s) > 0 {
				t.C <- s
			}

		}

	}

}

func (t *VpsMonitor) GetList() map[string]*Vps {
	return t.valid
}
