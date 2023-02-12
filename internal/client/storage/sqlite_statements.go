package storage

import (
	"context"
	"database/sql"
)

// SQL statements.
const (
	stmtCreateTables = `
	CREATE TABLE IF NOT EXISTS vault (
		id INTEGER NOT NULL PRIMARY KEY,
		name TEXT NOT NULL,
		type TEXT,
		hash BLOB,
		data BLOB,
		UNIQUE (name, type)
	);
	CREATE TABLE IF NOT EXISTS revision (
		id INTEGER NOT NULL,
		revision BLOB NOT NULL
	);
	CREATE INDEX IF NOT EXISTS item_id ON vault(id);
	CREATE INDEX IF NOT EXISTS item_name_type ON vault(name,type);
	`

	stmtGetRevision    = `SELECT revision FROM revision WHERE id=0`
	stmtUpdateRevision = `UPDATE revision SET revision = ? WHERE id=0`

	stmtCreateItem   = `INSERT INTO vault (id, name, type, hash, data) VALUES(?,?,?,?,?)`
	stmtGetItem      = `SELECT data FROM vault where name = ? AND type = ?`
	stmtGetItemsList = `SELECT id, name, type, hash FROM vault ORDER BY name ASC`
	stmtUpdateItems  = `UPDATE vault SET name = ?, hash = ? , data = ? WHERE id = ?`
	stmtDeleteItems  = `DELETE FROM vault WHERE id = ?`
)

// prepareStatements is a helper which prepares SQL statements for database.
func (s *SQLite) prepareStatements(ctx context.Context) (err error) {
	stmtsDict := map[string]string{
		"getRev":       stmtGetRevision,
		"updateRev":    stmtUpdateRevision,
		"createItems":  stmtCreateItem,
		"getItem":      stmtGetItem,
		"getItemsList": stmtGetItemsList,
		"updateItems":  stmtUpdateItems,
		"deleteItems":  stmtDeleteItems,
	}

	s.stmts = make(map[string]*sql.Stmt)

	//nolint:sqlclosecheck // statements have to be open
	for k, v := range stmtsDict {
		if s.stmts[k], err = s.db.PrepareContext(ctx, v); err != nil {
			return err
		}
	}

	return nil
}
