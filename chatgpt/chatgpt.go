package chatgpt

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/uerax/all-in-one-bot/common"
	"github.com/uerax/goconf"
)

type ChatGPT struct {
	apiKey string
	ch     chan<- common.AioEvent
}

func NewChatGPT(ch ...chan<- common.AioEvent) *ChatGPT {
	var eventCh chan<- common.AioEvent
	if len(ch) > 0 {
		eventCh = ch[0]
	}
	return &ChatGPT{
		apiKey: goconf.VarStringOrDefault("", "chatgpt", "key"),
		ch:     eventCh,
	}
}

type ChatGPTRequest struct {
	Model    string     `json:"model"`
	Messages []*Message `json:"messages"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type Choice struct {
	Index        int     `json:"index"`
	Message      Message `json:"message"`
	FinishReason string  `json:"finish_reason"`
}

type ChatGPTResponse struct {
	ID      string   `json:"id"`
	Object  string   `json:"object"`
	Created int64    `json:"created"`
	Choices []Choice `json:"choices"`
}

func (t *ChatGPT) Ask(id int64, msg string) {
	if t.apiKey == "" {
		return
	}
	model := goconf.VarStringOrDefault("gpt-3.5-turbo", "chatgpt", "model")

	url := "https://api.openai.com/v1/chat/completions"

	requestBody, err := json.Marshal(&ChatGPTRequest{
		Model: model,
		Messages: []*Message{
			{
				Role:    "user",
				Content: msg,
			},
		},
	})
	if err != nil {
		log.Println("请求转换成结构体异常")
		return
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(requestBody))
	if err != nil {
		log.Println("请求生成异常")
		return
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", t.apiKey))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Println("请求发送失败")
		return
	}

	body, err := io.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		log.Println("响应body读取异常")
		return
	}

	var respBody *ChatGPTResponse
	err = json.Unmarshal(body, &respBody)
	if err != nil {
		log.Println("响应转换成结构体异常")
		return
	}

	common.Send(t.ch, common.TextTo(id, respBody.Choices[0].Message.Content))

}
