package crypto

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/uerax/goconf"
)

type Track struct {
	C chan string
	Newest string

}

type tx struct {
	Buy string
	Sell string
}

func (t *Track) WalletTracking(addr string) {
	apiKey := goconf.VarStringOrDefault("", "crypto", "etherscan", "apiKey")
	if apiKey == "" {
		t.C <- "未读取到etherscan的apikey无法启动监控"
		return
	}
	url := "https://api.etherscan.io/api?module=account&action=tokentx&page=1&offset=%s&sort=desc&address=%s&apikey=%s"
	r, err := http.Get(fmt.Sprintf(url, "20", addr, apiKey))
	if err != nil {
		fmt.Println("请求失败")
		t.C <- "etherscan请求失败"
		return
	}
	defer r.Body.Close()
	b, err := io.ReadAll(r.Body)
	if err != nil {
		fmt.Println("读取body失败")
		t.C <- "读取body失败"
		return
	}
	scan := new(TokenTxResp)
	err = json.Unmarshal(b, &scan)
	if err != nil {
		fmt.Println("json转换失败")
		t.C <- "json转换失败"
		return
	}

	if scan.Status != "1" {
		t.C <- "返回码不为1,检查地址是否正确"
		return
	}

}
