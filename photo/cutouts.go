package photo

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"

	"github.com/uerax/goconf"
)

type Cutouts struct {
	rmbgKey string
	pixianKey string
	path string
	C chan map[int64]string
}

func NewCutouts() *Cutouts {
	return &Cutouts{
		rmbgKey: goconf.VarStringOrDefault("", "photo", "removebg", "apikey"),
		pixianKey: goconf.VarStringOrDefault("", "photo", "pixian", "authorization"),
		C: make(chan map[int64]string, 5),
		path: goconf.VarStringOrDefault("", "photo", "path"),
	}
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
		fmt.Println(err)
		return
	}
	_, err = formField.Write([]byte(uri))

	writer.Close()

	filename := fmt.Sprintf("%suser_%d.jpg", t.path, id)
	if _, err := os.Stat(t.path); os.IsNotExist(err) { // 检查目录是否存在
		err := os.Mkdir(t.path, os.ModePerm) // 创建目录
		if err != nil {
			fmt.Println("创建本地临时文件失败")
			return
		}
	}

	r, err := http.NewRequest(http.MethodPost, "https://api.pixian.ai/api/v1/remove-background", form)
	if err != nil {
		fmt.Println("Remove Background请求生成失败")
		return
	}

	r.Header.Add("Authorization", t.pixianKey)
	r.Header.Set("Content-Type", writer.FormDataContentType())
	

	resp, err := http.DefaultClient.Do(r)
	if err != nil {
		fmt.Println("Remove Background请求发送失败")
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return
	}

	file, err := os.Create(filename) // 创建本地文件
	if err != nil {
		return
	}
	defer file.Close()

	_, err = io.Copy(file, resp.Body) // 将响应体中的数据写入文件
	if err != nil {
		return
	}

	t.C <- map[int64]string{
		id:filename,
	}
	
}

func (t *Cutouts) removebg(id int64, uri string) {
	form := new(bytes.Buffer)
	writer := multipart.NewWriter(form)
	formField, err := writer.CreateFormField("image_url")
	if err != nil {
		fmt.Println(err)
		return
	}
	_, err = formField.Write([]byte(uri))

	formField, err = writer.CreateFormField("size")
	if err != nil {
		fmt.Println(err)
		return
	}
	_, err = formField.Write([]byte("auto"))

	writer.Close()

	filename := fmt.Sprintf("%suser_%d.jpg", t.path, id)
	if _, err := os.Stat(t.path); os.IsNotExist(err) { // 检查目录是否存在
		err := os.Mkdir(t.path, os.ModePerm) // 创建目录
		if err != nil {
			fmt.Println("创建本地临时文件失败")
			return
		}
	}
	
	r, err := http.NewRequest(http.MethodPost, "https://api.remove.bg/v1.0/removebg", form)
	if err != nil {
		fmt.Println("Remove Background请求生成失败")
		return
	}

	r.Header.Set("X-API-Key", t.rmbgKey)
	r.Header.Set("Content-Type", writer.FormDataContentType())
	

	resp, err := http.DefaultClient.Do(r)
	if err != nil {
		fmt.Println("Remove Background请求发送失败")
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return
	}

	file, err := os.Create(filename) // 创建本地文件
	if err != nil {
		return
	}
	defer file.Close()

	_, err = io.Copy(file, resp.Body) // 将响应体中的数据写入文件
	if err != nil {
		return
	}

	t.C <- map[int64]string{
		id:filename,
	}
}