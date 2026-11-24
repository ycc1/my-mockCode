"""Telegram Bot API and optional OpenAI Responses API."""
from __future__ import annotations
import json
from datetime import datetime
from zoneinfo import ZoneInfo
import requests

JST = ZoneInfo("Asia/Tokyo")


def now_jst(ms: int) -> str:
    return datetime.fromtimestamp(ms / 1000, JST).strftime("%Y-%m-%d %H:%M:%S JST")


def num(v, fmt=".2f") -> str:
    return "未取得" if v is None else format(v, fmt)


def short_usdt(v):
    if v is None:
        return "未取得"
    return f"{v/1e6:+,.2f} 百萬 USDT"


def build_base_message(s, signal, reasons, ai_result=None, news=None):
    i = s["indicators"]
    trend15 = i["15m"]
    h = i["1h"]
    q = i["4h"]
    lines = [
        f"📡 ETH/USDT 每分鐘行情｜{now_jst(s['quote_time_ms'])}",
        f"💰 最新合約成交價：{s['price']:,.2f} USDT（報價延遲 {s['quote_age_sec']} 秒）",
        f"📊 24H：{num(float(s['24h']['priceChangePercent']) if s.get('24h') else None, '+.2f')}%"
        f"｜5M：{num(s['change_5m_pct'], '+.2f')}%",
        f"方向：{signal['direction']}（規則分數 {signal['score']:+d}；非勝率）",
        f"15M RSI {num(trend15['rsi14'], '.1f')} / MACD柱 {num((trend15['macd'] or {}).get('hist'), '+.3f')}",
        f"1H RSI {num(h['rsi14'], '.1f')}｜4H RSI {num(q['rsi14'], '.1f')}",
        f"15M EMA20 {num(trend15['ema20'])} / EMA50 {num(trend15['ema50'])}",
        f"支撐 {s['support']:,.2f}｜阻力 {s['resistance']:,.2f}（近20根已收15M）",
        f"1M 量比 {num(s['volume_ratio_1m'])}x（上一根 vs 前20根）",
        f"1H 期貨主動買賣差 {short_usdt(s['futures_taker_net_1h_usdt'])}",
        f"4H 期貨主動買賣差 {short_usdt(s['futures_taker_net_4h_usdt'])}",
        f"1H 現貨主動買賣差 {short_usdt(s['spot_taker_net_1h_usdt'])}",
        f"OI ≈ {short_usdt(s['oi_usdt_approx'])}｜OI 1H {num(s['oi_change_1h_pct'], '+.2f')}%",
        f"多方帳戶 {num(s['long_accounts_percent'], '.1f')}% / 空方 {num(s['short_accounts_percent'], '.1f')}%（5M 比率）",
        f"資金費率 {num(s['funding_rate_percent'], '+.4f')}%（最新發布值，非固定8H）",
    ]
    if signal["direction"] == "觀望":
        lines.append("🎯 觀望：15M / 1H 不一致；先不進場")
    else:
        lines.append(f"🎯 {signal['direction']}參考 {signal['entry_reference']:,.1f}｜止損 {signal['stop_reference']:,.1f}｜目標 {signal['target_reference']:,.1f}")
    lines.append("觸發條件：" + signal["condition"])
    if reasons:
        lines.append("🚨 警示：" + "；".join(reasons[:4]))
    if ai_result:
        lines.extend(["", "🤖 GPT 解讀：", ai_result.strip()[:850]])
    elif ai_result is None:
        lines.append("🤖 GPT 解讀：未啟用或暫時不可用；上面是規則式分析")
    if news:
        lines.append("📰 最新新聞（標題≠已證實的跌漲原因）：")
        for n in news[:2]:
            lines.append(f"• {n['title'][:96]}（{n['published_utc'][:16]} UTC）")
            lines.append(n["url"])
    if s["missing"]:
        lines.append("⚠️ 未取得：" + "、".join(s["missing"]))
    lines.append("⚠️ 僅為行情觀察，不是自動交易或獲利保證。")
    return "\n".join(lines)[:3900]


def telegram_send(token: str, chat_id: str, text: str):
    # Do not log exceptions containing the token-bearing URL.
    try:
        r = requests.post(f"https://api.telegram.org/bot{token}/sendMessage",
                          json={"chat_id": chat_id, "text": text, "disable_web_page_preview": True},
                          timeout=(4, 12))
        if r.status_code != 200 or not r.json().get("ok"):
            raise RuntimeError(f"Telegram HTTP {r.status_code}; 檢查 token/chat_id/權限")
    except (requests.RequestException, ValueError) as e:
        raise RuntimeError(f"Telegram 發送失敗（{type(e).__name__}；不輸出機密）") from None


def discover_chat_ids(token: str):
    try:
        r = requests.get(f"https://api.telegram.org/bot{token}/getUpdates", timeout=(4, 10))
        r.raise_for_status()
        data = r.json()
        if not data.get("ok"):
            raise RuntimeError("Telegram token 無效或無法取得更新")
        ids = {}
        for u in data.get("result", []):
            c = (u.get("message") or u.get("edited_message") or u.get("channel_post") or {}).get("chat")
            if c:
                ids[str(c["id"])] = c.get("title") or c.get("username") or c.get("first_name") or ""
        return ids
    except (requests.RequestException, ValueError):
        raise RuntimeError("查詢 chat_id 失敗；先對自己的 Bot 傳 /start，再重試") from None


def ai_analysis(api_key: str, model: str, snapshot: dict, plan: dict, news: list[dict]):
    if not api_key:
        return None
    try:
        from openai import OpenAI
        # Only supplied numerical data, no fabricated live market or unverified news claims.
        dataset = {
            "snapshot": snapshot,
            "rule_based_plan": plan,
            "news_headlines_with_source_and_time": news[:4],
        }
        client = OpenAI(api_key=api_key, timeout=18.0, max_retries=0)
        response = client.responses.create(
            model=model,
            instructions=(
                "你是保守、嚴謹的加密市場播報員。輸出繁體中文，120字內。"
                "只能根據 JSON 提供的數值/時間分析，不得聲稱掌握即時其他價格或新聞，"
                "不得憑空編造 MACD、OI、資金流、消息或勝率。"
                "不要把新聞標題當作已證實的漲跌原因。"
                "明確指出現階段偏多/偏空/觀望，簡述證據與無效條件。"
                "規則式計畫若為觀望，不要命令用戶開倉。"
                "不保證收益。"
            ),
            input=json.dumps(dataset, ensure_ascii=False, separators=(",", ":")),
            max_output_tokens=250,
            store=False,
        )
        return (response.output_text or "").strip() or None
    except Exception as exc:
        # Avoid key leakage by only reporting exception CLASS, not str(exc).
        print(f"[AI] 無法分析: {type(exc).__name__}", flush=True)
        return None
