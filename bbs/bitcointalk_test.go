package bbs

import (
	"context"
	"reflect"
	"testing"
)

func TestNewBitcointalk(t *testing.T) {
	got := NewBitcointalk()
	got.FilterFill()
	got.Monitor()
}

func TestBitcointalk_FilterFill(t *testing.T) {
	type fields struct {
		url     string
		old     map[string]struct{}
		ctx     context.Context
		cancel  context.CancelFunc
		C       chan string
		filter  map[string]struct{}
		notifi  bool
		path    string
		running bool
	}
	tests := []struct {
		name   string
		fields fields
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &Bitcointalk{
				url:     tt.fields.url,
				old:     tt.fields.old,
				ctx:     tt.fields.ctx,
				cancel:  tt.fields.cancel,
				C:       tt.fields.C,
				filter:  tt.fields.filter,
				notifi:  tt.fields.notifi,
				path:    tt.fields.path,
				running: tt.fields.running,
			}
			b.FilterFill()
		})
	}
}

func TestBitcointalk_CronDump(t *testing.T) {
	type fields struct {
		url     string
		old     map[string]struct{}
		ctx     context.Context
		cancel  context.CancelFunc
		C       chan string
		filter  map[string]struct{}
		notifi  bool
		path    string
		running bool
	}
	tests := []struct {
		name   string
		fields fields
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &Bitcointalk{
				url:     tt.fields.url,
				old:     tt.fields.old,
				ctx:     tt.fields.ctx,
				cancel:  tt.fields.cancel,
				C:       tt.fields.C,
				filter:  tt.fields.filter,
				notifi:  tt.fields.notifi,
				path:    tt.fields.path,
				running: tt.fields.running,
			}
			b.CronDump()
		})
	}
}

func TestBitcointalk_Dump(t *testing.T) {
	type fields struct {
		url     string
		old     map[string]struct{}
		ctx     context.Context
		cancel  context.CancelFunc
		C       chan string
		filter  map[string]struct{}
		notifi  bool
		path    string
		running bool
	}
	tests := []struct {
		name   string
		fields fields
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &Bitcointalk{
				url:     tt.fields.url,
				old:     tt.fields.old,
				ctx:     tt.fields.ctx,
				cancel:  tt.fields.cancel,
				C:       tt.fields.C,
				filter:  tt.fields.filter,
				notifi:  tt.fields.notifi,
				path:    tt.fields.path,
				running: tt.fields.running,
			}
			b.Dump()
		})
	}
}

func TestBitcointalk_Recover(t *testing.T) {
	type fields struct {
		url     string
		old     map[string]struct{}
		ctx     context.Context
		cancel  context.CancelFunc
		C       chan string
		filter  map[string]struct{}
		notifi  bool
		path    string
		running bool
	}
	tests := []struct {
		name   string
		fields fields
		want   map[string]struct{}
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &Bitcointalk{
				url:     tt.fields.url,
				old:     tt.fields.old,
				ctx:     tt.fields.ctx,
				cancel:  tt.fields.cancel,
				C:       tt.fields.C,
				filter:  tt.fields.filter,
				notifi:  tt.fields.notifi,
				path:    tt.fields.path,
				running: tt.fields.running,
			}
			if got := b.Recover(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Bitcointalk.Recover() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBitcointalk_Start(t *testing.T) {
	type fields struct {
		url     string
		old     map[string]struct{}
		ctx     context.Context
		cancel  context.CancelFunc
		C       chan string
		filter  map[string]struct{}
		notifi  bool
		path    string
		running bool
	}
	tests := []struct {
		name   string
		fields fields
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &Bitcointalk{
				url:     tt.fields.url,
				old:     tt.fields.old,
				ctx:     tt.fields.ctx,
				cancel:  tt.fields.cancel,
				C:       tt.fields.C,
				filter:  tt.fields.filter,
				notifi:  tt.fields.notifi,
				path:    tt.fields.path,
				running: tt.fields.running,
			}
			b.Start()
		})
	}
}

func TestBitcointalk_Monitor(t *testing.T) {
	type fields struct {
		url     string
		old     map[string]struct{}
		ctx     context.Context
		cancel  context.CancelFunc
		C       chan string
		filter  map[string]struct{}
		notifi  bool
		path    string
		running bool
	}
	tests := []struct {
		name   string
		fields fields
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &Bitcointalk{
				url:     tt.fields.url,
				old:     tt.fields.old,
				ctx:     tt.fields.ctx,
				cancel:  tt.fields.cancel,
				C:       tt.fields.C,
				filter:  tt.fields.filter,
				notifi:  tt.fields.notifi,
				path:    tt.fields.path,
				running: tt.fields.running,
			}
			b.Monitor()
		})
	}
}

func TestBitcointalk_Stop(t *testing.T) {
	type fields struct {
		url     string
		old     map[string]struct{}
		ctx     context.Context
		cancel  context.CancelFunc
		C       chan string
		filter  map[string]struct{}
		notifi  bool
		path    string
		running bool
	}
	tests := []struct {
		name   string
		fields fields
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &Bitcointalk{
				url:     tt.fields.url,
				old:     tt.fields.old,
				ctx:     tt.fields.ctx,
				cancel:  tt.fields.cancel,
				C:       tt.fields.C,
				filter:  tt.fields.filter,
				notifi:  tt.fields.notifi,
				path:    tt.fields.path,
				running: tt.fields.running,
			}
			b.Stop()
		})
	}
}
