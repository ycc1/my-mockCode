#!/usr/bin/env python3
"""ETH/USDT minute scanner. Read-only: no keys or endpoints for placing trades."""
from __future__ import annotations
import argparse
import json
import os
import sys
import time
from datetime import datetime
from pathlib import Path
from zoneinfo import ZoneInfo
from dotenv import load_dotenv

from market import snapshot, MarketError
from strategy import judge, alerts
from notify import telegram_send, discover_chat_ids, ai_analysis, build_base_message
from news import fetch_news

ROOT = Path(__file__).resolve().parent
load_dotenv(ROOT / ".env")
TZ = ZoneInfo("Asia/Tokyo")


def fenv(key, default):
    try:
        return float(os.getenv(key, default))
    except ValueError:
        raise SystemExit(f"設定錯誤：{key} 必須是數字")


def ienv(key, default):
    try:
        return int(os.getenv(key, default))
    except ValueError:
        raise SystemExit(f"設定錯誤：{key} 必須是整數")


def config():
    return {
        "symbol": os.getenv("SYMBOL", "ETHUSDT").strip().upper(),
        "telegram_token": os.getenv("TELEGRAM_BOT_TOKEN", "").strip(),
        "telegram_chat_id": os.getenv("TELEGRAM_CHAT_ID", "").strip(),
        "openai_api_key": os.getenv("OPENAI_API_KEY", "").strip(),
        "openai_model": os.getenv("OPENAI_MODEL", "gpt-5.4-mini").strip(),
        "notify_mode": os.getenv("NOTIFY_MODE", "every_minute").strip(),
        "ai_interval": ienv("AI_INTERVAL_MINUTES", 1),
        "news_interval": ienv("NEWS_INTERVAL_MINUTES", 30),
        "max_age": ienv("MAX_QUOTE_AGE_SECONDS", 90),
        "include_spot": os.getenv("INCLUDE_SPOT", "true").lower() in {"1", "true", "yes"},
        "thresholds": {
            "price_pct_1m": fenv("PRICE_ALERT_1M_PERCENT", 0.35),
            "price_pct_5m": fenv("PRICE_ALERT_5M_PERCENT", 1.0),
            "volume_multiple": fenv("VOLUME_ALERT_MULTIPLE", 2.5),
            "funding_abs_pct": fenv("FUNDING_ALERT_ABS_PERCENT", 0.03),
        },
        "state_path": ROOT / "data" / "state.json",
        "alert_cooldown": ienv("ALERT_COOLDOWN_MINUTES", 10),
    }


def load_state(path):
    try:
        data = json.loads(path.read_text(encoding="utf-8"))
        return data if isinstance(data, dict) else {}
    except (OSError, json.JSONDecodeError):
        return {}


def save_state(path, state):
    path.parent.mkdir(parents=True, exist_ok=True)
    temp = path.with_suffix(".tmp")
    temp.write_text(json.dumps(state, ensure_ascii=False, indent=2), encoding="utf-8")
    temp.replace(path)


def run_one(cfg, state, dry_run):
    now = time.time()
    snap = snapshot(cfg["symbol"], cfg["max_age"], cfg["include_spot"])
    plan = judge(snap)
    issues = alerts(snap, state, cfg["thresholds"])
    if state.get("last_signal") and state["last_signal"] != plan["direction"]:
        issues.insert(0, f"交易方向由 {state['last_signal']} 變為 {plan['direction']}")
    news = state.get("news", [])
    if now - state.get("news_last_fetch", 0) >= cfg["news_interval"] * 60:
        news = fetch_news()
        state["news_last_fetch"] = now
        state["news"] = news
    # Run AI when requested, or when an alert occurs. Even if AI fails, send deterministic data.
    ai_text = None
    last_ai = state.get("last_ai_at", 0)
    if cfg["openai_api_key"] and (now - last_ai >= cfg["ai_interval"] * 60 - 5 or issues):
        ai_text = ai_analysis(cfg["openai_api_key"], cfg["openai_model"], snap, plan, news)
        if ai_text:
            state["last_ai_at"] = now
            state["last_ai_text"] = ai_text
    # The cached AI analysis is NOT reused; if AI skipped, don't misrepresent it as current.
    body = build_base_message(snap, plan, issues, ai_text, news)
    mode = cfg["notify_mode"]
    cooldown = cfg["alert_cooldown"] * 60
    if mode == "every_minute":
        should_send = True
    else:
        # First report; thereafter alerts with cooldown + an hourly health heartbeat.
        should_send = ("last_sent" not in state or
                       (bool(issues) and now - state.get("last_alert_sent", 0) >= cooldown) or
                       now - state.get("last_sent", 0) >= 3600)
    if should_send:
        if dry_run:
            print(body + "\n", flush=True)
        else:
            telegram_send(cfg["telegram_token"], cfg["telegram_chat_id"], body)
        state["last_sent"] = now
        if issues:
            state["last_alert_sent"] = now
    local = datetime.now(TZ).strftime("%Y-%m-%d %H:%M:%S")
    print(f"[{local} JST] {snap['symbol']} {snap['price']:,.2f}｜{plan['direction']}｜警示{len(issues)}｜送出{should_send}", flush=True)
    state.update({"last_price": snap["price"], "support": snap["support"],
                  "resistance": snap["resistance"], "last_signal": plan["direction"],
                  "last_quote_time_ms": snap["quote_time_ms"]})
    return state


def main():
    ap = argparse.ArgumentParser(description="每分鐘 ETH 永續行情與 GPT Telegram 播報")
    ap.add_argument("--once", action="store_true", help="只執行一次")
    ap.add_argument("--dry-run", action="store_true", help="不發 Telegram，但會抓行情並可呼叫 OpenAI")
    ap.add_argument("--chat-id", action="store_true", help="取得曾向 Bot 傳訊息的 chat_id")
    ap.add_argument("--check-config", action="store_true", help="只檢查必要設定是否齊全")
    args = ap.parse_args()
    cfg = config()
    if args.chat_id:
        if not cfg["telegram_token"]:
            raise SystemExit("請先在 .env 填寫 TELEGRAM_BOT_TOKEN")
        ids = discover_chat_ids(cfg["telegram_token"])
        if ids:
            print("可使用的 chat_id：")
            for k, v in ids.items():
                print(f"  {k}  {v}")
        else:
            print("找不到訊息。請先到 Telegram 向你的 Bot 傳送 /start，再重試。")
        return
    if cfg["notify_mode"] not in {"every_minute", "alerts_only"}:
        raise SystemExit("NOTIFY_MODE 僅接受 every_minute 或 alerts_only")
    if cfg["ai_interval"] < 1 or cfg["news_interval"] < 1:
        raise SystemExit("AI_INTERVAL_MINUTES / NEWS_INTERVAL_MINUTES 必須 >= 1")
    if args.check_config:
        print("設定檢查成功：", "交易對", cfg["symbol"], "通知", cfg["notify_mode"],
              "GPT", "已啟用" if cfg["openai_api_key"] else "未啟用（使用規則式分析）")
        return
    if not args.dry_run and (not cfg["telegram_token"] or not cfg["telegram_chat_id"]):
        raise SystemExit("請先在 .env 填寫 TELEGRAM_BOT_TOKEN 和 TELEGRAM_CHAT_ID，或使用 --dry-run")
    state = load_state(cfg["state_path"])
    print("啟動 ETH/USDT 監控；Ctrl+C 可停止。此程式不會提交交易。", flush=True)
    while True:
        try:
            state = run_one(cfg, state, args.dry_run)
            state.pop("last_error_sent", None)
        except (MarketError, ValueError, KeyError, IndexError, TypeError) as err:
            print(f"[行情錯誤] {type(err).__name__}: {err}", flush=True)
            # Failure must NOT be reported as an up-to-date market quote.
            now = time.time()
            if (not args.dry_run and now - state.get("last_error_sent", 0) >= 1800):
                try:
                    telegram_send(cfg["telegram_token"], cfg["telegram_chat_id"],
                                  "⚠️ ETH 監控：行情資料取得失敗或過期。暫停本次交易訊號；請勿沿用舊報價。")
                    state["last_error_sent"] = now
                except RuntimeError:
                    pass
        except RuntimeError as err:
            print(f"[通知錯誤] {err}", flush=True)
        finally:
            save_state(cfg["state_path"], state)
        if args.once:
            break
        # Restart next wall-clock minute, even if the last task took >60 seconds.
        time.sleep(max(1.0, 60 - (time.time() % 60)))


if __name__ == "__main__":
    try:
        main()
    except KeyboardInterrupt:
        print("使用者結束監控。")
        sys.exit(0)
