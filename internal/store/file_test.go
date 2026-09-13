package store

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/uerax/all-in-one-bot/lite/internal/config"
)

type dummyLogger struct{}

func (d *dummyLogger) Info(msg string, args ...any)  {}
func (d *dummyLogger) Warn(msg string, args ...any)  {}
func (d *dummyLogger) Error(msg string, args ...any) {}

func TestFileStore_LocalMode(t *testing.T) {
	tempDir := t.TempDir()
	fs := NewFileStore(config.Database{FilePath: tempDir}, &dummyLogger{})

	type Coin struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}

	coins := []Coin{
		{ID: "btc", Name: "Bitcoin"},
		{ID: "eth", Name: "Ethereum"},
	}

	// 1. Save
	if err := fs.Save("crypto", "coins", coins); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// 2. Load
	var loaded []Coin
	if err := fs.Load("crypto", "coins", &loaded); err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if len(loaded) != 2 || loaded[0].ID != "btc" || loaded[1].ID != "eth" {
		t.Fatalf("unexpected loaded data: %+v", loaded)
	}
}

func TestFileStore_RemoteMode_PrioritizeLocalOverRemote(t *testing.T) {
	type Coin struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}

	remoteCoins := []Coin{
		{ID: "remote-1", Name: "Remote 1"},
		{ID: "remote-2", Name: "Remote 2"},
	}

	var remoteHitCount int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&remoteHitCount, 1)
		_ = json.NewEncoder(w).Encode(remoteCoins)
	}))
	defer server.Close()

	localDir := t.TempDir()
	fs := NewFileStore(config.Database{FilePath: server.URL}, &dummyLogger{})
	fs.localDataDir = localDir

	// 1. First Load: No local file exists, must load from remote server
	var firstLoad []Coin
	if err := fs.Load("crypto", "list", &firstLoad); err != nil {
		t.Fatalf("First Load failed: %v", err)
	}
	if len(firstLoad) != 2 || firstLoad[0].ID != "remote-1" {
		t.Fatalf("expected remote data on first load, got %+v", firstLoad)
	}
	if atomic.LoadInt32(&remoteHitCount) != 1 {
		t.Fatalf("expected remote server hit count 1, got %d", remoteHitCount)
	}

	// 2. Save modified data (local write to localDataDir)
	modifiedCoins := []Coin{
		{ID: "local-modified", Name: "Local Modified"},
	}
	if err := fs.Save("crypto", "list", modifiedCoins); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// 3. Second Load: Local file exists now, must prioritize local file and NOT hit remote server
	var secondLoad []Coin
	if err := fs.Load("crypto", "list", &secondLoad); err != nil {
		t.Fatalf("Second Load failed: %v", err)
	}
	if len(secondLoad) != 1 || secondLoad[0].ID != "local-modified" {
		t.Fatalf("expected locally modified data on second load, got %+v", secondLoad)
	}
	// Server hit count should still be 1 (remote URL was bypassed)
	if atomic.LoadInt32(&remoteHitCount) != 1 {
		t.Fatalf("expected remote server hit count still 1, got %d", remoteHitCount)
	}
}
