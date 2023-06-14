package lists

import (
	"strings"

	"github.com/uerax/goconf"
)

type Lists struct {
	crypto []interface{}
	video []interface{}
	image []interface{}
	utils []interface{}
	list []interface{}
	C chan string
	ErrC chan string
}

func NewLists() *Lists {
	_crypto, _ := goconf.VarArray("command", "crypto")
	_video, _ := goconf.VarArray("command", "video")
	_image, _ := goconf.VarArray("command", "image")
	_utils, _ := goconf.VarArray("command", "utils")
	_list, _ := goconf.VarArray("command", "list")
	return &Lists{
		crypto: _crypto,
		video: _video,
		image: _image,
		utils: _utils,
		list: _list,
		C: make(chan string, 3),
		ErrC: make(chan string, 3),
	}
}

func (t *Lists) Crypto() {
	b := strings.Builder{}
	b.WriteString("加密货币相关命令:")
	for _, v := range t.crypto {
		b.WriteString("\n`/")
		b.WriteString(strings.ReplaceAll(v.(string), " -", "` -"))
	}
	t.C <- b.String()
}

func (t *Lists) Image() {
	b := strings.Builder{}
	b.WriteString("图片处理相关命令:")
	for _, v := range t.image {
		b.WriteString("\n`/")
		b.WriteString(strings.ReplaceAll(v.(string), " -", "` -"))
	}
	t.C <- b.String()
}
func (t *Lists) List() {
	b := strings.Builder{}
	b.WriteString("命令列表相关命令:")
	for _, v := range t.list {
		b.WriteString("\n`/")
		b.WriteString(strings.ReplaceAll(v.(string), " -", "` -"))
	}
	t.C <- b.String()
}
func (t *Lists) Utils() {
	b := strings.Builder{}
	b.WriteString("工具类相关命令:")
	for _, v := range t.utils {
		b.WriteString("\n`/")
		b.WriteString(strings.ReplaceAll(v.(string), " -", "` -"))
	}
	t.C <- b.String()
}

func (t *Lists) Video() {
	b := strings.Builder{}
	b.WriteString("视频相关命令:")
	for _, v := range t.video {
		b.WriteString("\n`/")
		b.WriteString(strings.ReplaceAll(v.(string), " -", "` -"))
	}
	t.C <- b.String()
}

func (t *Lists) All() {
	b := strings.Builder{}
	b.WriteString("所有命令:")
	for _, v := range t.crypto {
		b.WriteString("\n`/")
		b.WriteString(strings.ReplaceAll(v.(string), " -", "` -"))
	}
	for _, v := range t.image {
		b.WriteString("\n`/")
		b.WriteString(strings.ReplaceAll(v.(string), " -", "` -"))
	}
	for _, v := range t.list {
		b.WriteString("\n`/")
		b.WriteString(strings.ReplaceAll(v.(string), " -", "` -"))
	}
	for _, v := range t.utils {
		b.WriteString("\n`/")
		b.WriteString(strings.ReplaceAll(v.(string), " -", "` -"))
	}
	for _, v := range t.video {
		b.WriteString("\n`/")
		b.WriteString(strings.ReplaceAll(v.(string), " -", "` -"))
	}
	
	t.C <- b.String()
}

