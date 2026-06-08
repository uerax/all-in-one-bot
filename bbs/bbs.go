package bbs

import "github.com/uerax/all-in-one-bot/common"

type Bbs struct {
	Bitcointalk *Bitcointalk
	Nodeseek    *Nodeseek
}

func NewBbs(ch ...chan<- common.AioEvent) *Bbs {
	var eventCh chan<- common.AioEvent
	if len(ch) > 0 {
		eventCh = ch[0]
	}
	return &Bbs{
		Bitcointalk: NewBitcointalk(eventCh),
		Nodeseek:    NewNodeseek(eventCh),
	}
}
