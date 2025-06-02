package sqlx

import (
	"database/sql"
)

type RowMapper interface {
	MapRow(map[string]any) error
}

type DB struct {
	DB *sql.DB
}

func (db *DB) Select(dest any, query string, args ...any) error {
	return Select(db.DB, dest, query, args...)
}

func (db *DB) Exec(query string, args ...any) (sql.Result, error) {
	return Exec(db.DB, query, args...)
}

func (db *DB) Close() error {
	return db.DB.Close()
}

func (db *DB) Ping() error {
	return db.DB.Ping()
}

func (db *DB) Begin() (*Transaction, error) {
	sqlTx, err := db.DB.Begin()
	if err != nil {
		return nil, err
	}
	return &Transaction{tx: sqlTx}, nil
}

func Open(driverName, dataSourceName string) (*DB, error) {
	_db, err := sql.Open(driverName, dataSourceName)
	if err != nil {
		return nil, err
	}
	return &DB{DB: _db}, err
}
