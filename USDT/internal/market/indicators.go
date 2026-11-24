package market

import "math"

func EMA(values []float64, period int) []float64 {
	if period < 1 || len(values) < period {
		return nil
	}
	sum := 0.0
	for _, value := range values[:period] {
		sum += value
	}
	out := []float64{sum / float64(period)}
	alpha := 2 / float64(period+1)
	for _, value := range values[period:] {
		out = append(out, (value-out[len(out)-1])*alpha+out[len(out)-1])
	}
	return out
}

func RSI(values []float64, period int) *float64 {
	if len(values) < period+1 {
		return nil
	}
	gains, losses := 0.0, 0.0
	for i := 1; i <= period; i++ {
		change := values[i] - values[i-1]
		gains += math.Max(change, 0)
		losses += math.Max(-change, 0)
	}
	avgGain, avgLoss := gains/float64(period), losses/float64(period)
	for i := period + 1; i < len(values); i++ {
		change := values[i] - values[i-1]
		avgGain = (avgGain*float64(period-1) + math.Max(change, 0)) / float64(period)
		avgLoss = (avgLoss*float64(period-1) + math.Max(-change, 0)) / float64(period)
	}
	result := 50.0
	if avgLoss == 0 {
		if avgGain > 0 {
			result = 100
		}
	} else {
		rs := avgGain / avgLoss
		result = 100 - 100/(1+rs)
	}
	return &result
}

func MACD(values []float64) *MACDValue {
	fast, slow := EMA(values, 12), EMA(values, 26)
	if len(fast) == 0 || len(slow) == 0 {
		return nil
	}
	series := make([]float64, len(slow))
	for i := range slow {
		series[i] = fast[i+14] - slow[i]
	}
	signal := EMA(series, 9)
	if len(signal) == 0 {
		return nil
	}
	return &MACDValue{
		DIF:  series[len(series)-1],
		DEA:  signal[len(signal)-1],
		Hist: series[len(series)-1] - signal[len(signal)-1],
	}
}

func ATR(candles []Candle, period int) *float64 {
	if len(candles) < period+1 {
		return nil
	}
	ranges := make([]float64, 0, len(candles)-1)
	for i := 1; i < len(candles); i++ {
		previous, current := candles[i-1], candles[i]
		ranges = append(ranges, math.Max(
			current.High-current.Low,
			math.Max(
				math.Abs(current.High-previous.Close),
				math.Abs(current.Low-previous.Close),
			),
		))
	}
	result := 0.0
	for _, value := range ranges[:period] {
		result += value
	}
	result /= float64(period)
	for _, value := range ranges[period:] {
		result = (result*float64(period-1) + value) / float64(period)
	}
	return &result
}

func PercentChange(latest, previous float64) *float64 {
	if previous == 0 {
		return nil
	}
	result := (latest/previous - 1) * 100
	return &result
}

func TakerNetQuote(candles []Candle, periods int) *float64 {
	if len(candles) < periods {
		return nil
	}
	result := 0.0
	for _, candle := range candles[len(candles)-periods:] {
		result += 2*candle.TakerBuyQuote - candle.QuoteVolume
	}
	return &result
}

func ComputeIndicators(candles []Candle) Indicators {
	closes := make([]float64, len(candles))
	for i, candle := range candles {
		closes[i] = candle.Close
	}
	ema20, ema50 := EMA(closes, 20), EMA(closes, 50)
	result := Indicators{RSI14: RSI(closes, 14), MACD: MACD(closes), ATR14: ATR(candles, 14)}
	if len(ema20) > 0 {
		result.EMA20 = &ema20[len(ema20)-1]
	}
	if len(ema50) > 0 {
		result.EMA50 = &ema50[len(ema50)-1]
	}
	return result
}
