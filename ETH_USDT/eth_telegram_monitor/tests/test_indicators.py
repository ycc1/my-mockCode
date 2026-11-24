import unittest
import math
from indicators import ema, rsi, macd, atr, pct_change, sum_net_taker_quote, compute_indicators
from strategy import judge, alerts


class IndicatorTests(unittest.TestCase):
    def test_ema_constant(self):
        self.assertAlmostEqual(ema([10.0] * 60, 20)[-1], 10.0)

    def test_rsi_constant_not_nan(self):
        self.assertEqual(rsi([100.0] * 60), 50)

    def test_rsi_trending(self):
        self.assertEqual(rsi([float(i) for i in range(80)]), 100)
        self.assertAlmostEqual(rsi([float(80-i) for i in range(80)]), 0)

    def test_macd_on_constant(self):
        v = macd([5.0] * 90)
        self.assertAlmostEqual(v["hist"], 0)

    def test_atr(self):
        cs = [{"high": 102., "low": 98., "close": 100.} for _ in range(80)]
        self.assertAlmostEqual(atr(cs), 4.0)

    def test_taker_net_quote(self):
        cs = [{"quote_volume": 1000, "taker_buy_quote": 650} for _ in range(60)]
        self.assertEqual(sum_net_taker_quote(cs, 60), 18_000)

    def test_pct_change(self):
        self.assertAlmostEqual(pct_change(105,100), 5)

    def test_strategy_bearish(self):
        data = [dict(high=p+2, low=p-2, close=p, quote_volume=1000, taker_buy_quote=400)
                for p in (1800 - i*0.15 + math.sin(i/2)*5 for i in range(150))]
        ind = compute_indicators(data)
        s = {"price": data[-1]["close"], "indicators": {"1h": ind, "15m": ind}}
        got = judge(s)
        self.assertEqual(got["direction"], "偏空")
        self.assertGreater(got["stop_reference"], got["entry_reference"])
        self.assertLess(got["target_reference"], got["entry_reference"])

    def test_alerts(self):
        s = {"price": 102., "change_5m_pct": 2., "volume_ratio_1m": 3.0,
             "funding_rate_percent": .001}
        prev = {"last_price": 100., "support": 99., "resistance": 101.}
        t = {"price_pct_1m": .3, "price_pct_5m": 1., "volume_multiple": 2.5,
             "funding_abs_pct": .03}
        self.assertGreaterEqual(len(alerts(s, prev, t)), 3)


if __name__ == "__main__":
    unittest.main()
