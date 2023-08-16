package crypto

import (
	"testing"
	"time"
)

func TestNewPollingKey(t *testing.T) {
	ll := NewPollingKeyV2()
	ll.Add("1")
	go func ()  {
		for i := 0; i < 3; i++ {
			ll.GetKey()
		}
	}()
	
	go func ()  {
		for i := 0; i < 3; i++ {
			ll.GetKey()
		}
	}()
	
	go func ()  {
		for i := 0; i < 3; i++ {
			ll.GetKey()
		}
	}()
	time.Sleep(3*time.Second)
}

func TestPollingKey_IsNull(t *testing.T) {
	type fields struct {
		Keys []string
		idx  int32
	}
	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := &PollingKey{
				Keys: tt.fields.Keys,
				idx:  tt.fields.idx,
			}
			if got := tr.IsNull(); got != tt.want {
				t.Errorf("PollingKey.IsNull() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPollingKey_Len(t *testing.T) {
	type fields struct {
		Keys []string
		idx  int32
	}
	tests := []struct {
		name   string
		fields fields
		want   int
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := &PollingKey{
				Keys: tt.fields.Keys,
				idx:  tt.fields.idx,
			}
			if got := tr.Len(); got != tt.want {
				t.Errorf("PollingKey.Len() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPollingKey_AddKeys(t *testing.T) {
	type fields struct {
		Keys []string
		idx  int32
	}
	type args struct {
		keys []string
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
			tr := &PollingKey{
				Keys: tt.fields.Keys,
				idx:  tt.fields.idx,
			}
			tr.AddKeys(tt.args.keys...)
		})
	}
}

func TestPollingKey_GetKey(t *testing.T) {
	type fields struct {
		Keys []string
		idx  int32
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := &PollingKey{
				Keys: tt.fields.Keys,
				idx:  tt.fields.idx,
			}
			if got := tr.GetKey(); got != tt.want {
				t.Errorf("PollingKey.GetKey() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPollingKey_CallsPerM(t *testing.T) {
	type fields struct {
		Keys []string
		idx  int32
	}
	tests := []struct {
		name   string
		fields fields
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := &PollingKey{
				Keys: tt.fields.Keys,
				idx:  tt.fields.idx,
			}
			tr.CallsPerM()
		})
	}
}

func TestLinkedList_Add(t *testing.T) {
	type fields struct {
		Key  string
		Next *LinkedList
	}
	type args struct {
		key string
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
			tr := &LinkedList{
				Key:  tt.fields.Key,
				Next: tt.fields.Next,
			}
			tr.Add(tt.args.key)
		})
	}
}
