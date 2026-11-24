package market

type Candle struct {
	StartMS       int64   `json:"startMs"`
	EndMS         int64   `json:"endMs"`
	Open          float64 `json:"open"`
	High          float64 `json:"high"`
	Low           float64 `json:"low"`
	Close         float64 `json:"close"`
	Volume        float64 `json:"volume"`
	QuoteVolume   float64 `json:"quoteVolume"`
	TakerBuyQuote float64 `json:"takerBuyQuote"`
}

type MACDValue struct {
	DIF  float64 `json:"dif"`
	DEA  float64 `json:"dea"`
	Hist float64 `json:"hist"`
}

type Indicators struct {
	EMA20 *float64   `json:"ema20"`
	EMA50 *float64   `json:"ema50"`
	RSI14 *float64   `json:"rsi14"`
	MACD  *MACDValue `json:"macd"`
	ATR14 *float64   `json:"atr14"`
}

type Snapshot struct {
	Symbol                 string                `json:"symbol"`
	Price                  float64               `json:"price"`
	MarkPrice              float64               `json:"markPrice"`
	QuoteTimeMS            int64                 `json:"quoteTimeMs"`
	QuoteAgeSeconds        float64               `json:"quoteAgeSeconds"`
	FundingRatePercent     float64               `json:"fundingRatePercent"`
	NextFundingMS          int64                 `json:"nextFundingMs"`
	Change24HPercent       *float64              `json:"change24hPercent"`
	Change1MPercent        *float64              `json:"change1mPercent"`
	Change5MPercent        *float64              `json:"change5mPercent"`
	VolumeRatio1M          *float64              `json:"volumeRatio1m"`
	Support                float64               `json:"support"`
	Resistance             float64               `json:"resistance"`
	FuturesTakerNet1HUSDT  *float64              `json:"futuresTakerNet1hUsdt"`
	FuturesTakerNet4HUSDT  *float64              `json:"futuresTakerNet4hUsdt"`
	SpotTakerNet1HUSDT     *float64              `json:"spotTakerNet1hUsdt"`
	OpenInterestETH        *float64              `json:"openInterestEth"`
	OpenInterestUSDTApprox *float64              `json:"openInterestUsdtApprox"`
	OpenInterestChange1H   *float64              `json:"openInterestChange1hPercent"`
	LongAccountsPercent    *float64              `json:"longAccountsPercent"`
	ShortAccountsPercent   *float64              `json:"shortAccountsPercent"`
	Indicators             map[string]Indicators `json:"indicators"`
	Missing                []string              `json:"missing"`
}
