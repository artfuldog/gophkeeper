// Package db represents interface for iteracting with database.
//
// Package provides functions for neccessary CRUD operations and control db connections.
package db

import (
	"context"
	"errors"
	"fmt"

	"github.com/artfuldog/gophkeeper/internal/logger"
	"github.com/artfuldog/gophkeeper/internal/pb"
)

// Supported databases
const (
	TypePostgres = "postgres"
)

// Database fields' constraints
const (
	FRegexUsername = `^[a-zA-Z0-9_\-\.]+$`
	FRegexEmail    = `^[A-Za-z0-9._%-]+@[A-Za-z0-9.-]+[.][A-Za-z]+$`
)

// Custom types' definitions
type (
	SQLStatement = string
	Username     = string
	Password     = string
	OTPKey       = string
	ItemName     = string
	ItemType     = string
	CloseChannel = chan struct{}
)

// Errors
var (
	ErrDuplicateEntry      = errors.New(`duplicate entry`)
	ErrNotFound            = errors.New("entry not found")
	ErrConstraintViolation = errors.New("wrong field format")
	ErrOperationFailed     = errors.New("operation failed")
	ErrTransactionFailed   = errors.New("operation failed")
	ErrBadSQLQuery         = errors.New(`bad query`)
	ErrInternalDBError     = errors.New("DB error")
	ErrUndefinedError      = errors.New("undefined error")
)

// DB represents general Database inteface.
//
// DB contains all methods, which particular implemenation of DB must support.
type DB interface {
	// Perfrom initial connect to database.
	Connect(context.Context) error
	// Perfrom initial connect to database.
	Setup(context.Context) error
	// Perfrom initial connect to database and setup db schema.
	ConnectAndSetup(context.Context) error
	// Start interactions with database and control connetions.
	Run(context.Context, CloseChannel)
	// Delete all database's tables and records.
	Clear(context.Context)
	// Returns maximum available size of secret.
	GetMaxSecretSize() uint32

	// Register/create new user.
	CreateUser(context.Context, *pb.User) error
	// Read user data.
	GetUserByName(context.Context, Username) (*pb.User, error)
	// Return user's password hash and verification code.
	GetUserAuthData(context.Context, Username) (Password, OTPKey, error)
	// Return user's encryption key.
	GetUserEKey(context.Context, Username) ([]byte, error)
	// Return user's vault revision.
	GetUserRevision(context.Context, Username) ([]byte, error)
	// Update user's information. Empty fields are ignored.
	UpdateUser(context.Context, *pb.User) error
	// Update user's password and encryption key. Empty fields are ignored.
	UpdateUserSecrets(context.Context, *pb.User) error
	// Delete user.
	DeleteUserByName(context.Context, Username) error

	// Create new secured item.
	CreateItem(context.Context, Username, *pb.Item) error
	// Read secured item with provided name and type.
	GetItemByNameAndType(context.Context, Username, ItemName, ItemType) (*pb.Item, error)
	// Returns short representationg of all user's items
	GetItemList(ctx context.Context, username Username) ([]*pb.ItemShort, error)
	// Updates existing item.
	UpdateItem(context.Context, Username, *pb.Item) error
	// Delete item.
	DeleteItem(ctx context.Context, username Username, itemID int64) error
}

// New is a fabric method for create DB with provided type
func New(dbType string, params *DBParameters, logger logger.L) (DB, error) {
	switch dbType {
	case TypePostgres:
		return newDBPosgtre(params, logger)
	default:
		return nil, fmt.Errorf("undefined database type: %s", dbType)
	}
}

// DBParameters contains parameters for connection to database.
type DBParameters struct {
	address       string
	user          string
	password      string
	maxSecretSize uint32
}

// NewDBParameters creates new database connection parameters.
func NewDBParameters(address string, user string, password string, maxSecret uint32) *DBParameters {
	return &DBParameters{
		address:       address,
		user:          user,
		password:      password,
		maxSecretSize: maxSecret,
	}
}
