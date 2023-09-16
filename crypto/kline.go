package crypto

import (
	"strings"
	"sync/atomic"
	"time"
)

var klineIdx int32

func (t *Track) Kline(addr, token, start, end string) (float64, *HoneypotResp) {
	t.klineLock.RLock()
	if v, ok := t.klineCache[addr]; ok {
		if v1, ok := v[token]; ok {
			t.klineLock.RUnlock()
			return v1, nil
		} else {
			t.klineLock.RUnlock()
		}
	} else {
		t.klineLock.RUnlock()
		t.klineLock.Lock()
		t.klineCache[addr] = make(map[string]float64)
		t.klineLock.Unlock()
	}
	idx := klineIdx
	atomic.AddInt32(&klineIdx, 1)
	if idx%2 == 0 {
		f, hr, err := t.KlineAnalyze(token, start, end)
		if err == nil {
			from, err := time.ParseInLocation("2006-01-02_15:04:05", start, time.Local)
			if err != nil {
				return 0, nil
			}
			to := time.Now()
			if !strings.EqualFold(end, "now") {
				to, err = time.ParseInLocation("2006-01-02_15:04:05", end, time.Local)
				if err != nil {
					return 0, nil
				}
			}
			if to.Sub(from) > time.Hour*24 {
				t.klineLock.Lock()
				t.klineCache[addr][token] = f
				t.klineLock.Unlock()
			}

		}
		return f, hr
	} else if idx%2 == 1 {
		f, hr, err := t.PriceHighestAndNow(token, start, "now", true)
		if err == nil {
			from, err := time.ParseInLocation("2006-01-02_15:04:05", start, time.Local)
			if err != nil {
				return 0, nil
			}
			to := time.Now()
			if !strings.EqualFold(end, "now") {
				to, err = time.ParseInLocation("2006-01-02_15:04:05", end, time.Local)
				if err != nil {
					return 0, nil
				}
			}
			if to.Sub(from) > time.Hour*24 {
				t.klineLock.Lock()
				t.klineCache[addr][token] = f
				t.klineLock.Unlock()
			}
		}
		return f, hr
	}
	return 0, nil
}

func (t *Track) KlineAnalyze(token, start, end string) (float64, *HoneypotResp, error) {
	from, err := time.ParseInLocation("2006-01-02_15:04:05", start, time.Local)
	if err != nil {
		t.C <- "时间格式输入错误,请按照以下格式'2006-01-02_15:04:05'"
		return 0, nil, err
	}
	to := time.Now()
	if !strings.EqualFold(end, "now") {
		to, err = time.ParseInLocation("2006-01-02_15:04:05", end, time.Local)
		if err != nil {
			t.C <- "时间格式输入错误,请按照以下格式'2006-01-02_15:04:05'"
			return 0, nil, err
		}
	}

	resolution := 1
	duration := to.Sub(from)
	switch {
	case duration > 24*time.Hour:
		if strings.EqualFold(end, "now") {
			to = from.Add(24 * time.Hour)
			resolution = 5
		} else {
			resolution = 60
		}

	case duration > 5*time.Hour:
		resolution = 5
	}

	dk, err := t.api.GuruKline(token, from.Unix(), to.Unix(), resolution)
	if err != nil || dk == nil {
		return 0, nil, err
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
	return readP, nil, nil
}
