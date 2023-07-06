package crypto

import (
	"context"
	"fmt"
	"reflect"
	"sync"
	"testing"
	"time"
)

func TestNewTrack(t *testing.T) {
	fmt.Println(time.Now().After(time.Now().Add(4*time.Second)))
}

func TestTrack_recover(t *testing.T) {
	type fields struct {
		C            chan string
		Newest       map[string]string
		apiKey       string
		Task         map[string]context.CancelFunc
		api          *Crypto
		dumpPath     string
		Keys         *PollingKey
		trackingLock sync.Mutex
	}
	tests := []struct {
		name   string
		fields fields
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := &Track{
				C:            tt.fields.C,
				Newest:       tt.fields.Newest,
				apiKey:       tt.fields.apiKey,
				Task:         tt.fields.Task,
				api:          tt.fields.api,
				dumpPath:     tt.fields.dumpPath,
				Keys:         tt.fields.Keys,
				trackingLock: tt.fields.trackingLock,
			}
			tr.recover()
		})
	}
}

func TestTrack_clearInactiveAddr(t *testing.T) {
	type fields struct {
		C            chan string
		Newest       map[string]string
		apiKey       string
		Task         map[string]context.CancelFunc
		api          *Crypto
		dumpPath     string
		Keys         *PollingKey
		trackingLock sync.Mutex
	}
	tests := []struct {
		name   string
		fields fields
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := &Track{
				C:            tt.fields.C,
				Newest:       tt.fields.Newest,
				apiKey:       tt.fields.apiKey,
				Task:         tt.fields.Task,
				api:          tt.fields.api,
				dumpPath:     tt.fields.dumpPath,
				Keys:         tt.fields.Keys,
				trackingLock: tt.fields.trackingLock,
			}
			tr.clearInactiveAddr()
		})
	}
}

func TestTrack_CronTracking(t *testing.T) {
	type fields struct {
		C            chan string
		Newest       map[string]string
		apiKey       string
		Task         map[string]context.CancelFunc
		api          *Crypto
		dumpPath     string
		Keys         *PollingKey
		trackingLock sync.Mutex
	}
	type args struct {
		addr string
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
			tr := &Track{
				C:            tt.fields.C,
				Newest:       tt.fields.Newest,
				apiKey:       tt.fields.apiKey,
				Task:         tt.fields.Task,
				api:          tt.fields.api,
				dumpPath:     tt.fields.dumpPath,
				Keys:         tt.fields.Keys,
				trackingLock: tt.fields.trackingLock,
			}
			tr.CronTracking(tt.args.addr)
		})
	}
}

func TestTrack_StopTracking(t *testing.T) {
	type fields struct {
		C            chan string
		Newest       map[string]string
		apiKey       string
		Task         map[string]context.CancelFunc
		api          *Crypto
		dumpPath     string
		Keys         *PollingKey
		trackingLock sync.Mutex
	}
	type args struct {
		addr string
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
			tr := &Track{
				C:            tt.fields.C,
				Newest:       tt.fields.Newest,
				apiKey:       tt.fields.apiKey,
				Task:         tt.fields.Task,
				api:          tt.fields.api,
				dumpPath:     tt.fields.dumpPath,
				Keys:         tt.fields.Keys,
				trackingLock: tt.fields.trackingLock,
			}
			tr.StopTracking(tt.args.addr)
		})
	}
}

func TestTrack_Tracking(t *testing.T) {
	type fields struct {
		C            chan string
		Newest       map[string]string
		apiKey       string
		Task         map[string]context.CancelFunc
		api          *Crypto
		dumpPath     string
		Keys         *PollingKey
		trackingLock sync.Mutex
	}
	type args struct {
		addr string
		ctx  context.Context
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
			tr := &Track{
				C:            tt.fields.C,
				Newest:       tt.fields.Newest,
				apiKey:       tt.fields.apiKey,
				Task:         tt.fields.Task,
				api:          tt.fields.api,
				dumpPath:     tt.fields.dumpPath,
				Keys:         tt.fields.Keys,
				trackingLock: tt.fields.trackingLock,
			}
			tr.Tracking(tt.args.addr, tt.args.ctx)
		})
	}
}

func TestTrack_TrackingList(t *testing.T) {
	type fields struct {
		C            chan string
		Newest       map[string]string
		apiKey       string
		Task         map[string]context.CancelFunc
		api          *Crypto
		dumpPath     string
		Keys         *PollingKey
		trackingLock sync.Mutex
	}
	type args struct {
		tip bool
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := &Track{
				C:            tt.fields.C,
				Newest:       tt.fields.Newest,
				apiKey:       tt.fields.apiKey,
				Task:         tt.fields.Task,
				api:          tt.fields.api,
				dumpPath:     tt.fields.dumpPath,
				Keys:         tt.fields.Keys,
				trackingLock: tt.fields.trackingLock,
			}
			if got := tr.TrackingList(tt.args.tip); got != tt.want {
				t.Errorf("Track.TrackingList() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTrack_WalletTracking(t *testing.T) {
	type fields struct {
		C            chan string
		Newest       map[string]string
		apiKey       string
		Task         map[string]context.CancelFunc
		api          *Crypto
		dumpPath     string
		Keys         *PollingKey
		trackingLock sync.Mutex
	}
	type args struct {
		addr string
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
			tr := &Track{
				C:            tt.fields.C,
				Newest:       tt.fields.Newest,
				apiKey:       tt.fields.apiKey,
				Task:         tt.fields.Task,
				api:          tt.fields.api,
				dumpPath:     tt.fields.dumpPath,
				Keys:         tt.fields.Keys,
				trackingLock: tt.fields.trackingLock,
			}
			tr.WalletTracking(tt.args.addr)
		})
	}
}

func TestTrack_AnalyzeAddrTokenProfit(t *testing.T) {
	type fields struct {
		C            chan string
		Newest       map[string]string
		apiKey       string
		Task         map[string]context.CancelFunc
		api          *Crypto
		dumpPath     string
		Keys         *PollingKey
		trackingLock sync.Mutex
	}
	type args struct {
		addr  string
		token string
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
			tr := &Track{
				C:            tt.fields.C,
				Newest:       tt.fields.Newest,
				apiKey:       tt.fields.apiKey,
				Task:         tt.fields.Task,
				api:          tt.fields.api,
				dumpPath:     tt.fields.dumpPath,
				Keys:         tt.fields.Keys,
				trackingLock: tt.fields.trackingLock,
			}
			tr.AnalyzeAddrTokenProfit(tt.args.addr, tt.args.token)
		})
	}
}

func TestTrack_getBuyEthByHash(t *testing.T) {
	type fields struct {
		C            chan string
		Newest       map[string]string
		apiKey       string
		Task         map[string]context.CancelFunc
		api          *Crypto
		dumpPath     string
		Keys         *PollingKey
		trackingLock sync.Mutex
	}
	type args struct {
		hash string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   float64
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := &Track{
				C:            tt.fields.C,
				Newest:       tt.fields.Newest,
				apiKey:       tt.fields.apiKey,
				Task:         tt.fields.Task,
				api:          tt.fields.api,
				dumpPath:     tt.fields.dumpPath,
				Keys:         tt.fields.Keys,
				trackingLock: tt.fields.trackingLock,
			}
			if got := tr.getBuyEthByHash(tt.args.hash); got != tt.want {
				t.Errorf("Track.getBuyEthByHash() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTrack_getSellEthByHash(t *testing.T) {
	type fields struct {
		C            chan string
		Newest       map[string]string
		apiKey       string
		Task         map[string]context.CancelFunc
		api          *Crypto
		dumpPath     string
		Keys         *PollingKey
		trackingLock sync.Mutex
	}
	type args struct {
		hash string
		addr string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   float64
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := &Track{
				C:            tt.fields.C,
				Newest:       tt.fields.Newest,
				apiKey:       tt.fields.apiKey,
				Task:         tt.fields.Task,
				api:          tt.fields.api,
				dumpPath:     tt.fields.dumpPath,
				Keys:         tt.fields.Keys,
				trackingLock: tt.fields.trackingLock,
			}
			if got := tr.getSellEthByHash(tt.args.hash, tt.args.addr); got != tt.want {
				t.Errorf("Track.getSellEthByHash() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTrack_WalletTxAnalyze(t *testing.T) {
	type fields struct {
		C            chan string
		Newest       map[string]string
		apiKey       string
		Task         map[string]context.CancelFunc
		api          *Crypto
		dumpPath     string
		Keys         *PollingKey
		trackingLock sync.Mutex
	}
	type args struct {
		addr   string
		offset string
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
			tr := &Track{
				C:            tt.fields.C,
				Newest:       tt.fields.Newest,
				apiKey:       tt.fields.apiKey,
				Task:         tt.fields.Task,
				api:          tt.fields.api,
				dumpPath:     tt.fields.dumpPath,
				Keys:         tt.fields.Keys,
				trackingLock: tt.fields.trackingLock,
			}
			tr.WalletTxAnalyze(tt.args.addr, tt.args.offset)
		})
	}
}

func TestTrack_DumpTrackingList(t *testing.T) {
	type fields struct {
		C            chan string
		Newest       map[string]string
		apiKey       string
		Task         map[string]context.CancelFunc
		api          *Crypto
		dumpPath     string
		Keys         *PollingKey
		trackingLock sync.Mutex
	}
	type args struct {
		tip bool
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
			tr := &Track{
				C:            tt.fields.C,
				Newest:       tt.fields.Newest,
				apiKey:       tt.fields.apiKey,
				Task:         tt.fields.Task,
				api:          tt.fields.api,
				dumpPath:     tt.fields.dumpPath,
				Keys:         tt.fields.Keys,
				trackingLock: tt.fields.trackingLock,
			}
			tr.DumpTrackingList(tt.args.tip)
		})
	}
}

func Test_recoverTrackingList(t *testing.T) {
	tests := []struct {
		name string
		want map[string]string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := recoverTrackingList(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("recoverTrackingList() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTrack_DumpCron(t *testing.T) {
	type fields struct {
		C            chan string
		Newest       map[string]string
		apiKey       string
		Task         map[string]context.CancelFunc
		api          *Crypto
		dumpPath     string
		Keys         *PollingKey
		trackingLock sync.Mutex
	}
	tests := []struct {
		name   string
		fields fields
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := &Track{
				C:            tt.fields.C,
				Newest:       tt.fields.Newest,
				apiKey:       tt.fields.apiKey,
				Task:         tt.fields.Task,
				api:          tt.fields.api,
				dumpPath:     tt.fields.dumpPath,
				Keys:         tt.fields.Keys,
				trackingLock: tt.fields.trackingLock,
			}
			tr.DumpCron()
		})
	}
}

func TestTrack_SmartAddrFinder(t *testing.T) {
	type fields struct {
		C            chan string
		Newest       map[string]string
		apiKey       string
		Task         map[string]context.CancelFunc
		api          *Crypto
		dumpPath     string
		Keys         *PollingKey
		trackingLock sync.Mutex
	}
	type args struct {
		token  string
		offset string
		page   string
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
			tr := &Track{
				C:            tt.fields.C,
				Newest:       tt.fields.Newest,
				apiKey:       tt.fields.apiKey,
				Task:         tt.fields.Task,
				api:          tt.fields.api,
				dumpPath:     tt.fields.dumpPath,
				Keys:         tt.fields.Keys,
				trackingLock: tt.fields.trackingLock,
			}
			tr.SmartAddrFinder(tt.args.token, tt.args.offset, tt.args.page)
		})
	}
}

func TestTrack_TransferList(t *testing.T) {
	type fields struct {
		C            chan string
		Newest       map[string]string
		apiKey       string
		Task         map[string]context.CancelFunc
		api          *Crypto
		dumpPath     string
		Keys         *PollingKey
		trackingLock sync.Mutex
	}
	type args struct {
		addr  string
		token string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   []TokenTx
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := &Track{
				C:            tt.fields.C,
				Newest:       tt.fields.Newest,
				apiKey:       tt.fields.apiKey,
				Task:         tt.fields.Task,
				api:          tt.fields.api,
				dumpPath:     tt.fields.dumpPath,
				Keys:         tt.fields.Keys,
				trackingLock: tt.fields.trackingLock,
			}
			if got := tr.TransferList(tt.args.addr, tt.args.token); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Track.TransferList() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTrack_WalletTrackingV2(t *testing.T) {
	type fields struct {
		C            chan string
		Newest       map[string]string
		apiKey       string
		Task         map[string]context.CancelFunc
		api          *Crypto
		dumpPath     string
		Keys         *PollingKey
		trackingLock sync.Mutex
	}
	type args struct {
		addr string
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
			tr := &Track{
				C:            tt.fields.C,
				Newest:       tt.fields.Newest,
				apiKey:       tt.fields.apiKey,
				Task:         tt.fields.Task,
				api:          tt.fields.api,
				dumpPath:     tt.fields.dumpPath,
				Keys:         tt.fields.Keys,
				trackingLock: tt.fields.trackingLock,
			}
			tr.WalletTrackingV2(tt.args.addr)
		})
	}
}

func Test_isNull(t *testing.T) {
	type args struct {
		addr string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isNull(tt.args.addr); got != tt.want {
				t.Errorf("isNull() = %v, want %v", got, tt.want)
			}
		})
	}
}
