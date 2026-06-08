package photo

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"

	"github.com/uerax/all-in-one-bot/common"
	"github.com/uerax/goconf"
)

type Cutouts struct {
	rmbgKey   string
	pixianKey string
	path      string
	ch        chan<- common.AioEvent
}

func NewCutouts(ch ...chan<- common.AioEvent) *Cutouts {
	var eventCh chan<- common.AioEvent
	if len(ch) > 0 {
		eventCh = ch[0]
	}
	return &Cutouts{
		rmbgKey:   goconf.VarStringOrDefault("", "photo", "removebg", "apikey"),
		pixianKey: goconf.VarStringOrDefault("", "photo", "pixian", "authorization"),
		ch:        eventCh,
		path:      goconf.VarStringOrDefault("/tmp/aio-tgbot/", "photo", "path"),
	}
}

func (t *Cutouts) sendError(msg string) {
	common.Send(t.ch, common.Text(msg))
}

func (t *Cutouts) sendPhoto(id int64, path string) {
	common.Send(t.ch, common.PhotoTo(id, path))
}

func (t *Cutouts) RemoveBackground(id int64, uri string) {
	if t.pixianKey != "" {
		t.pixian(id, uri)
	}
	if t.rmbgKey != "" {
		t.removebg(id, uri)
	}

}

func (t *Cutouts) pixian(id int64, uri string) {
	form := new(bytes.Buffer)
	writer := multipart.NewWriter(form)
	formField, err := writer.CreateFormField("image.url")
	if err != nil {
		log.Println(err)
		t.sendError("创建表单参数失败")
		return
	}

	_, err = formField.Write([]byte(uri))
	if err != nil {
		log.Println("表单填充异常")
		t.sendError("表单填充参数失败")
		return
	}

	writer.Close()

	filename := fmt.Sprintf("%suser_%d.jpg", t.path, id)
	if _, err := os.Stat(t.path); os.IsNotExist(err) { // 检查目录是否存在
		err := os.MkdirAll(t.path, os.ModePerm) // 创建目录
		if err != nil {
			log.Println("创建本地临时文件失败")
			t.sendError("创建本地临时文件失败")
			return
		}
	}

	r, err := http.NewRequest(http.MethodPost, "https://api.pixian.ai/api/v1/remove-background", form)
	if err != nil {
		log.Println("Remove Background请求生成失败")
		t.sendError("Remove Background请求生成失败")
		return
	}

	r.Header.Add("Authorization", t.pixianKey)
	r.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := http.DefaultClient.Do(r)
	if err != nil {
		log.Println("Remove Background请求发送失败")
		t.sendError("Remove Background请求发送失败")
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.sendError("Remove Background请求响应异常")
		return
	}

	file, err := os.Create(filename) // 创建本地文件
	if err != nil {
		t.sendError("创建本地临时文件失败")
		return
	}
	defer file.Close()

	_, err = io.Copy(file, resp.Body) // 将响应体中的数据写入文件
	if err != nil {
		t.sendError("将响应体中的数据写入文件")
		return
	}

	t.sendPhoto(id, filename)

	go common.DeleteFileAfterTime(filename, 2)

}

func (t *Cutouts) removebg(id int64, uri string) {
	form := new(bytes.Buffer)
	writer := multipart.NewWriter(form)
	formField, err := writer.CreateFormField("image_url")
	if err != nil {
		log.Println(err)
		t.sendError("创建本地临时文件失败")
		return
	}
	_, err = formField.Write([]byte(uri))
	if err != nil {
		log.Println(err)
		t.sendError("表单填充参数失败")
		return
	}

	formField, err = writer.CreateFormField("size")
	if err != nil {
		log.Println(err)
		t.sendError("创建表单失败")
		return
	}

	_, err = formField.Write([]byte("auto"))
	if err != nil {
		log.Println(err)
		t.sendError("表单填充参数失败")
		return
	}

	writer.Close()

	filename := fmt.Sprintf("%suser_%d.jpg", t.path, id)
	if _, err := os.Stat(t.path); os.IsNotExist(err) { // 检查目录是否存在
		err := os.MkdirAll(t.path, os.ModePerm) // 创建目录
		if err != nil {
			log.Println("创建本地临时文件失败")
			t.sendError("创建本地临时文件失败")
			return
		}
	}

	r, err := http.NewRequest(http.MethodPost, "https://api.remove.bg/v1.0/removebg", form)
	if err != nil {
		log.Println("Remove Background请求生成失败")
		t.sendError("Remove Background请求生成失败")
		return
	}

	r.Header.Set("X-API-Key", t.rmbgKey)
	r.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := http.DefaultClient.Do(r)
	if err != nil {
		log.Println("Remove Background请求发送失败")
		t.sendError("Remove Background请求发送失败")
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.sendError("Remove Background请求响应异常")
		return
	}

	file, err := os.Create(filename) // 创建本地文件
	if err != nil {
		t.sendError("创建本地文件失败")
		return
	}
	defer file.Close()

	_, err = io.Copy(file, resp.Body) // 将响应体中的数据写入文件
	if err != nil {
		t.sendError("将响应体中的数据写入文件失败")
		return
	}

	t.sendPhoto(id, filename)
	go common.DeleteFileAfterTime(filename, 2)
}
