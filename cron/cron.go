package cron

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/uerax/all-in-one-bot/common"
)

type Task struct {
	task map[string]context.CancelFunc
	ch   chan<- common.AioEvent
	idx  int
}

func NewTask(ch ...chan<- common.AioEvent) *Task {
	var eventCh chan<- common.AioEvent
	if len(ch) > 0 {
		eventCh = ch[0]
	}
	return &Task{
		task: make(map[string]context.CancelFunc),
		ch:   eventCh,
		idx:  0,
	}
}

func (t *Task) Once(itv string, msg string) {
	i, err := strconv.ParseInt(itv, 10, 64)
	if err != nil {
		return
	}

	common.Send(t.ch, common.Text(fmt.Sprintf("%s, %s", time.Now().Format("2006年1月2日 15:04"), msg)))

	time.Sleep(time.Duration(i) * time.Hour)

	common.Send(t.ch, common.Text(fmt.Sprintf("%d 小时了, %s", i, msg)))

}

func (t *Task) AddTask(itv string, msg string) {
	i, err := strconv.ParseInt(itv, 10, 64)
	if err != nil {
		return
	}
	ctx, cancel := context.WithCancel(context.Background())
	t.task[strconv.Itoa(t.idx)] = cancel
	go t.Do(i, msg, ctx, t.idx)
	t.Increase()
	common.Send(t.ch, common.Text(fmt.Sprintf("定时提醒编号为: %d 已启动", t.idx)))
}

func (t *Task) Increase() {
	t.idx++
}

func (t *Task) CloseTask(idx string) {
	if v, ok := t.task[idx]; ok {
		v()
		//delete(t.task, idx)
	}
}

func (t *Task) Do(itv int64, msg string, ctx context.Context, idx int) {
	ticker := time.NewTicker(time.Duration(itv) * time.Second)
	defer ticker.Stop()

	m := fmt.Sprintf("编号: %d 的定时提醒, 间隔为 %ds: \n%s", idx, itv, msg)

	for {
		select {
		case <-ticker.C:
			common.Send(t.ch, common.Text(m))
		case <-ctx.Done():
			common.Send(t.ch, common.Text(fmt.Sprintf("编号: %d 的定时提醒已关闭", idx)))
			return
		}
	}
}
