const $ = (id) => document.getElementById(id);
let refreshing = false;

function number(value, digits = 2) {
  if (value === null || value === undefined) return "未取得";
  return Number(value).toLocaleString("en-US", {
    minimumFractionDigits: digits,
    maximumFractionDigits: digits,
  });
}

function percent(value) {
  if (value === null || value === undefined) return "未取得";
  const output = `${value >= 0 ? "+" : ""}${number(value, 2)}%`;
  return `<span class="${value >= 0 ? "up" : "down"}">${output}</span>`;
}

function usdtMillions(value) {
  if (value === null || value === undefined) return "未取得";
  const output = `${value >= 0 ? "+" : ""}${number(value / 1e6, 2)} 百萬`;
  return `<span class="${value >= 0 ? "up" : "down"}">${output} USDT</span>`;
}

function tokyoTime(ms) {
  if (!ms) return "未取得";
  return new Intl.DateTimeFormat("zh-TW", {
    timeZone: "Asia/Tokyo",
    year: "numeric",
    month: "2-digit",
    day: "2-digit",
    hour: "2-digit",
    minute: "2-digit",
    second: "2-digit",
    hour12: false,
  }).format(ms) + " JST";
}

function setFlow(id, barId, value) {
  $(id).innerHTML = usdtMillions(value);
  if ($(barId)) {
    const scale = Math.min(100, Math.abs(value || 0) / 1e6);
    $(barId).style.opacity = value === null || value === undefined ? ".15" : ".65";
    $(barId).style.transform = `scaleX(${Math.max(.12, scale / 100)})`;
  }
}

function render(data) {
  const snapshot = data.snapshot;
  const plan = data.signal;
  $("price").textContent = number(snapshot.price, 2);
  $("quoteTime").textContent =
    `交易所報價 ${tokyoTime(snapshot.quoteTimeMs)}｜延遲 ${number(snapshot.quoteAgeSeconds, 1)} 秒`;
  $("change24h").innerHTML = percent(snapshot.change24hPercent);
  $("change5m").innerHTML = percent(snapshot.change5mPercent);
  $("volumeRatio").textContent =
    snapshot.volumeRatio1m === null ? "未取得" : `${number(snapshot.volumeRatio1m, 2)}x`;
  $("direction").textContent = plan.direction;
  $("score").textContent = `非勝率分數 ${plan.score > 0 ? "+" : ""}${plan.score}`;
  $("signalCard").className =
    `signal ${plan.direction === "偏多" ? "bullish" : plan.direction === "偏空" ? "bearish" : ""}`;

  const intervals = ["15m", "1h", "4h"];
  const heads = ["週期", "EMA20", "EMA50", "RSI14", "MACD柱"];
  $("indicatorTable").innerHTML =
    heads.map((head) => `<div class="head">${head}</div>`).join("") +
    intervals.map((interval) => {
      const item = snapshot.indicators[interval];
      return `<div>${interval.toUpperCase()}</div>
        <div>${number(item.ema20)}</div>
        <div>${number(item.ema50)}</div>
        <div>${number(item.rsi14, 1)}</div>
        <div>${number(item.macd?.hist, 3)}</div>`;
    }).join("");
  $("support").textContent = `${number(snapshot.support)} USDT`;
  $("resistance").textContent = `${number(snapshot.resistance)} USDT`;
  $("oi").textContent = `${number(snapshot.openInterestEth, 0)} ETH`;
  $("oiChange").innerHTML = percent(snapshot.openInterestChange1hPercent);
  $("longRatio").innerHTML = `${number(snapshot.longAccountsPercent, 1)}%`;
  $("shortRatio").innerHTML = `${number(snapshot.shortAccountsPercent, 1)}%`;
  $("funding").innerHTML = percent(snapshot.fundingRatePercent);
  $("nextFunding").textContent = tokyoTime(snapshot.nextFundingMs);
  setFlow("futures1h", "futures1hBar", snapshot.futuresTakerNet1hUsdt);
  setFlow("futures4h", "futures4hBar", snapshot.futuresTakerNet4hUsdt);
  setFlow("spot1h", "spot1hBar", snapshot.spotTakerNet1hUsdt);
  $("condition").textContent = plan.condition;
  $("entry").textContent = plan.entryReference === null ? "觀望" : number(plan.entryReference, 1);
  $("stop").textContent = plan.stopReference === null ? "—" : number(plan.stopReference, 1);
  $("target").textContent = plan.targetReference === null ? "—" : number(plan.targetReference, 1);
  $("reasons").innerHTML = (plan.reasons || []).map((reason) => `<li>${reason}</li>`).join("");

  const missing = snapshot.missing || [];
  $("notice").hidden = missing.length === 0;
  $("notice").textContent = missing.length ? `未取得：${missing.join("、")}。其他數據仍為本次有效行情。` : "";
  $("liveStatus").className = "live ok";
  $("liveStatus").innerHTML = "<i></i>公開行情已更新";
}

async function refresh() {
  if (refreshing) return;
  refreshing = true;
  $("refreshButton").disabled = true;
  try {
    const response = await fetch("/api/snapshot", { cache: "no-store" });
    const payload = await response.json();
    if (!response.ok) throw new Error(payload.error || `HTTP ${response.status}`);
    render(payload);
  } catch (error) {
    $("liveStatus").className = "live error";
    $("liveStatus").innerHTML = "<i></i>資料暫時無法更新";
    $("notice").hidden = false;
    $("notice").textContent = error.message;
  } finally {
    refreshing = false;
    $("refreshButton").disabled = false;
  }
}

$("refreshButton").addEventListener("click", refresh);
refresh();
setInterval(refresh, 15000);
