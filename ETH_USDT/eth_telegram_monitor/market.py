"""Binance USDⓈ-M public market data; Binance keys are NOT required."""
from __future__ import annotations
import time
import requests
from requests.adapters import HTTPAdapter
from urllib3.util.retry import Retry
from indicators import compute_indicators, pct_change, sum_net_taker_quote

FUTURES = "https://fapi.binance.com"
SPOT = "https://api.binance.com"

class MarketError(Exception):
    pass


def get_session() -> requests.Session:
    s = requests.Session()
    retry = Retry(total=2, backoff_factor=0.4, status_forcelist=[429, 500, 502, 503, 504],
                  allowed_methods=["GET"], respect_retry_after_header=True)
    adapter = HTTPAdapter(max_retries=retry, pool_connections=10, pool_maxsize=10)
    s.mount("https://", adapter)
    s.headers.update({"User-Agent": "ETHMonitor/1.0"})
    return s


def _get(session: requests.Session, base: str, path: str, params: dict | None = None):
    try:
        res = session.get(base + path, params=params, timeout=(4, 8))
        res.raise_for_status()
        return res.json()
    except (requests.RequestException, ValueError) as exc:
        raise MarketError(f"行情來源 {path} 取得失敗（{type(exc).__name__}）") from None


def parse_candles(raw: list, now_ms: int, interval_ms: int) -> list[dict]:
    data = []
    for c in raw:
        if not isinstance(c, (list, tuple)) or len(c) < 11:
            continue
        # The endpoint includes the currently-forming candle: exclude it.
        if int(c[6]) >= now_ms:
            continue
        data.append({"start_ms": int(c[0]), "end_ms": int(c[6]),
                     "open": float(c[1]), "high": float(c[2]), "low": float(c[3]),
                     "close": float(c[4]), "volume": float(c[5]),
                     "quote_volume": float(c[7]), "taker_buy_quote": float(c[10])})
    if len(data) < 60:
        raise MarketError("已收盤 K 線不足 60 根，拒絕輸出交易訊號")
    if now_ms - data[-1]["end_ms"] > interval_ms * 2 + 5000:
        raise MarketError("K 線資料過期，拒絕使用舊資料")
    return data


def _candles(session, symbol: str, interval: str, now_ms: int, market: str = "futures"):
    minutes = {"1m": 1, "15m": 15, "1h": 60, "4h": 240}[interval]
    base = FUTURES if market == "futures" else SPOT
    path = "/fapi/v1/klines" if market == "futures" else "/api/v3/klines"
    raw = _get(session, base, path, {"symbol": symbol, "interval": interval, "limit": 200})
    return parse_candles(raw, now_ms, minutes * 60_000)


def _optional(session, base, path, params=None):
    try:
        return _get(session, base, path, params)
    except MarketError:
        return None


def snapshot(symbol: str, max_age_seconds: int = 90, include_spot: bool = True) -> dict:
    session = get_session()
    try:
        ticker = _get(session, FUTURES, "/fapi/v2/ticker/price", {"symbol": symbol})
        mark = _get(session, FUTURES, "/fapi/v1/premiumIndex", {"symbol": symbol})
        price = float(ticker["price"])
        ts_ms = int(ticker["time"])
        now_ms = int(time.time() * 1000)
        age = abs(now_ms - ts_ms) / 1000
        if not 0 < price or age > max_age_seconds:
            raise MarketError(f"報價無效或超過 {max_age_seconds} 秒（偏差 {age:.0f} 秒）")
        if abs(now_ms - int(mark["time"])) > max_age_seconds * 1000:
            raise MarketError("標記價格時間戳過期")
        intervals = {x: _candles(session, symbol, x, now_ms) for x in ["1m", "15m", "1h", "4h"]}
        stats = {interval: compute_indicators(c) for interval, c in intervals.items()}
        one = intervals["1m"]
        volume_baseline = sum(c["quote_volume"] for c in one[-21:-1]) / 20
        latest_closed = one[-1]
        change_5m = pct_change(price, one[-5]["close"])
        change_1m = pct_change(price, one[-1]["close"])
        levels = intervals["15m"][-21:-1]  # exclude most recent completed candle
        result = {
            "symbol": symbol, "price": price, "quote_time_ms": ts_ms,
            "quote_age_sec": round(age, 1),
            "mark_price": float(mark["markPrice"]),
            "funding_rate_percent": float(mark["lastFundingRate"]) * 100,
            "next_funding_ms": int(mark["nextFundingTime"]),
            "24h": _optional(session, FUTURES, "/fapi/v1/ticker/24hr", {"symbol": symbol}),
            "indicators": stats,
            "support": min(c["low"] for c in levels),
            "resistance": max(c["high"] for c in levels),
            "change_1m_pct": change_1m,
            "change_5m_pct": change_5m,
            "volume_ratio_1m": round(latest_closed["quote_volume"] / volume_baseline, 2) if volume_baseline else None,
            "futures_taker_net_1h_usdt": sum_net_taker_quote(one, 60),
            "futures_taker_net_4h_usdt": None,
            "spot_taker_net_1h_usdt": None,
            "oi_eth": None, "oi_usdt_approx": None, "oi_change_1h_pct": None,
            "long_accounts_percent": None, "short_accounts_percent": None,
            "long_short_timestamp_ms": None,
            "missing": [],
        }
        # 4H futures kline-based taker delta: four closed 1H candles.
        result["futures_taker_net_4h_usdt"] = sum_net_taker_quote(intervals["1h"], 4)
        oi = _optional(session, FUTURES, "/fapi/v1/openInterest", {"symbol": symbol})
        if oi and abs(now_ms - int(oi.get("time", 0))) <= max_age_seconds * 1000:
            result["oi_eth"] = float(oi["openInterest"])
            result["oi_usdt_approx"] = result["oi_eth"] * float(mark["markPrice"])
        else:
            result["missing"].append("OI")
        oih = _optional(session, FUTURES, "/futures/data/openInterestHist",
                        {"symbol": symbol, "period": "5m", "limit": 14})
        if (isinstance(oih, list) and len(oih) >= 13
                and abs(now_ms - int(oih[-1].get("timestamp", 0))) <= 20 * 60_000):
            last = float(oih[-1]["sumOpenInterest"])
            before = float(oih[-13]["sumOpenInterest"])
            result["oi_change_1h_pct"] = pct_change(last, before)
        else:
            result["missing"].append("OI 1H 變化")
        ls = _optional(session, FUTURES, "/futures/data/globalLongShortAccountRatio",
                       {"symbol": symbol, "period": "5m", "limit": 2})
        if (isinstance(ls, list) and ls
                and abs(now_ms - int(ls[-1].get("timestamp", 0))) <= 20 * 60_000):
            result["long_accounts_percent"] = float(ls[-1]["longAccount"]) * 100
            result["short_accounts_percent"] = float(ls[-1]["shortAccount"]) * 100
            result["long_short_timestamp_ms"] = int(ls[-1]["timestamp"])
        else:
            result["missing"].append("多空帳戶比")
        if include_spot:
            try:
                spot = _candles(session, symbol, "1m", now_ms, market="spot")
                result["spot_taker_net_1h_usdt"] = sum_net_taker_quote(spot, 60)
            except MarketError:
                result["missing"].append("現貨主動買賣差")
        return result
    finally:
        session.close()
