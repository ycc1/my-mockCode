"""Optional Google News RSS headlines: context, not proven causes of market moves."""
from __future__ import annotations
from urllib.parse import urlencode
from datetime import datetime, timezone
from email.utils import parsedate_to_datetime
import xml.etree.ElementTree as ET
import requests


def fetch_news() -> list[dict]:
    q = urlencode({"q": 'Ethereum ETH crypto when:1d', "hl": "en-US", "gl": "US", "ceid": "US:en"})
    url = "https://news.google.com/rss/search?" + q
    try:
        response = requests.get(url, timeout=(4, 7), headers={"User-Agent": "ETHMonitor/1.0"})
        response.raise_for_status()
        root = ET.fromstring(response.content)
        headlines = []
        for item in root.findall("./channel/item"):
            title = (item.findtext("title") or "").strip()
            link = (item.findtext("link") or "").strip()
            pub = item.findtext("pubDate")
            if not title or not link or not pub:
                continue
            try:
                when = parsedate_to_datetime(pub).astimezone(timezone.utc)
            except (TypeError, ValueError):
                continue
            age = (datetime.now(timezone.utc) - when).total_seconds()
            if not 0 <= age <= 86400 + 120:
                continue
            if not ("ethereum" in title.lower() or "eth" in title.lower()):
                continue
            headlines.append({"title": title[:180], "published_utc": when.isoformat(), "url": link})
            if len(headlines) >= 4:
                break
        return headlines
    except (requests.RequestException, ET.ParseError):
        return []
