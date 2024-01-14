package bbs

import (
	"context"
	"time"
)

type Bitcointalk struct {
	URL string
	ctx context.Context
	cancel context.CancelFunc
}

func (b *Bitcointalk) Start() {

	ticker := time.NewTicker(60 * time.Second)

	for {
		select {
		case <-ticker.C:
			b.Monitor()
		case <-b.ctx.Done():
			return
		}
	}
}

func (b *Bitcointalk) Monitor() {
	
}

func (b *Bitcointalk) Stop() {
	b.cancel()
}

func (b *Bitcointalk) GetStatus() string {
	return ""
}
