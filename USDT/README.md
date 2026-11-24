# ETH/USDT Go 行情中心

將原 Python `ETH_USDT/eth_telegram_monitor` 的 Binance 公開行情、已收盤 K 線技術指標與規則方向轉為 Go，並提供本機前台展示頁。

## 功能

- 讀取 Binance USDⓈ-M `ETHUSDT` 永續合約公開行情。
- 1M、15M、1H、4H 已收盤 K 線：EMA20、EMA50、RSI14、MACD、ATR14。
- OI、1H OI 變化、資金費率、多空帳戶比例。
- 期貨 1H/4H 與現貨 1H 主動買賣差。
- 可解釋規則方向、支撐阻力與風險參考。
- 缺少選配資料時顯示「未取得」；核心報價逾時時拒絕顯示舊資料。
- 響應式前台每 15 秒更新。

## 啟動

需要 Go 1.22+。

```powershell
cd D:\PM\外案\外案mock\my-mockCode\USDT
Copy-Item .env.example .env # 可選；Go 版目前使用環境變數
$env:SYMBOL="ETHUSDT"
go run .
```

開啟 `http://localhost:8080`。

## 設定

| 環境變數 | 預設 | 說明 |
|---|---:|---|
| `ADDR` | `:8080` | HTTP 監聽位址 |
| `SYMBOL` | `ETHUSDT` | Binance USDⓈ-M 交易對 |
| `MAX_QUOTE_AGE_SECONDS` | `90` | 核心報價最大允許延遲 |
| `INCLUDE_SPOT` | `true` | 是否抓現貨主動買賣差 |
| `CACHE_SECONDS` | `15` | 後端資料快取秒數 |

## 驗證

```powershell
go test ./...
go build ./...
```

## API

`GET /api/snapshot` 回傳公開行情與規則分析 JSON。若核心行情失敗則回傳 `502`，不回傳舊快取。

## 安全與限制

此版本不連接 Telegram/OpenAI、不需要 Binance API 金鑰，也不支援下單。技術分數不是勝率、收益保證或投資建議。
