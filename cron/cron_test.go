package cron

import (
	"context"
	"fmt"
	"testing"
)

func TestNewTask(t *testing.T) {
	
	calulate := func (array []int, part int) [][]int {
		ret := make([][]int, 0)
		if len(array) == 0 || part < 2 {
			ret = append(ret, array)
			return ret
		}
		l := len(array) / part
		for i,j := 0, 0; i < part; i,j=i+1,j+l {
			if i + 1 == part {
				ret = append(ret, array[j:])
			} else {
				ret = append(ret, array[j:j+l])
			}
		}

		return ret
	}

	fmt.Println(calulate([]int{}, 10))
	fmt.Println(calulate([]int{1,2,3,4,5,6,7,8,9,10}, 0))
	// fmt.Println(calulate([]int{1,2,3,4,5,6,7,8,9,10}, 2))
	fmt.Println(calulate([]int{1,2,3,4,5,6,7,8,9,10}, 7))
}

func TestTask_AddTask(t *testing.T) {
	type fields struct {
		task map[string]context.CancelFunc
		C    chan string
		idx  int
	}
	type args struct {
		itv string
		msg string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := &Task{
				task: tt.fields.task,
				C:    tt.fields.C,
				idx:  tt.fields.idx,
			}
			tr.AddTask(tt.args.itv, tt.args.msg)
		})
	}
}

func TestTask_Increase(t *testing.T) {
	type fields struct {
		task map[string]context.CancelFunc
		C    chan string
		idx  int
	}
	tests := []struct {
		name   string
		fields fields
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := &Task{
				task: tt.fields.task,
				C:    tt.fields.C,
				idx:  tt.fields.idx,
			}
			tr.Increase()
		})
	}
}

func TestTask_CloseTask(t *testing.T) {
	type fields struct {
		task map[string]context.CancelFunc
		C    chan string
		idx  int
	}
	type args struct {
		idx string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := &Task{
				task: tt.fields.task,
				C:    tt.fields.C,
				idx:  tt.fields.idx,
			}
			tr.CloseTask(tt.args.idx)
		})
	}
}

func TestTask_Do(t *testing.T) {
	type fields struct {
		task map[string]context.CancelFunc
		C    chan string
		idx  int
	}
	type args struct {
		itv int64
		msg string
		ctx context.Context
		idx int
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := &Task{
				task: tt.fields.task,
				C:    tt.fields.C,
				idx:  tt.fields.idx,
			}
			tr.Do(tt.args.itv, tt.args.msg, tt.args.ctx, tt.args.idx)
		})
	}
}
