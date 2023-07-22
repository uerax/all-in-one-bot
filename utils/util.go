package utils

import (
	"bytes"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
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

	ts, err := time.Parse(t.format, date)
	if err != nil {
		log.Println(err)
		t.ErrC <- "时间格式有误"
		return
	}

	t.MsgC <- fmt.Sprintf("`%d`", ts.Unix())
	
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

func (t *Utils) String2Hex(str string) {
	t.MsgC <- hex.EncodeToString([]byte(str))
}

func (t *Utils) Hex2String(str string) {
	s, err := hex.DecodeString(str)
	if err != nil {
		s = []byte("转换失败")
	}
	
	t.MsgC <- string(s)
}

func (t *Utils) DecimalConv(str string, from, to int) {
	num, err := strconv.ParseInt(str, from, 64)
	if err != nil {
		t.MsgC <- "转换失败"
		return
	}
	t.MsgC <- strconv.FormatInt(num, to)
}
