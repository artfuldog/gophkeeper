package db

import "fmt"

// PGTable represents desctiption of PosgreSQL' table.
type PGTable struct {
	Name      string       // table name
	Statement SQLStatement // SQL statement for creating table
}

// createDBSchema returns description of required PostgreSQL tables.
//
// item_types constrains:
//   - l - login item
//   - c - card item
//   - n - secured note item
//   - d - secured data item
//
// cutsom_fields_types constrains:
//   - t - plain-text key-value
//   - h - key with hidden value
//   - b - boolean
func createDBSchema(params *Parameters) []PGTable {
	PGTableUsers := PGTable{
		Name: "users",
		Statement: `
			create table if not exists users (
				id int generated always as identity primary key,
				username varchar(50) not null unique check (username ~ '` + FRegexUsername + `'),
				email varchar unique check (email ~ '` + FRegexEmail + `'),
				pwdhash varchar not null check (pwdhash <> ''),
				otpkey varchar,
				ekey bytea not null,
				revision bytea,
				updated timestamptz,
				regdate timestamptz
			)`,
	}
	PGTableItems := PGTable{
		Name: "items",
		Statement: `
			create table if not exists items (
				id int generated always as identity primary key,
				user_id integer not null references users (id) on delete cascade,
				name varchar(128) check (name <> ''),
				type char(1) not null check(type in ('l', 'c', 'n', 'd')),
				reprompt boolean,
				updated timestamptz,
				hash bytea,
				unique (user_id, name, type)
			)`,
	}
	PGTableSecrets := PGTable{
		Name: "secrets",
		Statement: `
			create table if not exists secrets (
				id int generated always as identity primary key,
				item_id integer not null unique references items (id) on delete cascade,
				notes bytea,
				secret bytea check (length(secret) <= ` + fmt.Sprint(params.maxSecretSize+4) + `)
			)`,
	}
	PGTableAdditions := PGTable{
		Name: "additions",
		Statement: `
			create table if not exists additions (
				id int generated always as identity primary key,
				item_id integer not null unique references items (id) on delete cascade,
				uris bytea,
				custom_fields bytea
			)`,
	}

	return []PGTable{
		PGTableUsers,
		PGTableItems,
		PGTableSecrets,
		PGTableAdditions,
	}
}
