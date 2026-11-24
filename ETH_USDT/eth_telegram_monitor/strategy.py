"""Transparent rule-based signal. Not trade execution or an assurance of returns."""
from __future__ import annotations
from indicators import pct_change


def judge(s: dict) -> dict:
    price = s["price"]
    ind = s["indicators"]
    h = ind["1h"]
    q = ind["15m"]
    m = q["macd"] or {}
    score = 0
    reasons = []
    if h["ema20"] is not None and h["ema50"] is not None:
        if h["ema20"] > h["ema50"]:
            score += 2
            reasons.append("1H EMA20 高於 EMA50")
        elif h["ema20"] < h["ema50"]:
            score -= 2
            reasons.append("1H EMA20 低於 EMA50")
    if q["ema20"] is not None:
        score += 1 if price > q["ema20"] else -1
        reasons.append("價格高於 15M EMA20" if price > q["ema20"] else "價格低於 15M EMA20")
    if m:
        score += 1 if m["hist"] > 0 else -1
        reasons.append("15M MACD 柱為正" if m["hist"] > 0 else "15M MACD 柱為負")
    r = q["rsi14"]
    if r is not None:
        if r >= 55:
            score += 1
            reasons.append("15M RSI 偏強")
        elif r <= 45:
            score -= 1
            reasons.append("15M RSI 偏弱")
    # Avoid directional signals near extremes: momentum may reverse abruptly.
    extreme = r is not None and (r >= 75 or r <= 25)
    if score >= 3 and not extreme:
        signal = "偏多"
    elif score <= -3 and not extreme:
        signal = "偏空"
    else:
        signal = "觀望"
    atr = q["atr14"] or price * 0.008
    # Quote levels in USD only, rounded to a reasonable 0.1 precision.
    if signal == "偏多":
        # Use breakout confirmation rather than buying into resistance.
        trigger = max(price, q["ema20"] or price)
        entry = round(trigger, 1)
        stop = round(entry - max(atr * 1.3, price * .005), 1)
        target = round(entry + (entry - stop) * 1.8, 1)
        condition = "等 15M 收盤站穩 EMA20，且量能不萎縮；不追瞬間刺穿。"
    elif signal == "偏空":
        trigger = min(price, q["ema20"] or price)
        entry = round(trigger, 1)
        stop = round(entry + max(atr * 1.3, price * .005), 1)
        target = round(entry - (stop - entry) * 1.8, 1)
        condition = "等反彈測試 15M EMA20 受阻；不在低點直接追空。"
    else:
        entry = stop = target = None
        condition = "不開新倉，等待 15M 與 1H 方向一致。"
    return {"direction": signal, "score": score, "reasons": reasons,
            "entry_reference": entry, "stop_reference": stop, "target_reference": target,
            "condition": condition, "extreme": extreme}


def alerts(s: dict, previous: dict | None, thresholds: dict) -> list[str]:
    found = []
    p = s["price"]
    prev_p = previous.get("last_price") if previous else None
    if prev_p:
        minute_delta = pct_change(p, prev_p)
        if minute_delta is not None and abs(minute_delta) >= thresholds["price_pct_1m"]:
            found.append(f"相較上次檢查價格變動 {minute_delta:+.2f}%")
    c5 = s.get("change_5m_pct")
    if c5 is not None and abs(c5) >= thresholds["price_pct_5m"]:
        found.append(f"近 5 分鐘價格變動 {c5:+.2f}%")
    vr = s.get("volume_ratio_1m")
    if vr is not None and vr >= thresholds["volume_multiple"]:
        found.append(f"上一根已收盤 1M 量能達均量 {vr:.1f} 倍")
    # Compare price against previous snapshot's support/resistance to avoid moving thresholds.
    if prev_p and previous:
        upper = previous.get("resistance")
        lower = previous.get("support")
        if upper and prev_p <= upper < p:
            found.append(f"上破先前阻力 {upper:,.2f}")
        if lower and prev_p >= lower > p:
            found.append(f"跌破先前支撐 {lower:,.2f}")
    if abs(s.get("funding_rate_percent") or 0) >= thresholds["funding_abs_pct"]:
        found.append(f"資金費率偏高：{s['funding_rate_percent']:+.4f}%")
    return found
