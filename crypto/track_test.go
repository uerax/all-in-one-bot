package crypto

import (
	"context"
	"reflect"
	"sync"
	"testing"

	"github.com/uerax/goconf"
)

func TestNewTrack(t *testing.T) {
	goconf.LoadConfig("../all-in-one-bot.yml")
	t2 := NewTrack()
	t2.getEthByHtml("0x2826fc2dca18c56d92380381b420bbdaa2fe6d4712f1d63573a839d47e016320")
}

func TestTrack_recover(t *testing.T) {
	type fields struct {
		C            chan string
		Newest       map[string]*newest
		apiKey       string
		Task         map[string]context.CancelFunc
		api          *Crypto
		dumpPath     string
		Keys         *PollingKey
		trackingLock sync.RWMutex
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
		Newest       map[string]*newest
		apiKey       string
		Task         map[string]context.CancelFunc
		api          *Crypto
		dumpPath     string
		Keys         *PollingKey
		trackingLock sync.RWMutex
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
		Newest       map[string]*newest
		apiKey       string
		Task         map[string]context.CancelFunc
		api          *Crypto
		dumpPath     string
		Keys         *PollingKey
		trackingLock sync.RWMutex
	}
	type args struct {
		addr   string
		remark string
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
			tr.CronTracking(tt.args.addr, tt.args.remark)
		})
	}
}

func TestTrack_StopTracking(t *testing.T) {
	type fields struct {
		C            chan string
		Newest       map[string]*newest
		apiKey       string
		Task         map[string]context.CancelFunc
		api          *Crypto
		dumpPath     string
		Keys         *PollingKey
		trackingLock sync.RWMutex
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
		Newest       map[string]*newest
		apiKey       string
		Task         map[string]context.CancelFunc
		api          *Crypto
		dumpPath     string
		Keys         *PollingKey
		trackingLock sync.RWMutex
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
		Newest       map[string]*newest
		apiKey       string
		Task         map[string]context.CancelFunc
		api          *Crypto
		dumpPath     string
		Keys         *PollingKey
		trackingLock sync.RWMutex
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
		Newest       map[string]*newest
		apiKey       string
		Task         map[string]context.CancelFunc
		api          *Crypto
		dumpPath     string
		Keys         *PollingKey
		trackingLock sync.RWMutex
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
		Newest       map[string]*newest
		apiKey       string
		Task         map[string]context.CancelFunc
		api          *Crypto
		dumpPath     string
		Keys         *PollingKey
		trackingLock sync.RWMutex
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
		Newest       map[string]*newest
		apiKey       string
		Task         map[string]context.CancelFunc
		api          *Crypto
		dumpPath     string
		Keys         *PollingKey
		trackingLock sync.RWMutex
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
		Newest       map[string]*newest
		apiKey       string
		Task         map[string]context.CancelFunc
		api          *Crypto
		dumpPath     string
		Keys         *PollingKey
		trackingLock sync.RWMutex
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
		Newest       map[string]*newest
		apiKey       string
		Task         map[string]context.CancelFunc
		api          *Crypto
		dumpPath     string
		Keys         *PollingKey
		trackingLock sync.RWMutex
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
		Newest       map[string]*newest
		apiKey       string
		Task         map[string]context.CancelFunc
		api          *Crypto
		dumpPath     string
		Keys         *PollingKey
		trackingLock sync.RWMutex
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
		want map[string]*newest
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
		Newest       map[string]*newest
		apiKey       string
		Task         map[string]context.CancelFunc
		api          *Crypto
		dumpPath     string
		Keys         *PollingKey
		trackingLock sync.RWMutex
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
		Newest       map[string]*newest
		apiKey       string
		Task         map[string]context.CancelFunc
		api          *Crypto
		dumpPath     string
		Keys         *PollingKey
		trackingLock sync.RWMutex
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
		Newest       map[string]*newest
		apiKey       string
		Task         map[string]context.CancelFunc
		api          *Crypto
		dumpPath     string
		Keys         *PollingKey
		trackingLock sync.RWMutex
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
		Newest       map[string]*newest
		apiKey       string
		Task         map[string]context.CancelFunc
		api          *Crypto
		dumpPath     string
		Keys         *PollingKey
		trackingLock sync.RWMutex
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

func Test_zeroCal(t *testing.T) {
	type args struct {
		str string
	}
	tests := []struct {
		name     string
		args     args
		wantZero int
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotZero := zeroCal(tt.args.str); gotZero != tt.wantZero {
				t.Errorf("zeroCal() = %v, want %v", gotZero, tt.wantZero)
			}
		})
	}
}

func TestTrack_WalletLastTransaction(t *testing.T) {
	type fields struct {
		C            chan string
		Newest       map[string]*newest
		apiKey       string
		Task         map[string]context.CancelFunc
		api          *Crypto
		dumpPath     string
		Keys         *PollingKey
		trackingLock sync.RWMutex
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
			tr.WalletLastTransaction()
		})
	}
}

func TestTrack_BotAddrFinder(t *testing.T) {
	type fields struct {
		C            chan string
		Newest       map[string]*newest
		apiKey       string
		Task         map[string]context.CancelFunc
		api          *Crypto
		dumpPath     string
		Keys         *PollingKey
		trackingLock sync.RWMutex
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
			tr.BotAddrFinder(tt.args.token, tt.args.offset, tt.args.page)
		})
	}
}

func TestTrack_getSourceCode(t *testing.T) {
	type fields struct {
		C            chan string
		Newest       map[string]*newest
		apiKey       string
		Task         map[string]context.CancelFunc
		api          *Crypto
		dumpPath     string
		Keys         *PollingKey
		trackingLock sync.RWMutex
	}
	type args struct {
		addr string
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
			if got := tr.getSourceCode(tt.args.addr); got != tt.want {
				t.Errorf("Track.getSourceCode() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getLinks(t *testing.T) {
	type args struct {
		code string
	}
	tests := []struct {
		name string
		args args
		want map[string]string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getLinks(tt.args.code); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getLinks() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTrack_WalletTxInfo(t *testing.T) {
	type fields struct {
		C            chan string
		Newest       map[string]*newest
		apiKey       string
		Task         map[string]context.CancelFunc
		api          *Crypto
		dumpPath     string
		Keys         *PollingKey
		trackingLock sync.RWMutex
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
			tr.WalletTxInfo(tt.args.addr)
		})
	}
}

func TestTrack_GetTax(t *testing.T) {
	type fields struct {
		C            chan string
		Newest       map[string]*newest
		apiKey       string
		Task         map[string]context.CancelFunc
		api          *Crypto
		dumpPath     string
		Keys         *PollingKey
		trackingLock sync.RWMutex
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
			tr.GetTax(tt.args.addr)
		})
	}
}

func TestTrack_CronTaxTracking(t *testing.T) {
	type fields struct {
		C            chan string
		Newest       map[string]*newest
		apiKey       string
		Task         map[string]context.CancelFunc
		api          *Crypto
		dumpPath     string
		Keys         *PollingKey
		trackingLock sync.RWMutex
	}
	type args struct {
		addr string
		buy  string
		sell string
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
			tr.CronTaxTracking(tt.args.addr, tt.args.buy, tt.args.sell)
		})
	}
}

func TestTrack_TaxTracking(t *testing.T) {
	type fields struct {
		C            chan string
		Newest       map[string]*newest
		apiKey       string
		Task         map[string]context.CancelFunc
		api          *Crypto
		dumpPath     string
		Keys         *PollingKey
		trackingLock sync.RWMutex
	}
	type args struct {
		addr string
		buy  int
		sell int
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
			tr.TaxTracking(tt.args.addr, tt.args.buy, tt.args.sell, tt.args.ctx)
		})
	}
}

func TestTrack_getEthByHtml(t *testing.T) {
	type fields struct {
		C            chan string
		Newest       map[string]*newest
		apiKey       string
		Task         map[string]context.CancelFunc
		api          *Crypto
		dumpPath     string
		Keys         *PollingKey
		trackingLock sync.RWMutex
	}
	type args struct {
		hash string
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
			tr.getEthByHtml(tt.args.hash)
		})
	}
}
