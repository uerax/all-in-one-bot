package crypto

import (
	"time"
)

type TokenBucket struct {
	tokens        chan struct{} // 令牌通道
	tokenRate     int           // 令牌生成速率
	//availableTime time.Time     // 下一个令牌可用时间
}

// 初始化令牌桶
func NewTokenBucket(tokenRate int) *TokenBucket {
	tb := &TokenBucket{
		tokens:        make(chan struct{}, tokenRate),
		tokenRate:     tokenRate,
		//availableTime: time.Now(),
	}

	// 启动令牌生成协程
	go tb.generateTokens()

	return tb
}

// 生成令牌
func (tb *TokenBucket) generateTokens() {
	t := time.NewTicker(time.Second / time.Duration(tb.tokenRate))
	for range t.C {
		tb.tokens <- struct{}{} // 生成一个令牌
	}
}

// 获取令牌
func (tb *TokenBucket) GetToken() {
	// 等待令牌可用
	// for time.Now().Before(tb.availableTime) {
	// 	time.Sleep(time.Until(tb.availableTime))
	// }

	// 获取令牌
	<-tb.tokens

	// 更新下一个令牌可用时间
	// tb.availableTime = tb.availableTime.Add(time.Second / time.Duration(tb.tokenRate))
}