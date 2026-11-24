# 給 Codex 的後續需求提示

請在目前的 `eth_telegram_monitor` Python 專案中工作。不要讀取或輸出 `.env` 的任何秘密，不要提交 `.env`，不要要求交易所交易 API key。請保留「只讀行情，不自動下單」原則。先閱讀 README.md、monitor.py、market.py、strategy.py、notify.py 與 tests。

需求：
1. 驗證程式可在 Windows / Python 3.11+ 執行，修復依賴或 API 版本問題。
2. 每分鐘準時檢查 ETHUSDT 永續合約，訊號使用已收盤 K 線，保留每次報價的交易所時間戳與新鮮度檢查。
3. Telegram 每分鐘發完整繁體中文播報，或依 `.env` 的 NOTIFY_MODE 只發異常警示。
4. 指標涵蓋 15M/1H/4H MACD、RSI、量比、主動買賣差、OI、資金費率、多空比例；缺失值應標記「未取得」，不可編造。
5. OpenAI Responses API 支援 GPT 解讀，不要以舊的 AI 文字冒充新分析。呼叫失敗時保留規則式播報。
6. 新增更完整的單元測試，必要時以注入 mock REST 資料驗證；不要用實際交易。
7. README 補充部署細節與費用控制；保留 API key 安全性與風險警告。

最終給出修改清單、測試結果及如何本機啟動。
