import unittest
from notify import build_base_message


class NotifyTests(unittest.TestCase):
    def test_message_without_optional_data(self):
        p = {'ema20': 1750., 'ema50': 1760., 'rsi14': 43.,
             'atr14': 12., 'macd': {'hist': -2., 'dif': -5., 'dea': -3.}}
        snap = dict(symbol='ETHUSDT', price=1743.25, quote_time_ms=1763630000000,
                    quote_age_sec=1., **{'24h': None}, indicators={'15m': p, '1h': p, '4h': p},
                    support=1720., resistance=1770., change_5m_pct=None,
                    volume_ratio_1m=None, futures_taker_net_1h_usdt=None,
                    futures_taker_net_4h_usdt=None, spot_taker_net_1h_usdt=None,
                    oi_usdt_approx=None, oi_change_1h_pct=None,
                    long_accounts_percent=None, short_accounts_percent=None,
                    funding_rate_percent=0.01, missing=['OI'])
        plan = {'direction': '觀望', 'score': 0, 'condition': '等待確認'}
        msg = build_base_message(snap, plan, [], None, None)
        self.assertIn('未取得', msg)
        self.assertIn('觀望', msg)
        self.assertIn('1743.25', msg.replace(',', ''))
        self.assertNotIn('止盈價：0', msg)


if __name__ == '__main__':
    unittest.main()
