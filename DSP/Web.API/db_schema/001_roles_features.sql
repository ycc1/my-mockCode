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
    ('FEAT_OFFER', 'offer', 'Offer management', 'Create and manage offers'),
    ('FEAT_ROLE_MANAGE', 'role.manage', 'Role management', 'Manage roles and feature assignments'),
    ('FEAT_FEATURE_MANAGE', 'feature.manage', 'Feature management', 'Manage available features')
ON DUPLICATE KEY UPDATE name = VALUES(name), description = VALUES(description);

INSERT INTO role_features (role_id, feature_id)
SELECT 'ROLE_ADMIN', feature_id FROM features
WHERE code IN ('offer', 'role.manage', 'feature.manage')
ON DUPLICATE KEY UPDATE role_id = VALUES(role_id);
