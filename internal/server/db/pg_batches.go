package db

import (
	"fmt"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/artfuldog/gophkeeper/internal/common"
	"github.com/artfuldog/gophkeeper/internal/crypt"
	"github.com/artfuldog/gophkeeper/internal/pb"
	"github.com/jackc/pgx/v4"
)

// newCreateItemBatch is a helper function for construct pgx.Batch, used in item creation.
func (db *Posgtre) newCreateItemBatch(username string, item *pb.Item) (*pgx.Batch, error) {
	componentName := "Postgre:newCreateItemBatch"

	b := new(pgx.Batch)
	psql := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

	updated, hash := getHashUpdatedItem(item.Name, item.Type)

	itemSQ := psql.
		Select("id").
		Column(sq.Placeholders(5), item.Name, item.Type, item.Reprompt, hash, updated).
		From("users").Where(sq.Eq{"username": username})

	stmtItem, argsItem, err := psql.
		Insert("items").
		Columns("user_id, name, type, reprompt, hash, updated").
		Select(itemSQ).ToSql()

	if err != nil {
		return nil, err
	}

	db.logger.Debug(fmt.Sprintf("queue SQL: %s , args: %v", stmtItem, argsItem), componentName)
	b.Queue(stmtItem, argsItem...)

	if item.Secrets == nil {
		item.Secrets = new(pb.Secrets)
	}

	secretSQ := psql.
		Select("items.id").
		Column(sq.Placeholders(2), item.Secrets.Notes, item.Secrets.Secret).
		From("items").LeftJoin("users on items.user_id=users.id").
		Where(sq.Eq{"username": username}).
		Where(sq.Eq{"items.name": item.Name})

	stmtSecret, argsSecret, err := psql.
		Insert("secrets").
		Columns("item_id, notes, secret").
		Select(secretSQ).ToSql()

	if err != nil {
		return nil, err
	}

	db.logger.Debug(fmt.Sprintf("queue SQL: %s , args: %v", stmtSecret, argsSecret), componentName)
	b.Queue(stmtSecret, argsSecret...)

	if item.Additions == nil {
		item.Additions = new(pb.Additions)
	}

	// Only login item can contain URIs' fields
	if item.Type != common.ItemTypeLogin {
		item.Additions.Uris = nil
	}

	addsSQ := psql.
		Select("items.id").
		Column(sq.Placeholders(2), item.Additions.Uris, item.Additions.CustomFields).
		From("items").LeftJoin("users on items.user_id=users.id").
		Where(sq.Eq{"username": username}).
		Where(sq.Eq{"items.name": item.Name})

	stmtAdds, argsAdds, err := psql.
		Insert("additions").
		Columns("item_id, uris, custom_fields").
		Select(addsSQ).ToSql()

	if err != nil {
		return nil, err
	}

	db.logger.Debug(fmt.Sprintf("queue SQL: %s , args: %v", stmtAdds, argsAdds), componentName)
	b.Queue(stmtAdds, argsAdds...)

	newRevision := crypt.GetSHA256hash(username + item.Name + item.Type + item.Updated.String())
	stmtRevision, argsRevision, err := psql.
		Update("users").Set("revision", newRevision).Where(sq.Eq{"username": username}).ToSql()

	if err != nil {
		return nil, err
	}

	db.logger.Debug(fmt.Sprintf("queue SQL: %s , args: %v", stmtRevision, argsRevision), componentName)
	b.Queue(stmtRevision, argsRevision...)

	return b, nil
}

// newUpdateItemBatch is a helper function for construct pgx.Batch, used for update item.
func (db *Posgtre) newUpdateItemBatch(username string, item *pb.Item) (*pgx.Batch, error) {
	componentName := "Postgre:newUpdateItemBatch"

	b := new(pgx.Batch)
	psql := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

	updated, hash := getHashUpdatedItem(item.Name, item.Type)

	stmtItem, argsItem, err := psql.
		Update("items").
		Set("name", sq.Expr("coalesce(?, name)", item.Name)).
		Set("reprompt", sq.Expr("coalesce(?, reprompt)", item.Reprompt)).
		Set("updated", updated).
		Set("hash", hash).
		Where(sq.Eq{"id": item.Id}).
		Where(sq.Expr("user_id = (select users.id from users where username = ?)", username)).ToSql()

	if err != nil {
		return nil, err
	}

	db.logger.Debug(fmt.Sprintf("queue SQL: %s , args: %v", stmtItem, argsItem), componentName)
	b.Queue(stmtItem, argsItem...)

	if item.Secrets != nil {
		stmtSecret, argsSecret, err := psql.
			Update("secrets").
			Set("notes", sq.Expr("coalesce(?, notes)", item.Secrets.Notes)).
			Set("secret", sq.Expr("coalesce(?, secret)", item.Secrets.Secret)).
			Where(sq.Eq{"item_id": item.Id}).ToSql()

		if err != nil {
			return nil, err
		}

		db.logger.Debug(fmt.Sprintf("queue SQL: %s , args: %v", stmtSecret, argsSecret), componentName)
		b.Queue(stmtSecret, argsSecret...)
	}

	if item.Additions != nil {
		// Only login item can contain URIs' fields
		if item.Type != common.ItemTypeLogin {
			item.Additions.Uris = nil
		}

		stmtAdds, argsAdds, err := psql.
			Update("additions").
			Set("uris", sq.Expr("coalesce(?, uris)", item.Additions.Uris)).
			Set("custom_fields", sq.Expr("coalesce(?, custom_fields)", item.Additions.CustomFields)).
			Where(sq.Eq{"item_id": item.Id}).ToSql()

		if err != nil {
			return nil, err
		}

		db.logger.Debug(fmt.Sprintf("queue SQL: %s , args: %v", stmtAdds, argsAdds), componentName)
		b.Queue(stmtAdds, argsAdds...)
	}

	newRevision := crypt.GetSHA256hash(username + item.Name + item.Type + item.Updated.String())
	stmtRevision, argsRevision, err := psql.
		Update("users").Set("revision", newRevision).Where(sq.Eq{"username": username}).ToSql()

	if err != nil {
		return nil, err
	}

	db.logger.Debug(fmt.Sprintf("queue SQL: %s , args: %v", stmtRevision, argsRevision), componentName)
	b.Queue(stmtRevision, argsRevision...)

	return b, nil
}

// newDeleteItemBatch is a helper function for construct pgx.Batch, used for deleteitem.
func (db *Posgtre) newDeleteItemBatch(username string, itemID int64) (*pgx.Batch, error) {
	componentName := "Postgre:newDeleteItemBatch"

	b := new(pgx.Batch)
	psql := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

	stmtItem, argsItem, err := psql.
		Delete("items").Where(sq.Eq{"id": itemID}).
		Where(sq.Expr("user_id = (select users.id from users where username = ?)", username)).ToSql()

	if err != nil {
		return nil, err
	}

	db.logger.Debug(fmt.Sprintf("queue SQL: %s , args: %v", stmtItem, argsItem), componentName)
	b.Queue(stmtItem, argsItem...)

	newRevision := crypt.GetSHA256hash(fmt.Sprintf("%s|%d<>%v", username, itemID, time.Now()))
	stmtRevision, argsRevision, err := psql.
		Update("users").Set("revision", newRevision).Where(sq.Eq{"username": username}).ToSql()

	if err != nil {
		return nil, err
	}

	db.logger.Debug(fmt.Sprintf("queue SQL: %s , args: %v", stmtRevision, argsRevision), componentName)
	b.Queue(stmtRevision, argsRevision...)

	return b, nil
}
