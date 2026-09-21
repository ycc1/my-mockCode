CREATE TABLE IF NOT EXISTS ads_partners (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    ads_partner_id VARCHAR(64) NOT NULL,
    name VARCHAR(150) NOT NULL,
    ads_merchant_id VARCHAR(20) NOT NULL,
    api_key VARCHAR(128) NOT NULL,
    security_type VARCHAR(20) NOT NULL,
    create_time TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    update_time TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    create_by VARCHAR(100) NOT NULL,
    update_by VARCHAR(100) NOT NULL,
    PRIMARY KEY (id),
    UNIQUE KEY uq_ads_partners_id (ads_partner_id),
    UNIQUE KEY uq_ads_partners_merchant_id (ads_merchant_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
