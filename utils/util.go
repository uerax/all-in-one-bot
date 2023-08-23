package utils

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"
)

type Utils struct {
	format string
	MsgC   chan string
	ErrC   chan string
}

func NewUtils() *Utils {
	return &Utils{
		format: "2006-01-02 15:04:05",
		MsgC:   make(chan string, 2),
		ErrC:   make(chan string, 2),
	}
}

func (t *Utils) Base64Encode(str string) {
	encoded := base64.StdEncoding.EncodeToString([]byte(str))
	t.MsgC <- fmt.Sprintf("`%s`", encoded)
}

func (t *Utils) Base64Decode(str string) {
	decoded, err := base64.StdEncoding.DecodeString(str)
	if err != nil {
		log.Println(err)
		t.ErrC <- "解码失败"
		return
	}
	t.MsgC <- fmt.Sprintf("`%s`", decoded)
}

func (t *Utils) TimestampConvert(Timestamp string) {
	ts, err := strconv.ParseInt(Timestamp, 10, 64)
	if err != nil {
		log.Println(err)
		t.ErrC <- "时间戳格式有误"
		return
	}
	date := time.Unix(ts, 0)

	t.MsgC <- fmt.Sprintf("`%s`", date.Format(t.format))
}

func (t *Utils) TimeConvert(date string) {
	var timestamp int64
	if date != "" || !strings.EqualFold(date, "now") {
		ts, err := time.ParseInLocation(t.format, date, time.Local)
		if err != nil {
			log.Println(err)
			t.ErrC <- "时间格式有误"
			return
		}
		timestamp = ts.Unix()
	} else {
		timestamp = time.Now().Unix()
	}

	t.MsgC <- fmt.Sprintf("`%d`", timestamp)
}


func (t *Utils) JsonFormat(str string) {
	var out bytes.Buffer
	err := json.Indent(&out, []byte(str), "", "    ")
	if err != nil {
		log.Println(err)
		t.ErrC <- "格式化失败"
		return
	}
	t.MsgC <- fmt.Sprintf("`%s`", out.String())
}
