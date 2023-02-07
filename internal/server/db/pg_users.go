package db

import (
	"context"
	"errors"
	"fmt"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/artfuldog/gophkeeper/internal/pb"
	"github.com/georgysavva/scany/pgxscan"
)

// CreateUser creates new user.
//
// CreateUser generates regdate and updated fields in RFC3339 format during creation.
// In case of error during creation returns error, returns nil error only on succesfully creation.
func (db *DBPosgtre) CreateUser(ctx context.Context, user *pb.User) error {
	componentName := "DBPosgtre:CreateUser"

	regdate := time.Now().Format(time.RFC3339)

	stmtUser, argsUser, err := db.psql.
		Insert("users").
		Columns("username, email, pwdhash, otpkey, ekey, updated, regdate").
		Values(user.Username, user.Email, user.Pwdhash, user.OtpKey, user.Ekey, regdate, regdate).ToSql()
	if err != nil {
		return stackErrors(ErrInternalDBError, err)
	}

	db.logger.Debug(fmt.Sprintf("run SQL: %s , args: %v", stmtUser, argsUser), componentName)
	_, err = db.pool.Exec(ctx, stmtUser, argsUser...)
	if wrappedErr := wrapPgError(err); wrappedErr != nil {
		return wrappedErr
	}

	return nil
}

// GetUser returns user data (struct User) by provided user login.
//
// If no users were found GetUser returns nil and error (ErrUserNotFound).
func (db *DBPosgtre) GetUserByName(ctx context.Context, username Username) (*pb.User, error) {
	componentName := "DBPosgtre:GetUserByName"

	stmtUser, argsUser, err := db.psql.
		Select("username, email, revision, pwdhash, otpkey, ekey, updated, regdate").
		From("users").Where(sq.Eq{"username": username}).ToSql()
	if err != nil {
		return nil, stackErrors(ErrInternalDBError, err)
	}

	user := new(User)

	db.logger.Debug(fmt.Sprintf("run SQL: %s %v", stmtUser, argsUser), componentName)
	if err := pgxscan.Get(ctx, db.pool, user, stmtUser, argsUser...); err != nil {
		if pgxscan.NotFound(err) {
			return nil, stackErrors(ErrNotFound, err)
		}
		return nil, wrapPgError(err)
	}

	return user.toPB(), nil
}

// GetUserAuthData returns password hash and OTP secret key for particular user.
//
// If no users were found GetUserAuthData returns empty string and error (ErrUserNotFound).
// In case of processing error returns empty string and original error.
func (db *DBPosgtre) GetUserAuthData(ctx context.Context, username Username) (Password, OTPKey, error) {
	componentName := "DBPosgtre:GetUserPwdHash"
	var none string

	sqlStmt := `select pwdhash, coalesce (otpkey, '') from users where username = $1`

	var pwdhash, otpkey string

	db.logger.Debug(fmt.Sprintf("run SQL: %s", sqlStmt), componentName)
	if err := db.pool.QueryRow(ctx, sqlStmt, username).Scan(&pwdhash, &otpkey); err != nil {
		if pgxscan.NotFound(err) {
			return none, none, stackErrors(ErrNotFound, err)
		}
		return none, none, wrapPgError(err)
	}
	return pwdhash, otpkey, nil
}

// GetUserEKey returns user encryption key.
//
// If no users were found GetUserEKey returns empty slice and error (ErrUserNotFound).
// In case of processing error returns empty string and original error.
func (db *DBPosgtre) GetUserEKey(ctx context.Context, username Username) ([]byte, error) {
	componentName := "DBPosgtre:GetUserEKey"

	sqlStmt := `select ekey from users where username = $1`

	var ekey []byte

	db.logger.Debug(fmt.Sprintf("run SQL: %s", sqlStmt), componentName)
	if err := db.pool.QueryRow(ctx, sqlStmt, username).Scan(&ekey); err != nil {
		if pgxscan.NotFound(err) {
			return nil, stackErrors(ErrNotFound, err)
		}
		return nil, wrapPgError(err)
	}
	return ekey, nil
}

// GetUserRevision returns configuration revision for particular user.
//
// If no users were found GetUserRevision returns empty string and error (ErrUserNotFound).
// In case of processing error returns empty string and original error.
func (db *DBPosgtre) GetUserRevision(ctx context.Context, username Username) ([]byte, error) {
	componentName := "DBPosgtre:GetUserRevision"

	sqlStmt := `select revision from users where username = $1`

	var revision []byte

	db.logger.Debug(fmt.Sprintf("run SQL: %s, %s", sqlStmt, username), componentName)
	if err := db.pool.QueryRow(ctx, sqlStmt, username).Scan(&revision); err != nil {
		if pgxscan.NotFound(err) {
			return nil, stackErrors(ErrNotFound, err)
		}
		return nil, wrapPgError(err)
	}
	return revision, nil
}

// UpdateUser updates current user information.
//
// UpdateUser generates updated fields in RFC3339 format during creation.
// In case of error during update returns error, returns nil error only on success.
// ID, username and regdate cannot be updated.
func (db *DBPosgtre) UpdateUser(ctx context.Context, user *pb.User) error {
	componentName := "DBPosgtre:UpdateUser"

	sqlStmt := `
		update users SET
			email = coalesce($1, email),
			pwdhash = coalesce($2, pwdhash),
			otpkey = coalesce($3, otpkey),
			ekey = coalesce($4, ekey),
			revision = coalesce($5, revision),
			updated = coalesce($6, updated)
		WHERE username = $7`

	updated := time.Now().Format(time.RFC3339)

	db.logger.Debug(fmt.Sprintf("run SQL: %s", sqlStmt), componentName)
	ct, err := db.pool.Exec(ctx, sqlStmt, user.Email, user.Pwdhash, user.OtpKey, user.Ekey, user.Revision, updated, user.Username)
	if wrappedErr := wrapPgError(err); wrappedErr != nil {
		return wrappedErr
	}
	if ct.RowsAffected() < 1 {
		return stackErrors(ErrNotFound, errors.New(user.Username))
	}

	return nil
}

// TODO UpdateUserSecrets updates user's password and encryption key.
func (db *DBPosgtre) UpdateUserSecrets(ctx context.Context, user *pb.User) error {
	return nil
}

// DeleteUserByName deletes user by username.
//
// In case of error during deletion DeleteUserByLogin returns error,
// returns nil error only on succesfully deletion.
func (db *DBPosgtre) DeleteUserByName(ctx context.Context, username Username) error {
	componentName := "DBPosgtre:DeleteUserByName"

	sqlStmt := `delete from users cascade where username=$1`

	db.logger.Debug(fmt.Sprintf("run SQL: %s, %s", sqlStmt, username), componentName)
	ct, err := db.pool.Exec(ctx, sqlStmt, username)
	if wrappedErr := wrapPgError(err); wrappedErr != nil {
		return wrappedErr
	}
	if ct.RowsAffected() < 1 {
		return stackErrors(ErrNotFound, errors.New(username))
	}
	return nil
}
