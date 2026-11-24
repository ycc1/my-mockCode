package market

import (
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

const (
	futuresBase = "https://fapi.binance.com"
	spotBase    = "https://api.binance.com"
)

type Client struct {
	HTTPClient *http.Client
	Symbol     string
	MaxAge     time.Duration
}

func NewClient(symbol string, maxAge time.Duration) *Client {
	return &Client{
		HTTPClient: &http.Client{Timeout: 12 * time.Second},
		Symbol:     symbol,
		MaxAge:     maxAge,
	}
}

func (c *Client) get(base, path string, params url.Values, target any) error {
	endpoint := base + path
	if len(params) > 0 {
		endpoint += "?" + params.Encode()
	}
	request, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return err
	}
	request.Header.Set("User-Agent", "Megwin-ETHMonitor-Go/1.0")
	response, err := c.HTTPClient.Do(request)
	if err != nil {
		return fmt.Errorf("取得行情 %s 失敗: %w", path, err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("取得行情 %s 失敗: HTTP %d", path, response.StatusCode)
	}
	if err := json.NewDecoder(response.Body).Decode(target); err != nil {
		return fmt.Errorf("解析行情 %s 失敗: %w", path, err)
	}
	return nil
}

func (c *Client) candles(interval string, market string) ([]Candle, error) {
	base, path := futuresBase, "/fapi/v1/klines"
	if market == "spot" {
		base, path = spotBase, "/api/v3/klines"
	}
	var raw [][]json.RawMessage
	err := c.get(base, path, url.Values{
		"symbol":   {c.Symbol},
		"interval": {interval},
		"limit":    {"200"},
	}, &raw)
	if err != nil {
		return nil, err
	}
	return parseCandles(raw, time.Now(), intervalDuration(interval))
}

func parseCandles(raw [][]json.RawMessage, now time.Time, interval time.Duration) ([]Candle, error) {
	candles := make([]Candle, 0, len(raw))
	nowMS := now.UnixMilli()
	for _, values := range raw {
		if len(values) < 11 {
			continue
		}
		var candle Candle
		parts := []*int64{&candle.StartMS, &candle.EndMS}
		if err := json.Unmarshal(values[0], parts[0]); err != nil {
			continue
		}
		if err := json.Unmarshal(values[6], parts[1]); err != nil || candle.EndMS >= nowMS {
			continue
		}
		numbers := []struct {
			index int
			value *float64
		}{
			{1, &candle.Open}, {2, &candle.High}, {3, &candle.Low},
			{4, &candle.Close}, {5, &candle.Volume}, {7, &candle.QuoteVolume},
			{10, &candle.TakerBuyQuote},
		}
		valid := true
		for _, number := range numbers {
			text := ""
			if err := json.Unmarshal(values[number.index], &text); err != nil {
				valid = false
				break
			}
			parsed, err := strconv.ParseFloat(text, 64)
			if err != nil {
				valid = false
				break
			}
			*number.value = parsed
		}
		if valid {
			candles = append(candles, candle)
		}
	}
	if len(candles) < 60 {
		return nil, fmt.Errorf("已收盤 K 線不足 60 根，拒絕輸出訊號")
	}
	if nowMS-candles[len(candles)-1].EndMS > interval.Milliseconds()*2+5000 {
		return nil, fmt.Errorf("%s K 線資料過期", interval)
	}
	return candles, nil
}

func intervalDuration(value string) time.Duration {
	switch value {
	case "15m":
		return 15 * time.Minute
	case "1h":
		return time.Hour
	case "4h":
		return 4 * time.Hour
	default:
		return time.Minute
	}
}

func (c *Client) optional(base, path string, params url.Values, target any) bool {
	return c.get(base, path, params, target) == nil
}

func (c *Client) Snapshot(includeSpot bool) (*Snapshot, error) {
	var ticker struct {
		Price string `json:"price"`
		Time  int64  `json:"time"`
	}
	if err := c.get(futuresBase, "/fapi/v2/ticker/price", url.Values{"symbol": {c.Symbol}}, &ticker); err != nil {
		return nil, err
	}
	var premium struct {
		MarkPrice       string `json:"markPrice"`
		LastFundingRate string `json:"lastFundingRate"`
		NextFundingTime int64  `json:"nextFundingTime"`
		Time            int64  `json:"time"`
	}
	if err := c.get(futuresBase, "/fapi/v1/premiumIndex", url.Values{"symbol": {c.Symbol}}, &premium); err != nil {
		return nil, err
	}
	price, errPrice := strconv.ParseFloat(ticker.Price, 64)
	markPrice, errMark := strconv.ParseFloat(premium.MarkPrice, 64)
	funding, errFunding := strconv.ParseFloat(premium.LastFundingRate, 64)
	now := time.Now()
	age := now.Sub(time.UnixMilli(ticker.Time))
	if errPrice != nil || errMark != nil || errFunding != nil || price <= 0 || math.Abs(age.Seconds()) > c.MaxAge.Seconds() {
		return nil, fmt.Errorf("報價無效或超過 %.0f 秒", c.MaxAge.Seconds())
	}
	if now.Sub(time.UnixMilli(premium.Time)) > c.MaxAge {
		return nil, fmt.Errorf("標記價格時間戳過期")
	}

	intervals := make(map[string][]Candle)
	for _, interval := range []string{"1m", "15m", "1h", "4h"} {
		candles, err := c.candles(interval, "futures")
		if err != nil {
			return nil, err
		}
		intervals[interval] = candles
	}
	indicators := make(map[string]Indicators, len(intervals))
	for interval, candles := range intervals {
		indicators[interval] = ComputeIndicators(candles)
	}
	oneMinute := intervals["1m"]
	latestClosed := oneMinute[len(oneMinute)-1]
	baseline := 0.0
	for _, candle := range oneMinute[len(oneMinute)-21 : len(oneMinute)-1] {
		baseline += candle.QuoteVolume
	}
	baseline /= 20
	var volumeRatio *float64
	if baseline > 0 {
		value := latestClosed.QuoteVolume / baseline
		volumeRatio = &value
	}
	levels := intervals["15m"][len(intervals["15m"])-21 : len(intervals["15m"])-1]
	support, resistance := levels[0].Low, levels[0].High
	for _, candle := range levels[1:] {
		support, resistance = math.Min(support, candle.Low), math.Max(resistance, candle.High)
	}

	snapshot := &Snapshot{
		Symbol:                c.Symbol,
		Price:                 price,
		MarkPrice:             markPrice,
		QuoteTimeMS:           ticker.Time,
		QuoteAgeSeconds:       math.Abs(age.Seconds()),
		FundingRatePercent:    funding * 100,
		NextFundingMS:         premium.NextFundingTime,
		Change1MPercent:       PercentChange(price, latestClosed.Close),
		Change5MPercent:       PercentChange(price, oneMinute[len(oneMinute)-5].Close),
		VolumeRatio1M:         volumeRatio,
		Support:               support,
		Resistance:            resistance,
		FuturesTakerNet1HUSDT: TakerNetQuote(oneMinute, 60),
		FuturesTakerNet4HUSDT: TakerNetQuote(intervals["1h"], 4),
		Indicators:            indicators,
		Missing:               []string{},
	}

	var daily struct {
		PriceChangePercent string `json:"priceChangePercent"`
	}
	if c.optional(futuresBase, "/fapi/v1/ticker/24hr", url.Values{"symbol": {c.Symbol}}, &daily) {
		if value, err := strconv.ParseFloat(daily.PriceChangePercent, 64); err == nil {
			snapshot.Change24HPercent = &value
		}
	}

	var oi struct {
		OpenInterest string `json:"openInterest"`
		Time         int64  `json:"time"`
	}
	if c.optional(futuresBase, "/fapi/v1/openInterest", url.Values{"symbol": {c.Symbol}}, &oi) &&
		now.Sub(time.UnixMilli(oi.Time)) <= c.MaxAge {
		if value, err := strconv.ParseFloat(oi.OpenInterest, 64); err == nil {
			snapshot.OpenInterestETH = &value
			approx := value * markPrice
			snapshot.OpenInterestUSDTApprox = &approx
		}
	}
	if snapshot.OpenInterestETH == nil {
		snapshot.Missing = append(snapshot.Missing, "OI")
	}

	var oiHistory []struct {
		SumOpenInterest string `json:"sumOpenInterest"`
		Timestamp       int64  `json:"timestamp"`
	}
	if c.optional(futuresBase, "/futures/data/openInterestHist", url.Values{
		"symbol": {c.Symbol}, "period": {"5m"}, "limit": {"14"},
	}, &oiHistory) && len(oiHistory) >= 13 &&
		now.Sub(time.UnixMilli(oiHistory[len(oiHistory)-1].Timestamp)) <= 20*time.Minute {
		last, e1 := strconv.ParseFloat(oiHistory[len(oiHistory)-1].SumOpenInterest, 64)
		before, e2 := strconv.ParseFloat(oiHistory[len(oiHistory)-13].SumOpenInterest, 64)
		if e1 == nil && e2 == nil {
			snapshot.OpenInterestChange1H = PercentChange(last, before)
		}
	}
	if snapshot.OpenInterestChange1H == nil {
		snapshot.Missing = append(snapshot.Missing, "OI 1H 變化")
	}

	var ratio []struct {
		LongAccount  string `json:"longAccount"`
		ShortAccount string `json:"shortAccount"`
		Timestamp    int64  `json:"timestamp"`
	}
	if c.optional(futuresBase, "/futures/data/globalLongShortAccountRatio", url.Values{
		"symbol": {c.Symbol}, "period": {"5m"}, "limit": {"2"},
	}, &ratio) && len(ratio) > 0 &&
		now.Sub(time.UnixMilli(ratio[len(ratio)-1].Timestamp)) <= 20*time.Minute {
		last := ratio[len(ratio)-1]
		longValue, e1 := strconv.ParseFloat(last.LongAccount, 64)
		shortValue, e2 := strconv.ParseFloat(last.ShortAccount, 64)
		if e1 == nil && e2 == nil {
			longValue, shortValue = longValue*100, shortValue*100
			snapshot.LongAccountsPercent = &longValue
			snapshot.ShortAccountsPercent = &shortValue
		}
	}
	if snapshot.LongAccountsPercent == nil {
		snapshot.Missing = append(snapshot.Missing, "多空帳戶比")
	}

	if includeSpot {
		spotCandles, err := c.candles("1m", "spot")
		if err == nil {
			snapshot.SpotTakerNet1HUSDT = TakerNetQuote(spotCandles, 60)
		} else {
			snapshot.Missing = append(snapshot.Missing, "現貨主動買賣差")
		}
	}
	return snapshot, nil
}
