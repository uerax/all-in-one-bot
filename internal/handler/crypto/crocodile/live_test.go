package crocodile

import (
	"fmt"
	"testing"
	"time"

	"github.com/uerax/all-in-one-bot/lite/internal/crypto/provider/geckoterminal"
	"github.com/uerax/all-in-one-bot/lite/internal/pkg/logger"
)

func TestRealFetchB3AndAuki(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping live network test in short mode")
	}

	log := logger.NewLogger()
	provider := geckoterminal.NewProvider("https://api.geckoterminal.com/api/v2", 20, log)

	rc := RuleConfig{
		Lookback:          5,
		YesterdayMultiple: 3.0,
		AverageMultiple:   2.0,
	}

	targets := []Item{
		{Network: "base", Name: "auki"},
		{Network: "base", Name: "b3"},
	}

	for _, item := range targets {
		query := item.QueryString()
		fmt.Printf("\n================================================================================\n")
		fmt.Printf("【正在拉取 DEX 真实数据】标的: [%s] %s | 查询参数: %s\n", item.Network, item.Name, query)
		fmt.Printf("================================================================================\n")

		klines, err := provider.GetDailyKline(query)
		if err != nil {
			t.Fatalf("[%s] GetDailyKline 失败: %v", item.Name, err)
		}

		fmt.Printf("成功拉取到 %d 根日 K 线数据（时间已升序归一）\n\n", len(klines))

		// 打印最近 8 根 K 线，展示完整的上下文
		startIdx := max(0, len(klines)-8)
		fmt.Println("---【最近 K 线明细（含今日未闭合柱子与历史柱子）】---")
		for _, k := range klines[startIdx:] {
			dtUTC := k.Timestamp.UTC().Format("2006-01-02 15:04:05 UTC")
			fmt.Printf("  %s | O=%s H=%s L=%s C=%s | 24h成交量: $%s\n",
				dtUTC,
				formatPrice(k.Open),
				formatPrice(k.High),
				formatPrice(k.Low),
				formatPrice(k.Close),
				fmt.Sprintf("%.2f", k.Volume),
			)
		}
		fmt.Println()

		// 执行 Crocodile 量能评估
		sig, err := evaluate(item, klines, rc)
		if err != nil {
			t.Fatalf("[%s] evaluate 评估失败: %v", item.Name, err)
		}

		fmt.Println("---【Crocodile 量能监控诊断与计算过程】---")
		fmt.Printf("  当前系统时间: %s UTC\n", time.Now().UTC().Format("2006-01-02 15:04:05"))
		fmt.Printf("  评估目标日 (昨天完整日): %s\n", sig.Time.Format("2006-01-02"))
		fmt.Printf("  评估收盘价: %s\n", formatPrice(sig.Close))
		fmt.Printf("  目标日成交量: $%.2f\n", sig.Volume)
		fmt.Printf("  前日成交量:   $%.2f\n", sig.PreviousVolume)
		fmt.Printf("  前 5 日均量:  $%.2f\n", sig.PreviousAverageVol)
		fmt.Printf("  昨日倍数: %.2fx (规则阈值: %.1fx) -> %t\n", sig.YesterdayRatio, sig.YesterdayMultiple, sig.YesterdayRatio >= sig.YesterdayMultiple)
		fmt.Printf("  均值倍数: %.2fx (规则阈值: %.1fx) -> %t\n", sig.AverageRatio, sig.AverageMultiple, sig.AverageRatio >= sig.AverageMultiple)
		fmt.Printf("  🔥 监控最终触发买入信号: %t\n", sig.Triggered)
		fmt.Printf("  Web 端核对链接: %s\n", item.DexLink())
	}
}
