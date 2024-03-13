package utils

import (
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type Qubic struct {  
	AverageScore     float64          `json:"averageScore"`    
	EstimatedIts     int64            `json:"estimatedIts"`    
	SolutionsPerHour int64            `json:"solutionsPerHour"`
}

var defaultToken = "eyJhbGciOiJIUzUxMiIsInR5cCI6IkpXVCJ9.eyJJZCI6ImM4NjVjNmU1LTBiOTQtNDdjNC04NzBkLThmNTRkOTQ5NzgzMiIsInN1YiI6ImRkYmVoZWFkQG91dGxvb2suY29tIiwianRpIjoiYjMwNGIxM2ItZmNjNi00MDhhLTk1MmMtODgwODMzMDgyMDk5IiwiUHVibGljIjoiIiwibmJmIjoxNzEwMzUwMDc1LCJleHAiOjE3MTA0MzY0NzUsImlhdCI6MTcxMDM1MDA3NSwiaXNzIjoiaHR0cHM6Ly9xdWJpYy5saS8iLCJhdWQiOiJodHRwczovL3F1YmljLmxpLyJ9.rT7y3FASfSSNIeBOKVVaxzWEixMvL7XkmfSzHNpmritvNc_Gpa3K9UKz0q-8mkLOo8et369daxe3HLVyelGaZw"

func (t *Utils) QubicProfit(token string) {
	qb, err := QubicInfo(token)
	if err != nil {
		t.ErrC <- err.Error()
		return
	}
	ep1, ep2 := 1035502957.0, 281213017.0

	now := time.Now()

	// totalScore := int(qb.AverageScore) * 676
	dayOfWeek := int(now.Weekday())
	earningPerHour := 0.0
	totalHours := 7 * 24
	if dayOfWeek < 3 {
		// 星期三晚上20点刷新，所以加4
		totalHours = 4 + (24 * (4 + dayOfWeek - 1)) + now.Hour()
	} else if dayOfWeek > 3 {
		totalHours = 4 + (24 * (dayOfWeek - 3 - 1)) + now.Hour()
	} else {
		if now.Hour() > 20 {
			totalHours = now.Hour() - 20
		} else {
			totalHours = 6 * 24 + now.Hour()
		}
	}
	earningPerHour = qb.AverageScore / float64(totalHours)
	// hoursUntilSunday := (7 * 24) - (dayOfWeek * 24 + now.Hour())
	totalEarning := float64(earningPerHour * (7 * 24))
	earn1, earn2 := ep1 / (totalEarning * 1.06), ep2 / (totalEarning * 1.06)

	sol := 1000.0 * float64(qb.SolutionsPerHour) / float64(qb.EstimatedIts)

	msg := fmt.Sprintf("当前全网算力: *%d*\n当前出块: *%d / h*\n当前平均分: *%.f*\n\n本周预计平均分: *%.f*\n\n1000算力平均1小时出块: *%.3f*\n1000算力平均24小时出块: *%.3f*\n1000算力平均7天出块: *%.3f*\n\n单个块收益预计: *%.f*\nEpoch1单块预计: *%.f*\nEpoch2单块预计: *%.f*", qb.EstimatedIts, qb.SolutionsPerHour, qb.AverageScore, totalEarning, sol, sol*24, sol*24*7, earn1 + earn2, earn1, earn2)

	t.MsgC <- msg
	
}

func QubicInfo(token string) (*Qubic, error) {
	
	url := "https://api.qubic.li/Score/Get"

	req, _ := http.NewRequest("GET", url, nil)

	if len(token) == 0 {
		token = defaultToken
	} else {
		defaultToken = token
	}

	req.Header.Add("Accept", "application/json")
	req.Header.Add("Authorization", "Bearer " + token)
	req.Header.Add("Sec-Fetch-Site", "same-site")
	req.Header.Add("Accept-Language", "zh-CN,zh-Hans;q=0.9")
	req.Header.Add("Accept-Encoding", "gzip, deflate, br")
	req.Header.Add("Sec-Fetch-Mode", "cors")
	req.Header.Add("Host", "api.qubic.li")
	req.Header.Add("Origin", "https://app.qubic.li")
	req.Header.Add("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.3 Safari/605.1.15")
	req.Header.Add("Connection", "keep-alive")
	req.Header.Add("Referer", "https://app.qubic.li/")
	req.Header.Add("Sec-Fetch-Dest", "empty")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	reader := res.Body
	if res.Header.Get("Content-Encoding") == "gzip" {
		reader, err = gzip.NewReader(res.Body)
        if err != nil {
            return nil, err
        }
        defer reader.Close()
	}
	body, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}
	qb := Qubic{}
	err = json.Unmarshal(body, &qb)

	return &qb, err
}