package cron

import (
	"context"
	"fmt"
	"strconv"
	"time"
)

type Task struct {
	task map[string]context.CancelFunc
	C    chan string
	idx  int
}

func NewTask() *Task {
	return &Task{
		task: make(map[string]context.CancelFunc),
		C:    make(chan string, 5),
		idx:  0,
	}
}

func (t *Task) Once(itv string, msg string) {
	i, err := strconv.ParseInt(itv, 10, 64)
	if err != nil {
		return
	}

	t.C <- msg

	time.Sleep(time.Duration(i) * time.Hour)

	t.C <- fmt.Sprintf("%d 小时了, %s", i, msg)

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
	t.C <- fmt.Sprintf("定时提醒编号为: %d 已启动", t.idx)
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
			t.C <- m
		case <-ctx.Done():
			t.C <- fmt.Sprintf("编号: %d 的定时提醒已关闭", idx)
			return
		}
	}
}
