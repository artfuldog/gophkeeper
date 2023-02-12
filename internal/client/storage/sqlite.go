package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"

	_ "github.com/mattn/go-sqlite3" // sqlite
)

// SQLite represents sqlite3 implementation of DB.
type SQLite struct {
	db       *sql.DB
	filepath string
	dbFile   *os.File
	stmts    map[string]*sql.Stmt
}

var _ S = (*SQLite)(nil)

// newSQLite creates new instance of sqlite3 storage.
func newSQLite(username string, dir string) *SQLite {
	return &SQLite{
		filepath: fmt.Sprintf("%s.%s.db", dir, username),
	}
}

// Connect is used for opening sqlite3 database file, creating db schema if needed,
// preparaing sql statements.
// Close channel shoud be passed for controlling storage disconnect and close processes.
func (s *SQLite) Connect(ctx context.Context, stopCh chan<- struct{}) (err error) {
	s.dbFile, err = os.OpenFile(s.filepath, os.O_CREATE, 0644)
	if err != nil && !errors.Is(err, os.ErrExist) {
		return err
	}

	s.db, err = sql.Open("sqlite3", s.filepath)
	if err != nil {
		return err
	}

	if _, err := s.db.Exec(stmtCreateTables); err != nil {
		return err
	}

	if err := s.prepareStatements(ctx); err != nil {
		return err
	}

	if _, err = s.db.Exec(`INSERT INTO revision (id, revision) VALUES(?,?)`, 0, []byte("")); err != nil {
		return err
	}

	go s.maintainConnectinon(ctx, stopCh)

	return nil
}

// maintainConnectinon is a helper function for closing database connection and file.
// After successfui stopping closes stopCh.
func (s *SQLite) maintainConnectinon(ctx context.Context, stopCh chan<- struct{}) {
	<-ctx.Done()

	for i := range s.stmts {
		s.stmts[i].Close()
	}

	s.db.Close()
	s.dbFile.Close()

	close(stopCh)
}

// Delete deletes db file from disk.
func (s *SQLite) Delete() {
	err := os.Remove(s.filepath)
	if err != nil {
		log.Printf("Failed to delete DB file: %v", err)
	}
}

// GetRevision returns current identificator of revision.
func (s *SQLite) GetRevision(ctx context.Context) ([]byte, error) {
	revision := []byte{}

	row := s.stmts["getRev"].QueryRow()
	if err := row.Scan(&revision); err != nil {
		return nil, err
	}

	return revision, nil
}

// GetRevision save new identificator of revision.
func (s *SQLite) SaveRevision(ctx context.Context, revision []byte) error {
	if _, err := s.stmts["updateRev"].ExecContext(ctx, revision); err != nil {
		return err
	}

	return nil
}

// CreateItems creates new item in database.
func (s *SQLite) CreateItems(ctx context.Context, items Items) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	defer tx.Rollback() //nolint:errcheck

	txStmt := tx.StmtContext(ctx, s.stmts["createItems"])
	defer txStmt.Close()

	for _, item := range items {
		if _, err = txStmt.ExecContext(ctx, item.ID, item.Name, item.Type,
			item.Hash, item.Data); err != nil {
			return err
		}
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}

// GetItem returns item's data form database.
func (s *SQLite) GetItem(ctx context.Context, itemName, itemType string) ([]byte, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}

	defer tx.Rollback() //nolint:errcheck

	txStmt := tx.StmtContext(ctx, s.stmts["getItem"])
	defer txStmt.Close()

	var data []byte

	row := txStmt.QueryRowContext(ctx, itemName, itemType)
	if err := row.Scan(&data); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return data, err
}

// GetItemsList returns short representation of all existing items.
//
// GetItemsList returns only descrition fields, without secret's data.
func (s *SQLite) GetItemsList(ctx context.Context) (Items, error) {
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return nil, err
	}

	defer tx.Rollback() //nolint:errcheck

	txStmt := tx.StmtContext(ctx, s.stmts["getItemsList"])
	defer txStmt.Close()

	rows, err := txStmt.QueryContext(ctx)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var items Items

	for rows.Next() {
		var (
			id    int64
			name  string
			iType string
			hash  []byte
		)

		if err := rows.Scan(&id, &name, &iType, &hash); err != nil {
			return nil, err
		}

		items = append(items, &Item{
			ID:   id,
			Name: name,
			Type: iType,
			Hash: hash,
		})
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return items, nil
}

// UpdateItems updates current items with provided data.
//
// UpdateItems doesn't control update errors in case of missed id, e.g. update
// of unexisted item will not return error.
func (s *SQLite) UpdateItems(ctx context.Context, items Items) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	defer tx.Rollback() //nolint:errcheck

	txStmt := tx.StmtContext(ctx, s.stmts["updateItems"])
	defer txStmt.Close()

	for _, item := range items {
		if _, err = txStmt.ExecContext(ctx, item.Name, item.Hash,
			item.Data, item.ID); err != nil {
			return err
		}
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}

// DeleteItems deletes existed items by its IDs.
//
// DeleteItems doesn't control delete errors in case of missed id, e.g. delete
// of unexisted item will not return error.
func (s *SQLite) DeleteItems(ctx context.Context, ids []int64) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	defer tx.Rollback() //nolint:errcheck

	txStmt := tx.StmtContext(ctx, s.stmts["deleteItems"])
	defer txStmt.Close()

	for _, id := range ids {
		if _, err = txStmt.ExecContext(ctx, id); err != nil {
			return err
		}
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}

// ClearItems deletes all items for database.
func (s *SQLite) ClearItems(ctx context.Context) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	defer tx.Rollback() //nolint:errcheck

	if _, err = tx.ExecContext(ctx, `DELETE FROM vault;`); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}
