package model

// ChannelPartnerOptions defines the dropdown choices used by the API and validation.
func ChannelPartnerOptions() map[string][]string {
	return map[string][]string{
		"security_type":     {"MD5", "SHA312", "AES", "DES"},
		"partner_type":      {"外部", "内部", "其他"},
		"service_category":  {"DSP", "大媒体采买", "网红/达人推广", "SEO", "SEM", "ASO", "社交媒体", "电子邮件营销", "短信营销", "信息流/原生广告", "其他"},
		"traffic_model":     {"cpc", "cpm", "cpa", "cpl", "收入分成", "其他"},
		"billing_model":     {"cpc", "cpm", "cpa", "cpl", "收入分成", "其他"},
		"primary_channel":   {"Google", "TikTok", "Meta", "X", "YouTube", "原生广告", "程序化广告", "seo", "其他"},
		"secondary_channel": {"Google", "TikTok", "Meta", "X", "YouTube", "原生广告", "程序化广告", "seo", "其他"},
		"market_contract":   {"是", "否", "其他"},
		"delivery_package":  {"苹果", "安卓", "PWA", "白标包体", "其他"},
		"data_system":       {"GA4", "appsflyer", "adjust", "自研系统", "其他"},
	}

}
