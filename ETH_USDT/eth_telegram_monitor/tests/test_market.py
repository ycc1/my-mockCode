import time
import unittest
from market import parse_candles, MarketError


class DataTests(unittest.TestCase):
    def test_excludes_unfinished_candle(self):
        now = int(time.time() * 1000)
        now = (now // 60_000) * 60_000 + 20_000
        raw = []
        for j in range(62):
            start = now // 60_000 * 60_000 - (61-j) * 60_000
            end = start + 59_999
            raw.append([start, "100", "110", "90", "101", "10", end,
                        "1000", 40, "5", "500", "0"])
        candles = parse_candles(raw, now, 60_000)
        self.assertEqual(len(candles), 61)
        self.assertLess(candles[-1]["end_ms"], now)

    def test_stale_fails_closed(self):
        now = int(time.time() * 1000)
        old = now - 3_600_000
        raw = [[old+i*60_000, "100", "110", "90", "101", "10",
                old+i*60_000+59_999, "1000", 40, "5", "500", "0"] for i in range(70)]
        with self.assertRaises(MarketError):
            parse_candles(raw, now + 5_000_000, 60_000)


if __name__ == "__main__":
    unittest.main()
