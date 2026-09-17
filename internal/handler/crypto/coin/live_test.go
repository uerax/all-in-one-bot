package coin

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	_ "github.com/joho/godotenv/autoload"
	"github.com/uerax/all-in-one-bot/lite/internal/config"
	"github.com/uerax/all-in-one-bot/lite/internal/crypto/provider"
	cgprovider "github.com/uerax/all-in-one-bot/lite/internal/crypto/provider/coingecko"
	gtprovider "github.com/uerax/all-in-one-bot/lite/internal/crypto/provider/geckoterminal"
	"github.com/uerax/all-in-one-bot/lite/internal/pkg/logger"
)

func TestRealCoinPriceSources(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping live network test in short mode")
	}

	// 1. 读取 data/coin/list.json
	dataPaths := []string{
		"../../../../data/coin/list.json",
		"data/coin/list.json",
		filepath.Join(os.Getenv("PWD"), "data/coin/list.json"),
	}

	var rawData []byte
	var err error
	for _, p := range dataPaths {
		rawData, err = os.ReadFile(p)
		if err == nil {
			break
		}
	}
	if err != nil {
		t.Fatalf("无法读取 data/coin/list.json: %v", err)
	}

	var list map[string]float64
	if err := json.Unmarshal(rawData, &list); err != nil {
		t.Fatalf("解析 data/coin/list.json 失败: %v", err)
	}

	fmt.Printf("成功从 data/coin/list.json 加载持仓列表，共 %d 个币种:\n", len(list))
	for k, v := range list {
		fmt.Printf("  - %s: %v\n", k, v)
	}
	fmt.Println()

	// 2. 初始化真实 Provider 和 Manager
	cfg := config.LoadConfig("")
	log := logger.NewLogger()
	cgProv := cgprovider.NewProvider(cfg.Coingecko, log)
	gtProv := gtprovider.NewProvider(cfg.GeckoTerminal.BaseURL, cfg.GeckoTerminal.Timeout, log)
	mgr := provider.NewManager(cgProv, cgProv, cgProv, gtProv, gtProv, gtProv, gtProv)

	// 3. 逐个真实请求并分析数据源
	keys := []string{
		"the-innovation-game",
		"wrapped-quil",
		"bitcoin",
		"ethereum",
		"qubic-network",
	}

	for _, key := range keys {
		count := list[key]
		fmt.Printf("================================================================================\n")
		fmt.Printf("【真实请求测试】标的: %s (配置持仓数量: %v)\n", key, count)
		fmt.Printf("================================================================================\n")

		// 3.1 独立测试 CoinGecko (CEX)
		cgData, cgErr := cgProv.GetPrice(key)
		if cgErr != nil {
			fmt.Printf("  [1. CoinGecko (CEX)] ❌ 请求失败: %v\n", cgErr)
		} else {
			fmt.Printf("  [1. CoinGecko (CEX)] ✅ 命中! Name: %s, Symbol: %s, PriceUSD: $%s, Change24h: %.2f%%\n",
				cgData.Name, cgData.Symbol, formatPrice(cgData.PriceUSD), cgData.Change24h)
		}

		// 3.2 独立测试 GeckoTerminal (DEX)
		gtData, gtErr := gtProv.GetPrice(key)
		if gtErr != nil {
			fmt.Printf("  [2. GeckoTerminal (DEX)] ❌ 请求失败: %v\n", gtErr)
		} else {
			fmt.Printf("  [2. GeckoTerminal (DEX)] ✅ 命中! Name: %s, Symbol: %s, PriceUSD: $%s, Chain: %s, DEX: %s\n",
				gtData.Name, gtData.Symbol, formatPrice(gtData.PriceUSD), gtData.Chain, gtData.DEX)
		}

		// 3.3 测试 Manager.GetPrice 统一路由决策
		data, err := mgr.GetPrice(key)
		if err != nil {
			fmt.Printf("  [3. Manager.GetPrice 最终决策] ❌ 全部失败: %v\n", err)
		} else {
			fmt.Printf("  [3. Manager.GetPrice 最终决策] 🎯 最终采用源: [%s] (类型: %s)\n", data.Source, data.SourceType)
			fmt.Printf("      标的详情: %s (%s) | 价格: $%s | 24h涨跌: %.2f%% | 对应持仓估值: $%.2f\n",
				data.Name, data.Symbol, formatPrice(data.PriceUSD), data.Change24h, count*data.PriceUSD)
			if data.Source == "geckoterminal" && cgErr != nil {
				fmt.Printf("      💡 路由说明: CEX (CoinGecko) 未命中，触发智能平滑回退 (Fallback) 到 DEX (GeckoTerminal)\n")
			} else if data.Source == "coingecko" {
				fmt.Printf("      💡 路由说明: 优先命中 CEX (CoinGecko)\n")
			}
		}
		fmt.Println()

		// 规避公共 API 快速请求频率限制
		time.Sleep(1500 * time.Millisecond)
	}
}
