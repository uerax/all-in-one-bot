package utils

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/uerax/all-in-one-bot/common"
	"github.com/uerax/all-in-one-bot/crypto"
)

type Utils struct {
	format string
	ch     chan<- common.AioEvent
	bn     *crypto.Crypto
}

func NewUtils(ch ...chan<- common.AioEvent) *Utils {
	var eventCh chan<- common.AioEvent
	if len(ch) > 0 {
		eventCh = ch[0]
	}
	return &Utils{
		format: "2006-01-02 15:04:05",
		ch:     eventCh,
	}
}

func (t *Utils) sendMarkdown(msg string) {
	common.Send(t.ch, common.Markdown(msg, false))
}

func (t *Utils) sendText(msg string) {
	common.Send(t.ch, common.Text(msg))
}

func (t *Utils) Base64Encode(str string) {
	encoded := base64.StdEncoding.EncodeToString([]byte(str))
	t.sendMarkdown(fmt.Sprintf("`%s`", encoded))
}

func (t *Utils) Base64Decode(str string) {
	decoded, err := base64.StdEncoding.DecodeString(str)
	if err != nil {
		log.Println(err)
		t.sendText("解码失败")
		return
	}
	t.sendMarkdown(fmt.Sprintf("`%s`", decoded))
}

func (t *Utils) TimestampConvert(Timestamp string) {
	ts, err := strconv.ParseInt(Timestamp, 10, 64)
	if err != nil {
		log.Println(err)
		t.sendText("时间戳格式有误")
		return
	}
	date := time.Unix(ts, 0)

	t.sendMarkdown(fmt.Sprintf("`%s`", date.Format(t.format)))
}

func (t *Utils) TimeConvert(date string) {
	var timestamp int64
	if date != "" && !strings.EqualFold(date, "now") {
		ts, err := time.ParseInLocation(t.format, date, time.Local)
		if err != nil {
			log.Println(err)
			t.sendText("时间格式有误")
			return
		}
		timestamp = ts.Unix()
	} else {
		timestamp = time.Now().Unix()
	}

	t.sendMarkdown(fmt.Sprintf("`%d`", timestamp))
}

func (t *Utils) JsonFormat(str string) {
	var out bytes.Buffer
	err := json.Indent(&out, []byte(str), "", "    ")
	if err != nil {
		log.Println(err)
		t.sendText("格式化失败")
		return
	}
	t.sendMarkdown(fmt.Sprintf("`%s`", out.String()))
}

func (t *Utils) RewardCal(h, d, r, time, val string) {
	hash, err := strconv.ParseFloat(h, 64)
	if err != nil {
		return
	}
	diff, err := strconv.ParseFloat(d, 64)
	if err != nil {
		return
	}
	reward, err := strconv.ParseFloat(r, 64)
	if err != nil {
		return
	}
	hour, err := strconv.ParseFloat(time, 64)
	if err != nil {
		return
	}
	value, err := strconv.ParseFloat(val, 64)
	if err != nil {
		return
	}
	cnt := 0.0
	if diff < 1 {
		cnt = hash * math.Pow(2, 10) / (diff * math.Pow(2, 32)) * reward * 60 * 60 * hour * value
	} else {
		cnt = hash * math.Pow(2, 10) / (diff) * reward * 60 * 60 * hour * value
	}

	t.sendMarkdown(fmt.Sprintf("`%.10f`", cnt))
}
