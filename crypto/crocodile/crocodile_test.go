package crocodile

import (
	"errors"
	"image/png"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/uerax/all-in-one-bot/common"
	"github.com/uerax/all-in-one-bot/crypto"
)

type fakeKlineSource struct {
	data map[string][]crypto.CoingeckoKline
	err  map[string]error
}

func (f fakeKlineSource) GetDailyKline(coin string) ([]crypto.CoingeckoKline, error) {
	if err, ok := f.err[coin]; ok {
		return nil, err
	}
	return f.data[coin], nil
}

func TestDecodeListNormalizesItems(t *testing.T) {
	items, err := decodeList([]byte(`[{"name":" BTC ","id":" bitcoin ","rule":" volume_spike "}]`))
	if err != nil {
		t.Fatalf("decodeList returned error: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("len = %d, want 1", len(items))
	}
	if items[0].Name != "BTC" || items[0].ID != "bitcoin" || items[0].Rule != defaultRuleName {
		t.Fatalf("unexpected item: %+v", items[0])
	}
}

func TestInspectRequiresID(t *testing.T) {
	c := NewCrocodileWithSource(fakeKlineSource{})
	_, err := c.inspect(Item{Name: "missing"})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestVolumeSpikeRuleMatches(t *testing.T) {
	rule := volumeSpikeRule{lookback: 5, yesterdayMultiple: 3, averageMultiple: 2}
	sig, err := rule.Evaluate(Item{Name: "BTC", ID: "bitcoin"}, klinesWithVolume(
		day(10, 100),
		day(10, 100),
		day(10, 100),
		day(10, 100),
		day(10, 100),
		day(13, 350),
	))
	if err != nil {
		t.Fatalf("Evaluate returned error: %v", err)
	}
	if sig == nil {
		t.Fatal("expected signal")
	}
	if sig.YesterdayRatio != 3.5 {
		t.Fatalf("yesterday ratio = %f, want 3.5", sig.YesterdayRatio)
	}
	if sig.AverageRatio != 3.5 {
		t.Fatalf("average ratio = %f, want 3.5", sig.AverageRatio)
	}
}

func TestVolumeSpikeRuleSkipsWhenNotMatched(t *testing.T) {
	rule := volumeSpikeRule{lookback: 5, yesterdayMultiple: 3, averageMultiple: 2}
	sig, err := rule.Evaluate(Item{Name: "BTC", ID: "bitcoin"}, klinesWithVolume(
		day(10, 100),
		day(10, 100),
		day(10, 100),
		day(10, 100),
		day(10, 100),
		day(11, 250),
	))
	if err != nil {
		t.Fatalf("Evaluate returned error: %v", err)
	}
	if sig != nil {
		t.Fatalf("expected no signal, got %+v", sig)
	}
}

func TestParseRuleConfig(t *testing.T) {
	cfg, err := parseRuleConfig("7 4.5 3.2")
	if err != nil {
		t.Fatalf("parseRuleConfig returned error: %v", err)
	}
	if cfg.Lookback != 7 || cfg.YesterdayMultiple != 4.5 || cfg.AverageMultiple != 3.2 {
		t.Fatalf("unexpected config: %+v", cfg)
	}
}

func TestRuleTipDescribesCurrentRule(t *testing.T) {
	c := NewCrocodileWithSource(fakeKlineSource{})
	c.SetRuleConfig(RuleConfig{Lookback: 5, YesterdayMultiple: 3, AverageMultiple: 2})

	tip := c.RuleTip()
	for _, want := range []string{
		"当前触发规则是",
		"交易量: 当天 > 3.00x 昨日, 当天 > 2.00x 平均",
		"平均数据取前 5 天",
		"`lookback today avg`",
		"`5 3 2`",
	} {
		if !strings.Contains(tip, want) {
			t.Fatalf("RuleTip missing %q in:\n%s", want, tip)
		}
	}
}

func TestFormatChartRuleLabelsUsesSignalRuleConfig(t *testing.T) {
	today, avg := formatChartRuleLabels(Signal{
		RuleConfig:     RuleConfig{Lookback: 7, YesterdayMultiple: 4.5, AverageMultiple: 3.2},
		YesterdayRatio: 5.5,
		AverageRatio:   4.4,
	})

	if today != "today: 5.5x >= 4.5x" {
		t.Fatalf("unexpected today label: %s", today)
	}
	if avg != "avg(7d): 4.4x >= 3.2x" {
		t.Fatalf("unexpected average label: %s", avg)
	}
}

func TestFormatSignalsUsesSignalRuleConfigLookback(t *testing.T) {
	c := NewCrocodileWithSource(fakeKlineSource{})
	msg := c.formatSignals([]Signal{{
		Item:                  Item{Name: "BTC", ID: "bitcoin"},
		Rule:                  defaultRuleName,
		RuleConfig:            RuleConfig{Lookback: 7, YesterdayMultiple: 4.5, AverageMultiple: 3.2},
		Time:                  time.Date(2024, 6, 6, 0, 0, 0, 0, time.UTC),
		Close:                 13,
		Volume:                350,
		PreviousVolume:        100,
		PreviousAverageVolume: 110,
		YesterdayRatio:        3.5,
		AverageRatio:          3.18,
	}})

	if !strings.Contains(msg, "前7日均量: 110.00") {
		t.Fatalf("formatSignals should use lookback from signal rule config:\n%s", msg)
	}
	if strings.Contains(msg, "前5日均量") {
		t.Fatalf("formatSignals should not use fixed 5-day wording:\n%s", msg)
	}
}

func TestSetRuleConfigUpdatesEvaluation(t *testing.T) {
	c := NewCrocodileWithSource(fakeKlineSource{
		data: map[string][]crypto.CoingeckoKline{
			"bitcoin": klinesWithVolume(
				day(10, 100),
				day(10, 100),
				day(10, 100),
				day(10, 100),
				day(10, 100),
				day(13, 350),
			),
		},
	})
	c.SetRuleConfig(RuleConfig{Lookback: 5, YesterdayMultiple: 4, AverageMultiple: 2})

	sig, err := c.inspect(Item{Name: "BTC", ID: "bitcoin"})
	if err != nil {
		t.Fatalf("inspect returned error: %v", err)
	}
	if sig != nil {
		t.Fatalf("expected no signal with stricter yesterday multiple, got %+v", sig)
	}

	c.SetRuleConfig(RuleConfig{Lookback: 5, YesterdayMultiple: 3, AverageMultiple: 2})
	sig, err = c.inspect(Item{Name: "BTC", ID: "bitcoin"})
	if err != nil {
		t.Fatalf("inspect returned error: %v", err)
	}
	if sig == nil {
		t.Fatal("expected signal after hot update")
	}
	if sig.RuleConfig.YesterdayMultiple != 3 || sig.RuleConfig.AverageMultiple != 2 || sig.RuleConfig.Lookback != 5 {
		t.Fatalf("unexpected signal rule config: %+v", sig.RuleConfig)
	}
}

func TestMarkTriggeredDeduplicatesByDay(t *testing.T) {
	c := NewCrocodileWithSource(fakeKlineSource{})
	if !c.markTriggered("bitcoin:volume_spike", "2024-06-01") {
		t.Fatal("first trigger should pass")
	}
	if c.markTriggered("bitcoin:volume_spike", "2024-06-01") {
		t.Fatal("same day should be deduplicated")
	}
	if !c.markTriggered("bitcoin:volume_spike", "2024-06-02") {
		t.Fatal("new day should pass")
	}
}

func TestCheckContinuesWhenOneItemFails(t *testing.T) {
	ch := make(chan common.AioEvent, 2)
	c := NewCrocodileWithSource(fakeKlineSource{
		data: map[string][]crypto.CoingeckoKline{
			"bitcoin": klinesWithVolume(
				day(10, 100),
				day(10, 100),
				day(10, 100),
				day(10, 100),
				day(10, 100),
				day(13, 350),
			),
		},
		err: map[string]error{
			"bad": errors.New("boom"),
		},
	}, ch)
	c.load = func() ([]Item, error) {
		return []Item{
			{Name: "bad", ID: "bad"},
			{Name: "BTC", ID: "bitcoin"},
		}, nil
	}

	c.check(true, false, false)

	select {
	case event := <-ch:
		if event.Kind != common.EventMarkdown || event.Text != "[BTC触发买入信号](https://www.coingecko.com/en/coins/bitcoin)" || !event.DisableWebPreview {
			t.Fatalf("expected buy signal markdown without preview, got %+v", event)
		}
	default:
		t.Fatal("expected signal message")
	}

	select {
	case event := <-ch:
		if event.Kind != common.EventPhoto || event.Path == "" {
			t.Fatalf("expected chart photo event, got %+v", event)
		}
	default:
		t.Fatal("expected chart photo event")
	}
}

func TestCheckDoesNotNotifyWhenNoSignal(t *testing.T) {
	ch := make(chan common.AioEvent, 1)
	c := NewCrocodileWithSource(fakeKlineSource{
		data: map[string][]crypto.CoingeckoKline{
			"bitcoin": klinesWithVolume(
				day(10, 100),
				day(10, 100),
				day(10, 100),
				day(10, 100),
				day(10, 100),
				day(11, 250),
			),
		},
	}, ch)
	c.load = func() ([]Item, error) {
		return []Item{{Name: "BTC", ID: "bitcoin"}}, nil
	}

	c.check(true, true, false)

	select {
	case event := <-ch:
		t.Fatalf("expected no notification, got %+v", event)
	default:
	}
}

func TestLoadRemoteList(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[{"name":"Wrapped QUIL","id":"wrapped-quil"}]`))
	}))
	defer srv.Close()

	c := NewCrocodileWithSource(fakeKlineSource{})
	c.listURL = srv.URL

	items, err := c.loadList()
	if err != nil {
		t.Fatalf("loadList returned error: %v", err)
	}
	if len(items) != 1 || items[0].Name != "Wrapped QUIL" || items[0].ID != "wrapped-quil" {
		t.Fatalf("unexpected items: %+v", items)
	}
}

func TestLoadListMergesRemoteAndLocalWithLocalOverride(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[{"name":"Remote BTC","id":"bitcoin"},{"name":"ETH","id":"ethereum"}]`))
	}))
	defer srv.Close()

	localPath := filepath.Join(t.TempDir(), "list.json")
	if err := os.WriteFile(localPath, []byte(`[{"name":"Local BTC","id":"bitcoin"},{"name":"Wrapped QUIL","id":"wrapped-quil"}]`), 0644); err != nil {
		t.Fatalf("write local list: %v", err)
	}

	c := NewCrocodileWithSource(fakeKlineSource{})
	c.listURL = srv.URL
	c.localPath = localPath

	items, err := c.loadList()
	if err != nil {
		t.Fatalf("loadList returned error: %v", err)
	}
	if len(items) != 3 {
		t.Fatalf("len = %d, want 3: %+v", len(items), items)
	}

	found := map[string]Item{}
	for _, item := range items {
		found[item.ID] = item
	}
	if found["bitcoin"].Name != "Local BTC" {
		t.Fatalf("local item should override remote duplicate: %+v", found["bitcoin"])
	}
	if found["wrapped-quil"].Name != "Wrapped QUIL" || found["ethereum"].Name != "ETH" {
		t.Fatalf("unexpected merged list: %+v", found)
	}
}

func TestAddItemWritesLocalListAndReplacesDuplicate(t *testing.T) {
	localPath := filepath.Join(t.TempDir(), "list.json")
	c := NewCrocodileWithSource(fakeKlineSource{})
	c.localPath = localPath

	if err := c.AddItem(Item{Name: "Wrapped QUIL", ID: "wrapped-quil"}); err != nil {
		t.Fatalf("AddItem returned error: %v", err)
	}
	if err := c.AddItem(Item{Name: "WQUIL", ID: "wrapped-quil"}); err != nil {
		t.Fatalf("AddItem duplicate returned error: %v", err)
	}

	items, err := c.loadLocalItems()
	if err != nil {
		t.Fatalf("loadLocalItems returned error: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("len = %d, want 1", len(items))
	}
	if items[0].Name != "WQUIL" || items[0].ID != "wrapped-quil" {
		t.Fatalf("unexpected item: %+v", items[0])
	}
}

func TestRenderChartCreatesReadablePNG(t *testing.T) {
	path := filepath.Join(t.TempDir(), "chart.png")
	sig := Signal{
		Item:           Item{Name: "Wrapped QUIL", ID: "wrapped-quil"},
		Time:           time.Date(2024, 6, 6, 0, 0, 0, 0, time.UTC),
		YesterdayRatio: 4.5,
		AverageRatio:   3.2,
	}

	err := renderChart(path, sig, klinesWithVolume(
		day(10, 100),
		day(11, 120),
		day(9, 90),
		day(12, 150),
		day(13, 200),
		day(15, 450),
	))
	if err != nil {
		t.Fatalf("renderChart returned error: %v", err)
	}

	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("open chart: %v", err)
	}
	defer f.Close()

	if _, err := png.Decode(f); err != nil {
		t.Fatalf("decode chart: %v", err)
	}
}

type klineDay struct {
	close  float64
	volume float64
}

func day(close, volume float64) klineDay {
	return klineDay{close: close, volume: volume}
}

func klinesWithVolume(days ...klineDay) []crypto.CoingeckoKline {
	start := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)
	out := make([]crypto.CoingeckoKline, 0, len(days))
	for i, day := range days {
		out = append(out, crypto.CoingeckoKline{
			Time:   start.AddDate(0, 0, i),
			Open:   day.close,
			High:   day.close,
			Low:    day.close,
			Close:  day.close,
			Volume: day.volume,
		})
	}
	return out
}
