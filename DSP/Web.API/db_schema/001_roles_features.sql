-- Role and feature authorization schema.
-- Run after the users table exists.

CREATE TABLE IF NOT EXISTS roles (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    role_id VARCHAR(64) NOT NULL,
    name VARCHAR(100) NOT NULL,
    description VARCHAR(255) NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    UNIQUE KEY uq_roles_role_id (role_id),
    UNIQUE KEY uq_roles_name (name)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS features (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    feature_id VARCHAR(64) NOT NULL,
    code VARCHAR(100) NOT NULL,
    name VARCHAR(100) NOT NULL,
    description VARCHAR(255) NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    UNIQUE KEY uq_features_feature_id (feature_id),
    UNIQUE KEY uq_features_code (code)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS role_features (
    role_id VARCHAR(64) NOT NULL,
    feature_id VARCHAR(64) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (role_id, feature_id),
    CONSTRAINT fk_role_features_role FOREIGN KEY (role_id) REFERENCES roles(role_id) ON DELETE CASCADE,
    CONSTRAINT fk_role_features_feature FOREIGN KEY (feature_id) REFERENCES features(feature_id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

ALTER TABLE users ADD COLUMN role_id VARCHAR(64) NULL;
ALTER TABLE users ADD CONSTRAINT fk_users_role FOREIGN KEY (role_id) REFERENCES roles(role_id) ON DELETE SET NULL;

INSERT INTO roles (role_id, name, description)
VALUES ('ROLE_ADMIN', 'Administrator', 'Full API access')
ON DUPLICATE KEY UPDATE name = VALUES(name);

INSERT INTO features (feature_id, code, name, description) VALUES
    ('FEAT_OFFER_CREATE', 'offer.create', 'Offer create', 'Create offers'),
    ('FEAT_OFFER_READ', 'offer.read', 'Offer read', 'Read offers'),
    ('FEAT_OFFER_UPDATE', 'offer.update', 'Offer update', 'Update offers'),
    ('FEAT_OFFER_DELETE', 'offer.delete', 'Offer delete', 'Delete offers'),
    ('FEAT_ROLE_CREATE', 'role.create', 'Role create', 'Create roles'),
    ('FEAT_ROLE_READ', 'role.read', 'Role read', 'Read roles'),
    ('FEAT_ROLE_UPDATE', 'role.update', 'Role update', 'Update roles and assignments'),
    ('FEAT_ROLE_DELETE', 'role.delete', 'Role delete', 'Delete roles'),
    ('FEAT_FEATURE_CREATE', 'feature.create', 'Feature create', 'Create features'),
    ('FEAT_FEATURE_READ', 'feature.read', 'Feature read', 'Read features'),
    ('FEAT_FEATURE_UPDATE', 'feature.update', 'Feature update', 'Update features'),
    ('FEAT_FEATURE_DELETE', 'feature.delete', 'Feature delete', 'Delete features'),
    ('FEAT_CHANNEL_PARTNER_CREATE', 'channel_partner.create', 'Channel partner create', 'Create channel partners'),
    ('FEAT_CHANNEL_PARTNER_READ', 'channel_partner.read', 'Channel partner read', 'Read channel partners'),
    ('FEAT_CHANNEL_PARTNER_UPDATE', 'channel_partner.update', 'Channel partner update', 'Update channel partners'),
    ('FEAT_CHANNEL_PARTNER_DELETE', 'channel_partner.delete', 'Channel partner delete', 'Delete channel partners')
ON DUPLICATE KEY UPDATE name = VALUES(name), description = VALUES(description);

INSERT INTO role_features (role_id, feature_id)
SELECT 'ROLE_ADMIN', feature_id FROM features
WHERE code IN ('offer.create', 'offer.read', 'offer.update', 'offer.delete', 'role.create', 'role.read', 'role.update', 'role.delete', 'feature.create', 'feature.read', 'feature.update', 'feature.delete', 'channel_partner.create', 'channel_partner.read', 'channel_partner.update', 'channel_partner.delete')
ON DUPLICATE KEY UPDATE role_id = VALUES(role_id);
