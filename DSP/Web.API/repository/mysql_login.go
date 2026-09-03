package repository

import (
	"context"
	"database/sql"
	"fmt"
	"regexp"
)

var safeIdentifier = regexp.MustCompile(`^[a-zA-Z0-9_]+$`)

type MySQLLoginRepository struct {
	db        *sql.DB
	tableName string
}

func NewMySQLLoginRepository(db *sql.DB, tableName string) (*MySQLLoginRepository, error) {
	if !safeIdentifier.MatchString(tableName) {
		return nil, fmt.Errorf("invalid login table name %q", tableName)
	}
	return &MySQLLoginRepository{db: db, tableName: tableName}, nil
}

func (r *MySQLLoginRepository) Validate(username, password string) (bool, error) {
	query := fmt.Sprintf("SELECT 1 FROM `%s` WHERE username = ? AND password = ? LIMIT 1", r.tableName)
	var exists int
	err := r.db.QueryRowContext(context.Background(), query, username, password).Scan(&exists)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}
