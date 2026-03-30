package polymarket

import "testing"

func TestParseHoldingsArgs(t *testing.T) {
	tests := []struct {
		name       string
		payload    string
		wantAddr   string
		wantLimit  int
		wantErr    bool
	}{
		{name: "address only", payload: "0x1111111111111111111111111111111111111111", wantAddr: "0x1111111111111111111111111111111111111111", wantLimit: 0, wantErr: false},
		{name: "address and limit", payload: "0x1111111111111111111111111111111111111111 15", wantAddr: "0x1111111111111111111111111111111111111111", wantLimit: 15, wantErr: false},
		{name: "invalid address", payload: "abc 10", wantErr: true},
		{name: "invalid limit", payload: "0x1111111111111111111111111111111111111111 x", wantErr: true},
		{name: "negative limit", payload: "0x1111111111111111111111111111111111111111 -1", wantErr: true},
		{name: "empty", payload: "", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			addr, limit, err := parseHoldingsArgs(tt.payload)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("parseHoldingsArgs() error=nil, want error")
				}
				return
			}
			if err != nil {
				t.Fatalf("parseHoldingsArgs() error=%v", err)
			}
			if addr != tt.wantAddr {
				t.Fatalf("addr=%s, want %s", addr, tt.wantAddr)
			}
			if limit != tt.wantLimit {
				t.Fatalf("limit=%d, want %d", limit, tt.wantLimit)
			}
		})
	}
}

func TestNormalizeAddress(t *testing.T) {
	valid := "0xAaAaAaAaAaAaAaAaAaAaAaAaAaAaAaAaAaAaAaAa"
	norm, err := normalizeAddress(valid)
	if err != nil {
		t.Fatalf("normalizeAddress() error=%v", err)
	}
	if norm != "0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa" {
		t.Fatalf("normalized=%s", norm)
	}

	if _, err := normalizeAddress("0x123"); err == nil {
		t.Fatal("normalizeAddress() expected error for short address")
	}
}
