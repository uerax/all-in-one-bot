package coin

import (
	"fmt"
	"strings"

	"github.com/uerax/all-in-one-bot/lite/internal/crypto/provider"
)

func formatPrice(p float64) string {
	switch {
	case p >= 1.0:
		return fmt.Sprintf("%.2f", p)
	case p >= 1e-4:
		return fmt.Sprintf("%.6f", p)
	case p >= 1e-8:
		return fmt.Sprintf("%.10f", p)
	default:
		return fmt.Sprintf("%.12f", p)
	}
}

func formatUSD(v float64) string {
	switch {
	case v >= 1e9:
		return fmt.Sprintf("$%.2fB", v/1e9)
	case v >= 1e6:
		return fmt.Sprintf("$%.2fM", v/1e6)
	case v >= 1e3:
		return fmt.Sprintf("$%.2fK", v/1e3)
	case v > 0:
		return fmt.Sprintf("$%.2f", v)
	default:
		return "$0.00"
	}
}

func formatPercent(pct float64) string {
	if pct >= 0 {
		return fmt.Sprintf("+%.2f%%", pct)
	}
	return fmt.Sprintf("%.2f%%", pct)
}

func FormatMarketData(data *provider.MarketData) string {
	if data == nil {
		return "未找到币种行情数据"
	}

	var sb strings.Builder

	title := fmt.Sprintf("🪙 *%s*", data.Name)
	if data.Symbol != "" {
		title += fmt.Sprintf(" (%s)", strings.ToUpper(data.Symbol))
	}
	if data.Chain != "" {
		chainInfo := strings.Title(data.Chain)
		if data.DEX != "" {
			chainInfo += " / " + data.DEX
		}
		title += fmt.Sprintf(" [%s]", chainInfo)
	}
	sb.WriteString(title + "\n")

	if data.ContractAddress != "" {
		sb.WriteString(fmt.Sprintf("合约: `%s`\n", data.ContractAddress))
	}
	if data.PoolAddress != "" {
		sb.WriteString(fmt.Sprintf("池子: `%s`\n", data.PoolAddress))
	}

	sb.WriteString(fmt.Sprintf("价格: *$%s*\n", formatPrice(data.PriceUSD)))

	// Multi-timeframe changes if available
	changes := []string{}
	if data.Change5m != 0 {
		changes = append(changes, fmt.Sprintf("5m: *%s*", formatPercent(data.Change5m)))
	}
	if data.Change1h != 0 {
		changes = append(changes, fmt.Sprintf("1h: *%s*", formatPercent(data.Change1h)))
	}
	if data.Change6h != 0 {
		changes = append(changes, fmt.Sprintf("6h: *%s*", formatPercent(data.Change6h)))
	}
	changes = append(changes, fmt.Sprintf("24h: *%s*", formatPercent(data.Change24h)))

	sb.WriteString(strings.Join(changes, " | ") + "\n")

	if data.Volume24hUSD > 0 {
		sb.WriteString(fmt.Sprintf("24h 成交量: *%s*\n", formatUSD(data.Volume24hUSD)))
	}
	if data.ReserveUSD > 0 {
		sb.WriteString(fmt.Sprintf("池子流动性: *%s*\n", formatUSD(data.ReserveUSD)))
	}
	if data.FDV > 0 {
		sb.WriteString(fmt.Sprintf("FDV: *%s*\n", formatUSD(data.FDV)))
	}
	if data.MarketCap > 0 {
		sb.WriteString(fmt.Sprintf("市值: *%s*", formatUSD(data.MarketCap)))
		if data.MarketCapRank > 0 {
			sb.WriteString(fmt.Sprintf(" (排名 #%d)", data.MarketCapRank))
		}
		sb.WriteString("\n")
	}

	sb.WriteString(fmt.Sprintf("数据源: `%s`", data.Source))
	return sb.String()
}

func FormatMultiSearchResult(res *provider.MultiSearchResult) string {
	if res == nil || (len(res.CEXResults) == 0 && len(res.DEXResults) == 0) {
		return "未找到相关币种或 DEX 池子"
	}

	var sb strings.Builder

	if len(res.CEXResults) > 0 {
		sb.WriteString("*CEX 市场结果 (CoinGecko)*\n")
		for _, item := range res.CEXResults {
			rank := "N/A"
			if item.MarketCapRank > 0 {
				rank = fmt.Sprintf("#%d", item.MarketCapRank)
			}
			sb.WriteString(fmt.Sprintf("`%s` *%s* (%s) 市值排名: %s\n",
				item.ID, item.Name, strings.ToUpper(item.Symbol), rank))
		}
		sb.WriteString("\n")
	}

	if len(res.DEXResults) > 0 {
		sb.WriteString("*DEX 链上池子 (GeckoTerminal)*\n")
		for _, item := range res.DEXResults {
			chain := strings.Title(item.Chain)
			sb.WriteString(fmt.Sprintf("[%s] *%s*\n", chain, item.Name))
			sb.WriteString(fmt.Sprintf("池子: `%s` | 价格: *$%s*\n", item.PoolAddress, formatPrice(item.PriceUSD)))
			if item.ReserveUSD > 0 {
				sb.WriteString(fmt.Sprintf("流动性: *%s* | 24h 量: *%s*\n", formatUSD(item.ReserveUSD), formatUSD(item.Volume24hUSD)))
			}
		}
	}

	return strings.TrimSpace(sb.String())
}

func FormatTrending(network string, list []provider.MarketData) string {
	if len(list) == 0 {
		return "暂无 DEX 热门池子数据"
	}

	header := "*DEX 热门池子榜单*"
	if network != "" {
		header = fmt.Sprintf("*DEX 热门池子榜单 (%s)*", strings.Title(network))
	}

	var lines []string
	lines = append(lines, header)

	for i, item := range list {
		if i >= 10 {
			break
		}
		dexInfo := ""
		if item.Chain != "" {
			dexInfo = fmt.Sprintf("[%s]", strings.Title(item.Chain))
		}
		lines = append(lines, fmt.Sprintf("%d. %s *%s* (%s)\n   价格: *$%s* | 24h: *%s* | 流动性: *%s*",
			i+1, dexInfo, item.Name, item.Symbol,
			formatPrice(item.PriceUSD), formatPercent(item.Change24h), formatUSD(item.ReserveUSD)))
	}

	return strings.Join(lines, "\n")
}
