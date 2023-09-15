package crypto

import (
	"strings"
	"sync/atomic"
	"time"
)

func (t *Track) Kline(token, start, end string) (float64, *HoneypotResp) {
	idx := t.KlineIdx
	atomic.AddInt32(&t.KlineIdx, 1)
	if idx % 2 == 0 {
		return t.KlineAnalyze(token, start, end)
	} else if idx % 2 == 1 {
		return t.PriceHighestAndNow(token, start, "now", true)
	}
	return 0, nil
}



func (t *Track) KlineAnalyze(token, start, end string) (float64, *HoneypotResp) {
	from, err := time.ParseInLocation("2006-01-02_15:04:05", start, time.Local)
	if err != nil {
		t.C <- "时间格式输入错误,请按照以下格式'2006-01-02_15:04:05'"
		return 0, nil
	}
	to := time.Now()
	if !strings.EqualFold(end, "now") {
		to, err = time.ParseInLocation("2006-01-02_15:04:05", end, time.Local)
		if err != nil {
			t.C <- "时间格式输入错误,请按照以下格式'2006-01-02_15:04:05'"
			return 0, nil
		}
	}

	resolution := 1
	duration := to.Sub(from)
	switch {
	case duration > 24*time.Hour:
		if strings.EqualFold(end, "now") {
			to = from.Add(24*time.Hour)
			resolution = 5
		} else {
			resolution = 60
		}
		
	case duration > 5*time.Hour:
		resolution = 5
	}

	dk := t.api.GuruKline(token, from.Unix(), to.Unix(), resolution)

	if dk == nil {
		return 0, nil
	}

	var o, h, readH float64
	// 大于初始价格K线数
	var oGto, cGto, hGto int
	if len(dk.CUsd) > 0 {
		o = dk.CUsd[0]
		if len(dk.OUsd) > 0 && dk.OUsd[0] > o {
			o = dk.OUsd[0]
		}
		if resolution > 5 {
			o = (o + dk.OUsd[0]) / 2.0
		}
	}

	for k := range dk.HUsd {
		if o < dk.HUsd[k] {
			hGto++
		}
		if dk.HUsd[k] > h {
			h = dk.HUsd[k]
		}
	}
	for k := range dk.OUsd {
		if o < dk.OUsd[k] {
			oGto++
		}
		if readH < dk.OUsd[k] {
			readH = dk.OUsd[k]
		}
	}
	for k := range dk.CUsd {
		if o < dk.CUsd[k] {
			cGto++
		}
		if readH < dk.CUsd[k] {
			readH = dk.CUsd[k]
		}
	}

	gto := cGto
	if gto < oGto {
		gto = oGto
	}

	_, readP := 0.0, 0.0
	if o != 0 {
		//profit = (h - o) / o
		readP = (readH - o) / o
	}

	if gto < 2 {
		readP = 0.0
	}
	return readP, nil
}