"""Pure-Python technical indicators. Use only CLOSED exchange candles."""
from __future__ import annotations
from statistics import mean


def ema(values: list[float], period: int) -> list[float]:
    if period < 1 or len(values) < period:
        return []
    alpha = 2 / (period + 1)
    result = [mean(values[:period])]
    for v in values[period:]:
        result.append((v - result[-1]) * alpha + result[-1])
    return result


def rsi(values: list[float], period: int = 14) -> float | None:
    if len(values) < period + 1:
        return None
    changes = [values[i] - values[i - 1] for i in range(1, len(values))]
    gains = [max(c, 0) for c in changes]
    losses = [max(-c, 0) for c in changes]
    avg_gain = mean(gains[:period])
    avg_loss = mean(losses[:period])
    for g, l in zip(gains[period:], losses[period:]):
        avg_gain = (avg_gain * (period - 1) + g) / period
        avg_loss = (avg_loss * (period - 1) + l) / period
    if avg_loss == 0:
        return 100.0 if avg_gain else 50.0
    rs = avg_gain / avg_loss
    return 100 - (100 / (1 + rs))


def macd(values: list[float]) -> dict[str, float] | None:
    fast = ema(values, 12)
    slow = ema(values, 26)
    if not fast or not slow:
        return None
    # fast starts at source index 11, slow at source index 25.
    series = [fast[i + 14] - slow[i] for i in range(len(slow))]
    signal = ema(series, 9)
    if not signal:
        return None
    return {"dif": series[-1], "dea": signal[-1], "hist": series[-1] - signal[-1]}


def atr(candles: list[dict], period: int = 14) -> float | None:
    if len(candles) < period + 1:
        return None
    true_ranges = []
    for prev, current in zip(candles, candles[1:]):
        true_ranges.append(max(current["high"] - current["low"],
                               abs(current["high"] - prev["close"]),
                               abs(current["low"] - prev["close"])))
    result = mean(true_ranges[:period])
    for tr in true_ranges[period:]:
        result = (result * (period - 1) + tr) / period
    return result


def pct_change(latest: float, previous: float | None) -> float | None:
    return None if previous is None or previous == 0 else (latest / previous - 1) * 100


def sum_net_taker_quote(candles: list[dict], periods: int) -> float | None:
    subset = candles[-periods:]
    if len(subset) < periods:
        return None
    return sum(2 * c["taker_buy_quote"] - c["quote_volume"] for c in subset)


def compute_indicators(candles: list[dict]) -> dict:
    closes = [c["close"] for c in candles]
    fast = ema(closes, 20)
    slow = ema(closes, 50)
    return {
        "ema20": fast[-1] if fast else None,
        "ema50": slow[-1] if slow else None,
        "rsi14": rsi(closes),
        "macd": macd(closes),
        "atr14": atr(candles),
    }
