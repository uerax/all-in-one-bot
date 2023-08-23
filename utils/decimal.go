package utils

import (
	"encoding/hex"
	"fmt"
	"strconv"
)

func (t *Utils) String2Hex(str string) {
	t.MsgC <- hex.EncodeToString([]byte(str))
}

func (t *Utils) Hex2String(str string) {
	s, err := hex.DecodeString(str)
	if err != nil {
		t.MsgC <- "转换失败"
		return
	}
	
	t.MsgC <- string(s)
}

func (t *Utils) StrDecimalConv(str string, to int) {
	num, err := strconv.ParseInt(hex.EncodeToString([]byte(str)), 16, 64)
	if err != nil {
		t.MsgC <- "转换失败"
		return
	}
	t.MsgC <- strconv.FormatInt(num, to)
}

func (t *Utils) DecimalStrConv(str string, from int) {
	num, err := strconv.ParseInt(str, from, 64)
	if err != nil {
		t.MsgC <- "转换失败"
		return
	}
	s, err := hex.DecodeString(fmt.Sprintf("%d", num))
	if err != nil {
		t.MsgC <- "转换失败"
		return
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