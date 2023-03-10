package db

import (
	"context"
	"fmt"

	"github.com/artfuldog/gophkeeper/internal/pb"
	"github.com/georgysavva/scany/pgxscan"
	"github.com/jackc/pgtype"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// CreateItem creates new item for particular user.
//
// CreateItem generates updated time field in RFC3339 format during creation.
// Returns nil error only on successfully creation.
func (db *Posgtre) CreateItem(ctx context.Context, username string, item *pb.Item) error {
	if username == "" {
		return ErrNotFound
	}

	componentName := "Postgre:CreateItem"

	b, err := db.newCreateItemBatch(username, item)
	if err != nil {
		return stackErrors(ErrInternalDBError, err)
	}

	return db.runBatch(ctx, b, componentName)
}

// GetItemByNameAndType gets item's information from DB.
func (db *Posgtre) GetItemByNameAndType(ctx context.Context, username Username,
	itemName string, itemType string) (*pb.Item, error) {

	componentName := "Postgre:GetItemByNameAndType"

	tx, err := db.beginTxRO(ctx, componentName)
	if err != nil {
		return nil, err
	}
	defer db.deferTxRollback(ctx, tx) //nolint:wsl

	stmtItem, argsItem, err := db.psql.
		Select("items.id, name, type, reprompt, hash").
		Column(`s.notes as "secrets.notes", s.secret as "secrets.secret"`).
		Column(`a.uris as "additions.uris", a.custom_fields as "additions.custom_fields"`).
		From("items").
		LeftJoin("users on user_id=users.id").
		LeftJoin("secrets s on items.id=s.item_id").
		LeftJoin("additions a on items.id=a.item_id").
		Where("users.username=? and items.name=? and items.type=?", username, itemName, itemType).
		ToSql()
	if err != nil {
		return nil, stackErrors(ErrInternalDBError, err)
	}

	db.logger.Debug(fmt.Sprintf("run SQL: %s , args: %v", stmtItem, argsItem), componentName)

	item := new(pb.Item)
	if err := pgxscan.Get(ctx, tx, item, stmtItem, argsItem...); err != nil {
		if pgxscan.NotFound(err) {
			return nil, stackErrors(ErrNotFound, err)
		}

		return nil, wrapPgError(err)
	}

	stmtUpdated, argsUpdated, err := db.psql.
		Select("items.updated").
		From("items").Join("users on user_id=users.id").
		Where("users.username=? and items.name=? and items.type=?", username, itemName, itemType).
		ToSql()
	if err != nil {
		return nil, stackErrors(ErrInternalDBError, err)
	}

	db.logger.Debug(fmt.Sprintf("run SQL: %s , args: %v", stmtUpdated, argsUpdated), componentName)

	var updated pgtype.Timestamptz
	if err := tx.QueryRow(ctx, stmtUpdated, argsUpdated...).Scan(&updated); err != nil {
		return nil, wrapPgError(err)
	}

	if updated.Status == pgtype.Present {
		item.Updated = timestamppb.New(updated.Time)
	}

	return item, nil
}

// GetItemList returns short representation of all user's items.
func (db *Posgtre) GetItemList(ctx context.Context, username Username) ([]*pb.ItemShort, error) {
	componentName := "Postgre:GetItemList"

	tx, err := db.beginTxRO(ctx, componentName)
	if err != nil {
		return nil, err
	}
	defer db.deferTxRollback(ctx, tx) //nolint:wsl

	stmtItems, argsItems, err := db.psql.
		Select("items.id, name, type, items.updated, hash").
		From("items").
		LeftJoin("users on user_id=users.id").
		Where("users.username=?", username).
		OrderBy("name").
		ToSql()
	if err != nil {
		return nil, stackErrors(ErrInternalDBError, err)
	}

	db.logger.Debug(fmt.Sprintf("run SQL: %s , args: %v", stmtItems, argsItems), componentName)

	rows, err := tx.Query(ctx, stmtItems, argsItems...)
	if err != nil {
		return nil, wrapPgError(err)
	}
	defer rows.Close()

	var items []*pb.ItemShort

	rs := pgxscan.NewRowScanner(rows)

	for rows.Next() {
		var item ItemShort
		if err := rs.Scan(&item); err != nil {
			return nil, stackErrors(ErrInternalDBError, err)
		}

		items = append(items, item.toPB())
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return items, nil
}

// GetItemsByID gets item's information from DB.
//
//nolint:cyclop // necessary evil
func (db *Posgtre) GetItemsByID(ctx context.Context, username Username, ids []int64) ([]*pb.Item, error) {
	componentName := "Postgre:GetItemsByID"

	tx, err := db.beginTxRO(ctx, componentName)
	if err != nil {
		return nil, err
	}
	defer db.deferTxRollback(ctx, tx) //nolint:wsl

	stmtItems, argsItems, err := db.psql.
		Select("items.id, name, type, reprompt, hash").
		Column(`s.notes as "secrets.notes", s.secret as "secrets.secret"`).
		Column(`a.uris as "additions.uris", a.custom_fields as "additions.custom_fields"`).
		From("items").
		LeftJoin("users on user_id=users.id").
		LeftJoin("secrets s on items.id=s.item_id").
		LeftJoin("additions a on items.id=a.item_id").
		Where("users.username=? and items.id = any(?)", username, ids).
		OrderBy("items.id").
		ToSql()
	if err != nil {
		return nil, stackErrors(ErrInternalDBError, err)
	}

	db.logger.Debug(fmt.Sprintf("run SQL: %s , args: %v", stmtItems, argsItems), componentName)

	rows, err := tx.Query(ctx, stmtItems, argsItems...)
	if err != nil {
		return nil, wrapPgError(err)
	}
	defer rows.Close()

	var items []*pb.Item

	rs := pgxscan.NewRowScanner(rows)

	for rows.Next() {
		var item pb.Item
		if err := rs.Scan(&item); err != nil {
			return nil, stackErrors(ErrInternalDBError, err)
		}

		items = append(items, &item)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	stmtUpdateds, argsUpdateds, err := db.psql.
		Select("items.updated").
		From("items").Join("users on user_id=users.id").
		Where("users.username=? and items.id = any(?)", username, ids).
		OrderBy("items.id").
		ToSql()
	if err != nil {
		return nil, stackErrors(ErrInternalDBError, err)
	}

	db.logger.Debug(fmt.Sprintf("run SQL: %s , args: %v", stmtUpdateds, argsUpdateds), componentName)

	rows, err = tx.Query(ctx, stmtUpdateds, argsUpdateds...)
	if err != nil {
		return nil, wrapPgError(err)
	}
	defer rows.Close()

	rs = pgxscan.NewRowScanner(rows)
	i := 0

	for rows.Next() {
		var updated pgtype.Timestamptz
		if err := rs.Scan(&updated); err != nil {
			return nil, stackErrors(ErrInternalDBError, err)
		}

		if updated.Status == pgtype.Present {
			items[i].Updated = timestamppb.New(updated.Time)
		}
		i++
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return items, nil
}

// GetItemsByID gets item's information from DB.
func (db *Posgtre) GetItemHashByID(ctx context.Context, id int64) ([]byte, error) {
	componentName := "Postgre:GetItemsByID"

	tx, err := db.beginTxRO(ctx, componentName)
	if err != nil {
		return nil, err
	}
	defer db.deferTxRollback(ctx, tx) //nolint:wsl

	sqlStmt := `select hash from items where id = $1`

	db.logger.Debug(fmt.Sprintf("run SQL: %s, %d", sqlStmt, id), componentName)

	var hash []byte
	if err := db.pool.QueryRow(ctx, sqlStmt, id).Scan(&hash); err != nil {
		if pgxscan.NotFound(err) {
			return nil, stackErrors(ErrNotFound, err)
		}

		return nil, wrapPgError(err)
	}

	return hash, nil
}

// UpdateItem updates existing item.
//
// UpdateItem generates updated time field in RFC3339 format during creation.
// Returns nil error only on successful update.
func (db *Posgtre) UpdateItem(ctx context.Context, username string, item *pb.Item) error {
	if username == "" {
		return ErrNotFound
	}
	componentName := "Postgre:UpdateItem"

	b, err := db.newUpdateItemBatch(username, item)
	if err != nil {
		return stackErrors(ErrInternalDBError, err)
	}

	return db.runBatch(ctx, b, componentName)
}

// GetItemList returns short representation of all user's items.
func (db *Posgtre) DeleteItem(ctx context.Context, username Username, itemID int64) error {
	if username == "" {
		return ErrNotFound
	}
	componentName := "Postgre:DeleteItem"

	b, err := db.newDeleteItemBatch(username, itemID)
	if err != nil {
		return stackErrors(ErrInternalDBError, err)
	}

	return db.runBatch(ctx, b, componentName)
}
