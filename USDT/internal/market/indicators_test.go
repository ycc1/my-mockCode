package market

import (
	"math"
	"testing"
)

func TestIndicatorsRejectShortInput(t *testing.T) {
	if value := EMA([]float64{1, 2}, 20); value != nil {
		t.Fatal("expected nil EMA")
	}
	if value := RSI([]float64{1, 2}, 14); value != nil {
		t.Fatal("expected nil RSI")
	}
	if value := PercentChange(100, 0); value != nil {
		t.Fatal("expected nil percent change")
	}
}

func TestEMAAndPercentChange(t *testing.T) {
	values := []float64{1, 2, 3}
	ema := EMA(values, 3)
	if len(ema) != 1 || ema[0] != 2 {
		t.Fatalf("unexpected EMA: %v", ema)
	}
	change := PercentChange(110, 100)
	if change == nil || math.Abs(*change-10) > 1e-9 {
		t.Fatalf("unexpected change: %v", change)
	}
}
