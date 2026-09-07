package repository

import (
	"context"
	"database/sql"
)

type MySQLAuthorizationRepository struct{ db *sql.DB }

func NewMySQLAuthorizationRepository(db *sql.DB) *MySQLAuthorizationRepository {
	return &MySQLAuthorizationRepository{db: db}
}

func (r *MySQLAuthorizationRepository) HasFeature(username, featureCode string) (bool, error) {
	const query = `
SELECT 1
FROM users u
JOIN role_features rf ON rf.role_id = u.role_id
JOIN features f ON f.feature_id = rf.feature_id
WHERE u.username = ? AND f.code = ?
LIMIT 1`
	var exists int
	err := r.db.QueryRowContext(context.Background(), query, username, featureCode).Scan(&exists)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}
