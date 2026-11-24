package signal

import (
	"math"

	"github.com/ycc1/my-mockCode/USDT/internal/market"
)

type Plan struct {
	Direction       string   `json:"direction"`
	Score           int      `json:"score"`
	Reasons         []string `json:"reasons"`
	EntryReference  *float64 `json:"entryReference"`
	StopReference   *float64 `json:"stopReference"`
	TargetReference *float64 `json:"targetReference"`
	Condition       string   `json:"condition"`
	Extreme         bool     `json:"extreme"`
}

func Judge(snapshot *market.Snapshot) Plan {
	hour := snapshot.Indicators["1h"]
	quarter := snapshot.Indicators["15m"]
	score := 0
	reasons := []string{}
	if hour.EMA20 != nil && hour.EMA50 != nil {
		if *hour.EMA20 > *hour.EMA50 {
			score += 2
			reasons = append(reasons, "1H EMA20 高於 EMA50")
		} else if *hour.EMA20 < *hour.EMA50 {
			score -= 2
			reasons = append(reasons, "1H EMA20 低於 EMA50")
		}
	}
	if quarter.EMA20 != nil {
		if snapshot.Price > *quarter.EMA20 {
			score++
			reasons = append(reasons, "價格高於 15M EMA20")
		} else {
			score--
			reasons = append(reasons, "價格低於 15M EMA20")
		}
	}
	if quarter.MACD != nil {
		if quarter.MACD.Hist > 0 {
			score++
			reasons = append(reasons, "15M MACD 柱為正")
		} else {
			score--
			reasons = append(reasons, "15M MACD 柱為負")
		}
	}
	if quarter.RSI14 != nil {
		if *quarter.RSI14 >= 55 {
			score++
			reasons = append(reasons, "15M RSI 偏強")
		} else if *quarter.RSI14 <= 45 {
			score--
			reasons = append(reasons, "15M RSI 偏弱")
		}
	}
	extreme := quarter.RSI14 != nil && (*quarter.RSI14 >= 75 || *quarter.RSI14 <= 25)
	direction := "觀望"
	plan := Plan{Score: score, Reasons: reasons, Extreme: extreme}
	atr := snapshot.Price * .008
	if quarter.ATR14 != nil {
		atr = *quarter.ATR14
	}
	switch {
	case score >= 3 && !extreme:
		direction = "偏多"
		trigger := snapshot.Price
		if quarter.EMA20 != nil {
			trigger = math.Max(trigger, *quarter.EMA20)
		}
		entry := math.Round(trigger*10) / 10
		stop := math.Round((entry-math.Max(atr*1.3, snapshot.Price*.005))*10) / 10
		target := math.Round((entry+(entry-stop)*1.8)*10) / 10
		plan.EntryReference, plan.StopReference, plan.TargetReference = &entry, &stop, &target
		plan.Condition = "等 15M 收盤站穩 EMA20，且量能不萎縮；不追瞬間刺穿。"
	case score <= -3 && !extreme:
		direction = "偏空"
		trigger := snapshot.Price
		if quarter.EMA20 != nil {
			trigger = math.Min(trigger, *quarter.EMA20)
		}
		entry := math.Round(trigger*10) / 10
		stop := math.Round((entry+math.Max(atr*1.3, snapshot.Price*.005))*10) / 10
		target := math.Round((entry-(stop-entry)*1.8)*10) / 10
		plan.EntryReference, plan.StopReference, plan.TargetReference = &entry, &stop, &target
		plan.Condition = "等反彈測試 15M EMA20 受阻；不在低點直接追空。"
	default:
		plan.Condition = "不開新倉，等待 15M 與 1H 方向一致。"
	}
	plan.Direction = direction
	return plan
}
