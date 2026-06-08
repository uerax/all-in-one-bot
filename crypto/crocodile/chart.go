package crocodile

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"math"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"

	"github.com/uerax/all-in-one-bot/common"
	"github.com/uerax/all-in-one-bot/crypto"
)

const chartKlineLimit = 30

var (
	chartBackground = color.RGBA{250, 250, 250, 255}
	chartPanel      = color.RGBA{255, 255, 255, 255}
	chartGrid       = color.RGBA{224, 228, 232, 255}
	chartAxis       = color.RGBA{80, 84, 88, 255}
	chartText       = color.RGBA{30, 34, 38, 255}
	chartSubText    = color.RGBA{88, 96, 104, 255}
	chartUp         = color.RGBA{22, 150, 86, 255}
	chartDown       = color.RGBA{214, 69, 65, 255}
	chartVolume     = color.RGBA{145, 155, 165, 255}
	chartHighlight  = color.RGBA{55, 105, 220, 255}
)

func (t *Crocodile) sendCharts(results []inspectionResult) {
	for _, result := range results {
		path, err := t.renderSignalChart(result.Signal, result.Klines)
		if err != nil {
			logChartError(result.Signal.Item, err)
			continue
		}
		common.Send(t.ch, common.PhotoTo(0, path))
	}
}

func (t *Crocodile) sendSignalMessages(results []inspectionResult) {
	for _, result := range results {
		item := result.Signal.Item
		name := displayName(item)
		link := "https://www.coingecko.com/en/coins/" + url.PathEscape(item.ID)
		msg := fmt.Sprintf("[%s触发买入信号](%s)", escapeMarkdownLinkText(name), link)
		common.Send(t.ch, common.Markdown(msg, false))
	}
}

func (t *Crocodile) renderSignalChart(sig Signal, klines []crypto.CoingeckoKline) (string, error) {
	dir := filepath.Join(os.TempDir(), "aio-crocodile-chart")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", err
	}

	name := safeChartFileName(sig.Item.ID)
	if name == "" {
		name = "unknown"
	}
	path := filepath.Join(dir, fmt.Sprintf("%s_%s.png", name, sig.Time.Format("2006-01-02")))
	return path, renderChart(path, sig, klines)
}

func renderChart(path string, sig Signal, klines []crypto.CoingeckoKline) error {
	klines = recentKlines(klines, chartKlineLimit)
	if len(klines) < 2 {
		return fmt.Errorf("kline data is insufficient")
	}

	minPrice, maxPrice, maxVolume := chartBounds(klines)
	if minPrice <= 0 || maxPrice <= minPrice || maxVolume <= 0 {
		return fmt.Errorf("kline range is invalid")
	}

	const (
		width        = 1100
		height       = 740
		left         = 70
		right        = 35
		headerTop    = 24
		priceTop     = 100
		priceBottom  = 505
		volumeTop    = 545
		volumeBottom = 690
	)

	img := image.NewRGBA(image.Rect(0, 0, width, height))
	fillRect(img, 0, 0, width-1, height-1, chartBackground)
	fillRect(img, left, priceTop, width-right, priceBottom, chartPanel)
	fillRect(img, left, volumeTop, width-right, volumeBottom, chartPanel)

	drawLabel(img, left, headerTop, fmt.Sprintf("%s (%s)", displayName(sig.Item), sig.Item.ID), chartText)
	cfg := sig.RuleConfig.normalize()
	drawLabel(img, left, headerTop+24, fmt.Sprintf("today: %.1fx >= %.1fx", sig.YesterdayRatio, cfg.YesterdayMultiple), chartSubText)
	drawLabel(img, left, headerTop+46, fmt.Sprintf("avg: %.1fx >= %.1fx", sig.AverageRatio, cfg.AverageMultiple), chartSubText)

	for i := 0; i <= 5; i++ {
		y := priceTop + (priceBottom-priceTop)*i/5
		drawLine(img, left, y, width-right, y, chartGrid)
		vy := volumeTop + (volumeBottom-volumeTop)*i/5
		drawLine(img, left, vy, width-right, vy, chartGrid)
	}
	drawLine(img, left, priceTop, left, priceBottom, chartAxis)
	drawLine(img, left, priceBottom, width-right, priceBottom, chartAxis)
	drawLine(img, left, volumeBottom, width-right, volumeBottom, chartAxis)

	priceRange := maxPrice - minPrice
	priceY := func(price float64) int {
		return priceBottom - int((price-minPrice)/priceRange*float64(priceBottom-priceTop))
	}
	volumeY := func(volume float64) int {
		return volumeBottom - int(volume/maxVolume*float64(volumeBottom-volumeTop))
	}

	plotWidth := width - left - right
	step := float64(plotWidth) / float64(len(klines))
	bodyWidth := int(step * 0.55)
	if bodyWidth < 4 {
		bodyWidth = 4
	}

	triggerDay := sig.Time.UTC().Format("2006-01-02")
	for i, kline := range klines {
		x := left + int((float64(i)+0.5)*step)
		c := chartUp
		if kline.Close < kline.Open {
			c = chartDown
		}
		if kline.Time.UTC().Format("2006-01-02") == triggerDay {
			c = chartHighlight
		}

		drawLine(img, x, priceY(kline.High), x, priceY(kline.Low), c)
		y0 := priceY(kline.Open)
		y1 := priceY(kline.Close)
		if y0 == y1 {
			y1++
		}
		fillRect(img, x-bodyWidth/2, y0, x+bodyWidth/2, y1, c)

		vc := chartVolume
		if kline.Time.UTC().Format("2006-01-02") == triggerDay {
			vc = chartHighlight
		}
		fillRect(img, x-bodyWidth/2, volumeY(kline.Volume), x+bodyWidth/2, volumeBottom, vc)
	}

	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return png.Encode(f, img)
}

func recentKlines(klines []crypto.CoingeckoKline, limit int) []crypto.CoingeckoKline {
	out := append([]crypto.CoingeckoKline(nil), klines...)
	sort.Slice(out, func(i, j int) bool {
		return out[i].Time.Before(out[j].Time)
	})
	if limit > 0 && len(out) > limit {
		return out[len(out)-limit:]
	}
	return out
}

func chartBounds(klines []crypto.CoingeckoKline) (float64, float64, float64) {
	minPrice := math.MaxFloat64
	maxPrice := 0.0
	maxVolume := 0.0
	for _, kline := range klines {
		if kline.Low < minPrice {
			minPrice = kline.Low
		}
		if kline.High > maxPrice {
			maxPrice = kline.High
		}
		if kline.Volume > maxVolume {
			maxVolume = kline.Volume
		}
	}

	padding := (maxPrice - minPrice) * 0.08
	return minPrice - padding, maxPrice + padding, maxVolume
}

func drawLabel(img *image.RGBA, x, y int, label string, c color.RGBA) {
	d := &font.Drawer{
		Dst:  img,
		Src:  image.NewUniform(c),
		Face: basicfont.Face7x13,
		Dot:  fixed.P(x, y+13),
	}
	d.DrawString(label)
}

func fillRect(img *image.RGBA, x0, y0, x1, y1 int, c color.RGBA) {
	if x0 > x1 {
		x0, x1 = x1, x0
	}
	if y0 > y1 {
		y0, y1 = y1, y0
	}
	bounds := img.Bounds()
	if x0 < bounds.Min.X {
		x0 = bounds.Min.X
	}
	if y0 < bounds.Min.Y {
		y0 = bounds.Min.Y
	}
	if x1 >= bounds.Max.X {
		x1 = bounds.Max.X - 1
	}
	if y1 >= bounds.Max.Y {
		y1 = bounds.Max.Y - 1
	}
	for y := y0; y <= y1; y++ {
		for x := x0; x <= x1; x++ {
			img.SetRGBA(x, y, c)
		}
	}
}

func drawLine(img *image.RGBA, x0, y0, x1, y1 int, c color.RGBA) {
	dx := int(math.Abs(float64(x1 - x0)))
	sx := -1
	if x0 < x1 {
		sx = 1
	}
	dy := -int(math.Abs(float64(y1 - y0)))
	sy := -1
	if y0 < y1 {
		sy = 1
	}
	err := dx + dy
	for {
		if image.Pt(x0, y0).In(img.Bounds()) {
			img.SetRGBA(x0, y0, c)
		}
		if x0 == x1 && y0 == y1 {
			break
		}
		e2 := 2 * err
		if e2 >= dy {
			err += dy
			x0 += sx
		}
		if e2 <= dx {
			err += dx
			y0 += sy
		}
	}
}

func safeChartFileName(name string) string {
	name = strings.ToLower(strings.TrimSpace(name))
	replacer := strings.NewReplacer("/", "-", "\\", "-", ":", "-", "*", "-", "?", "-", "\"", "-", "<", "-", ">", "-", "|", "-")
	return replacer.Replace(name)
}

func escapeMarkdownLinkText(text string) string {
	replacer := strings.NewReplacer(
		"\\", "\\\\",
		"[", "\\[",
		"]", "\\]",
	)
	return replacer.Replace(text)
}

func logChartError(item Item, err error) {
	fmt.Printf("crocodile render chart %s failed: %v\n", displayName(item), err)
}

func cleanupOldCharts(dir string, maxAge time.Duration) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return
	}
	cutoff := time.Now().Add(-maxAge)
	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil || info.IsDir() || info.ModTime().After(cutoff) {
			continue
		}
		_ = os.Remove(filepath.Join(dir, entry.Name()))
	}
}
