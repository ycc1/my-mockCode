# ETH/USDT 每分鐘 AI 行情播報（Telegram）

這是 **不下單、只讀公開行情** 的 Python 程式。你的電腦只要持續開機並執行程式，每分鐘便會抓取 **Binance USDⓈ-M `ETHUSDT` 永續合約** 的行情，計算技術指標與合約數據，透過 Telegram 發給你。

## 功能

- 每 60 秒左右檢查一次最新合約成交價，使用交易所回傳的報價時間；資料過期則不播報虛假的「最新價」。
- 1M、15M、1H、4H **已收盤** K 線；EMA20/EMA50、RSI14、MACD(12,26,9)、ATR14。
- Binance 期貨 OI、近 1 小時 OI 變動、資金費率、全站多空**帳戶數**比例（不是持倉金額）；期貨 1H/4H 主動買賣差與現貨 1H 主動買賣差（使用 K 線 taker buy quote 量估算）。
- 支撐/阻力：最近 20 根已收的 15M K 線高低點（排除最新一根）。
- 1 分鐘價格急變、5 分鐘急變、1 分鐘成交量異常、穿越支撐/壓力、資金費率異常、方向變更。
- 可選 GPT 分析：只根據抓到的 JSON 與有發布時間的新聞標題分析；GPT 失敗時顯示規則式報告，不會用舊 AI 分析冒充新訊號。
- 可選 Google News RSS 新聞，失敗時跳過，不杜撰新聞。
- `NOTIFY_MODE=every_minute`：每次檢查成功就 Telegram 播報。`alerts_only`：重要變化才通知，並每小時至少一次心跳。
- 不使用任何可以交易的 Binance API 金鑰，不執行訂單。

> 限制：每分鐘「檢查」而非毫秒級行情；API 抖動及電腦休眠會延後。動態 K 線使用已完成的 K 線以避免未收盤訊號反覆變化。OI 的名目 USDT 為 OI ETH 數量乘上標記價格的近似值；新聞並不能證明價格漲跌的因果。若你的網路或地區無法存取 Binance，監控會報錯，不會假設即時價格。

## 1. 取得 Telegram 機器人 Token 與 Chat ID

1. 手機開啟 Telegram，搜尋官方 [@BotFather](https://t.me/BotFather)，傳 `/newbot`，依提示建立 Bot。
2. 複製 BotFather 給你的 **Token**（不要把 Token 傳給其他人）。
3. 搜尋你剛建立的 Bot，傳 `/start`，確保已有一則對話。
4. 在專案根目錄將 `.env.example` 複製成 `.env`，先填入 `TELEGRAM_BOT_TOKEN=`。
5. 安裝程式後執行：`python monitor.py --chat-id`（Windows 可用 `.venv\Scripts\python monitor.py --chat-id`）。輸出中找到你的 `chat_id`，填入 `.env` 的 `TELEGRAM_CHAT_ID=`。

若 Telegram Bot 先前設了 webhook，`getUpdates` 可能無法使用；請先在原 Bot 管理端停用 webhook。不要用含 Token 的連結貼到公開聊天。

## 2. 取得 OpenAI API Key

1. 前往 https://platform.openai.com/api-keys 建立 API key。
2. 複製到 `.env` 的 `OPENAI_API_KEY=`。
3. `OPENAI_MODEL` 預設為 `gpt-5.4-mini`，如果你的 API 專案不支援，改成你帳戶可用的模型。
4. **OpenAI API 與 ChatGPT Plus 訂閱是分開計費的**；每分鐘呼叫一次約為每天 1,440 次。可把 `AI_INTERVAL_MINUTES=5` 減少費用，但行情仍每分鐘檢查。
5. 未填 API key 也能跑，會使用規則式技術分析，但不會提供 GPT 的文字判斷。

`.env` 包含機密，已列入 `.gitignore` 與 `.dockerignore`。**不要把 Bot Token 或 OpenAI API key 給我，也不要提交 GitHub。**

## 3. Windows 最簡單部署

需要 Windows 10/11、Python 3.11 以上（安裝時建議勾選 PATH）、網路連線及 Telegram。

1. 將 ZIP 解壓縮，例如 `C:\eth_telegram_monitor`。
2. 將 `.env.example` 複製為 `.env`，依前兩節設定 Telegram 與（可選）OpenAI key。
3. **雙擊 `start_windows.bat`**；第一次會建立 `.venv` 並安裝依賴，之後直接執行。
4. 看到 `啟動 ETH/USDT 監控` 後，Telegram 每分鐘會收到行情。若沒收到，查看視窗內的錯誤。
5. 按 `Ctrl+C` 停止。關閉視窗、電腦休眠或關機，都會停止播報。

手動命令（PowerShell，目錄需在專案內）：

```powershell
py -3 -m venv .venv
.venv\Scripts\python.exe -m pip install -r requirements.txt
Copy-Item .env.example .env  # 只需一次，再編輯 .env
.venv\Scripts\python.exe monitor.py --chat-id
.venv\Scripts\python.exe monitor.py --once --dry-run
.venv\Scripts\python.exe monitor.py
```

建議先執行 `--once --dry-run`，確認 Binance 公開行情在你的電腦可以正常使用；此模式仍可能使用 OpenAI API 並產生費用，但不會發送 Telegram。

### Windows 開機自動啟動（可選）

1. 開啟「工作排程器」，建立工作，觸發器選 **使用者登入時**。
2. 「動作」選 **啟動程式**，程式指定 `C:\eth_telegram_monitor\start_windows.bat`（依你的實際路徑調整）。
3. 「起始位置」指定該專案資料夾。請確保電腦不要休眠；否則不會持續監控。

## 4. Docker 24 小時持續運作（電腦或 VPS）

先安裝 Docker Desktop / Docker Engine。設定 `.env` 後執行：

```bash
docker compose up -d --build
docker compose logs -f eth-monitor
```

停止：`docker compose down`。Docker `restart: unless-stopped` 可以在主機或 Docker 重新啟動後自動恢復；**主機關機或休眠仍然無法監控**。

Mac / Linux 亦可用 `./start_mac_linux.sh` 執行 Python 虛擬環境版本。

## 5. 主要設定（編輯 `.env`）

| 設定 | 預設 | 用途 |
|---|---|---|
| `NOTIFY_MODE` | `every_minute` | 每分鐘發一則，改 `alerts_only` 只發警示及心跳 |
| `AI_INTERVAL_MINUTES` | `1` | GPT 每幾分鐘分析一次；行情仍每分鐘查詢 |
| `NEWS_INTERVAL_MINUTES` | `30` | 更新新聞標題的間隔 |
| `PRICE_ALERT_1M_PERCENT` | `0.35` | 與上次檢查相比變動百分比門檻 |
| `PRICE_ALERT_5M_PERCENT` | `1` | 與約 5 分鐘前 K 線收盤價比較的門檻 |
| `VOLUME_ALERT_MULTIPLE` | `2.5` | 上一根已收 1M K 線量能高於前 20 根均量倍數 |
| `FUNDING_ALERT_ABS_PERCENT` | `0.03` | 已發布資金費率絕對值門檻（百分點） |
| `INCLUDE_SPOT` | `true` | 是否額外抓現貨主動買賣差 |
| `MAX_QUOTE_AGE_SECONDS` | `90` | 拒絕時間戳過期的成交價 |

**特別注意：** 「方向」是 EMA、RSI、MACD 的可解釋規則分數，**不是**有回測支持的勝率。進出場價是 ATR 風險參考，進場前仍須確認條件，不能當作必然成交價格。

## 6. 驗證與測試

```bash
python -m unittest discover -s tests -v
python monitor.py --once --dry-run
python monitor.py --check-config
```

`--once --dry-run` 會訪問 Binance 公共 API，需要連線，並在終端機印出一份行情；真正 Telegram 通知請用 `python monitor.py` 或 `python monitor.py --once`。

## 7. 可以交給 Codex 擴充

把此資料夾在 Codex Desktop / VS Code 的 Codex 中開啟，依 `CODEX_PROMPT.md` 的內容提出修改需求。Codex 可以在它有權存取的本機工作區編輯檔案；此聊天無法因為你的電腦開著就自動取得本機檔案權限。

## 8. 官方文件

- Binance USDⓈ-M Futures Market Data: https://developers.binance.com/docs/derivatives/usds-margined-futures/market-data
- OpenAI Python SDK: https://github.com/openai/openai-python
- Telegram Bot API: https://core.telegram.org/bots/api

## 9. 目前未實作

- 自動下單、讀取用戶私有部位、收益率保證、逐秒 WebSocket 訂單簿、爆倉總量與新聞全文驗證；報表不會假造這些數據。
- 如果要「每分鐘由手機直接語音念出」，Telegram 文字通知之外還需要裝置端文字轉語音或額外 TTS 音檔流程。
