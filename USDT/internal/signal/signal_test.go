package signal

import (
	"testing"

	"github.com/ycc1/my-mockCode/USDT/internal/market"
)

func pointer(value float64) *float64 { return &value }

func TestJudgeBullish(t *testing.T) {
	snapshot := &market.Snapshot{
		Price: 100,
		Indicators: map[string]market.Indicators{
			"1h":  {EMA20: pointer(101), EMA50: pointer(99)},
			"15m": {EMA20: pointer(99), RSI14: pointer(60), ATR14: pointer(2), MACD: &market.MACDValue{Hist: 1}},
		},
	}
	plan := Judge(snapshot)
	if plan.Direction != "偏多" || plan.Score != 5 {
		t.Fatalf("unexpected plan: %+v", plan)
	}
}

func TestJudgeExtremeReturnsNeutral(t *testing.T) {
	snapshot := &market.Snapshot{
		Price: 100,
		Indicators: map[string]market.Indicators{
			"1h":  {EMA20: pointer(101), EMA50: pointer(99)},
			"15m": {EMA20: pointer(99), RSI14: pointer(80), MACD: &market.MACDValue{Hist: 1}},
		},
	}
	plan := Judge(snapshot)
	if plan.Direction != "觀望" || !plan.Extreme {
		t.Fatalf("unexpected plan: %+v", plan)
	}
}
