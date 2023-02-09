package db

import (
	"context"
	"errors"
	"fmt"
	"sync"

	sq "github.com/Masterminds/squirrel"
	"github.com/artfuldog/gophkeeper/internal/logger"
	"github.com/jackc/pgx/v4/pgxpool"
)

// Posgtre represents PostgreSQL implemenation of DB.
type Posgtre struct {
	// Database connections parameters
	config *pgxpool.Config
	// PGX connections pool
	pool *pgxpool.Pool
	// DB tables statements (schema)
	tables []PGTable
	// Logger
	logger logger.L
	// DSN string
	DSN string
	// Squirell statement builder for Postgres $ placeholder configuration
	psql sq.StatementBuilderType
	// Maximum size of secret in bytes
	maxSecretSize uint32
	// Mutex for sync connection and clear operations, which may concrutently use structure fields
	mu sync.Mutex
}

var _ DB = (*Posgtre)(nil)

// newPosgtre is used to create new DBPostres instance.
//
// Supported DSN formats:
//   - postgres://address/db
//   - postgres://user@address/db
//   - postgres://user:secret@address/db
func newPosgtre(params *Parameters, logger logger.L) (*Posgtre, error) {
	if params.address == "" {
		return nil, errors.New("missed database address")
	}

	componentName := "newDBPosgtre"

	db := new(Posgtre)

	db.logger = logger

	//nolint:gocritic
	if params.user == "" {
		db.DSN = fmt.Sprintf("postgres://%s", params.address)
		db.logger.Warn(nil, "used anonymous access to db", componentName)
	} else if params.password == "" {
		db.DSN = fmt.Sprintf("postgres://%s@%s", params.user, params.address)
		db.logger.Warn(nil, "user without password", componentName)
	} else {
		db.DSN = fmt.Sprintf("postgres://%s:%s@%s", params.user, params.password, params.address)
	}

	db.tables = createDBSchema(params)

	db.psql = sq.StatementBuilder.PlaceholderFormat(sq.Dollar)
	db.maxSecretSize = params.maxSecretSize

	return db, nil
}

// Connect is used for connect to database.
func (db *Posgtre) Connect(ctx context.Context) (err error) {
	db.mu.Lock()
	defer db.mu.Unlock()

	db.config, err = pgxpool.ParseConfig(db.DSN)
	if err != nil {
		return
	}

	db.pool, err = pgxpool.ConnectConfig(ctx, db.config)
	if err != nil {
		return
	}

	return
}

// Setup builds required tables in database.
func (db *Posgtre) Setup(ctx context.Context) (err error) {
	db.mu.Lock()
	defer db.mu.Unlock()

	componentName := "DBPosgtre:setup"

	for _, t := range db.tables {
		ct, execErr := db.pool.Exec(ctx, t.Statement)
		if execErr != nil {
			return execErr
		}

		db.logger.Info(fmt.Sprintf("table '%s': %s", t.Name, ct.String()), componentName)
	}

	return
}

// ConnectAndSetup does same as sequentially calling Connect and Setup function.
func (db *Posgtre) ConnectAndSetup(ctx context.Context) (err error) {
	if err := db.Connect(ctx); err != nil {
		return err
	}

	if err := db.Setup(ctx); err != nil {
		return err
	}

	return nil
}

// Run is used for controls database connections.
//
// Run uses context and closing channel for gracefully shutdown database's connections.
// After context expired or cancel function Run will close opened connections
// and close channel.
func (db *Posgtre) Run(ctx context.Context, closeCh CloseChannel) {
	componentName := "DBPosgtre:run"
	db.logger.Info("DB is running", componentName)

	<-ctx.Done()

	db.pool.Close()
	db.logger.Info("DB is stopped", componentName)
	close(closeCh)
}

// Clear is used to delete all database's tables and records.
func (db *Posgtre) Clear(ctx context.Context) {
	db.mu.Lock()
	defer db.mu.Unlock()

	componentName := "DBPosgtre:clear"

	rTables := make([]PGTable, (len(db.tables)))
	for i := 0; i < len(rTables); i++ {
		rTables[i] = db.tables[len(rTables)-1-i]
	}

	for _, t := range rTables {
		ct, err := db.pool.Exec(ctx, fmt.Sprintf("drop table if exists %s cascade", t.Name))
		if err != nil {
			db.logger.Error(err, fmt.Sprintf("table '%s': %s", t.Name, ct.String()), componentName)

			continue
		}

		db.logger.Info(fmt.Sprintf("table '%s': %s", t.Name, ct.String()), "componentName")
	}
}

func (db *Posgtre) GetMaxSecretSize() uint32 {
	return db.maxSecretSize
}
