package sqlx

import "database/sql"

type Transaction struct {
	tx *sql.Tx
}

func (t Transaction) Select(dest any, query string, args ...any) error {
	return Select(t.tx, dest, query, args...)
}

func (t Transaction) Exec(query string, args ...any) (sql.Result, error) {
	return Exec(t.tx, query, args...)
}

func (t Transaction) Commit() error {
	return t.tx.Commit()
}

func (t Transaction) Rollback() error {
	return t.tx.Rollback()
}
