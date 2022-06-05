// Package dbschema contains the database schema, migrations and seeding data.
package dbschema

import (
	"context"
	_ "embed" // Calls init function.
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/vincoll/vigie/internal/api/dbsqlx"
)

var (
	//go:embed sql/schema.sql
	schemaDoc string

	//go:embed sql/seed.sql
	seedDoc string

	//go:embed sql/delete.sql
	deleteDoc string
)

// Migrate attempts to bring the schema for dbsqlx up to date with the migrations
// defined in this package.
func Migrate(ctx context.Context, db *sqlx.DB) error {
	return nil
}

// Seed runs the set of seed-data queries against dbsqlx. The queries are ran in a
// transaction and rolled back if any fail.
func Seed(ctx context.Context, dbc *dbsqlx.Client) error {

	if err := dbc.StatusCheck(ctx); err != nil {
		return fmt.Errorf("status check database: %w", err)
	}

	tx, err := dbc.Pool.Begin()
	if err != nil {
		return err
	}

	if _, err := tx.Exec(seedDoc); err != nil {
		if err := tx.Rollback(); err != nil {
			return err
		}
		return err
	}

	return tx.Commit()
}

// DeleteAll runs the set of Drop-table queries against dbsqlx. The queries are ran in a
// transaction and rolled back if any fail.
func DeleteAll(db *sqlx.DB) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}

	if _, err := tx.Exec(deleteDoc); err != nil {
		if err := tx.Rollback(); err != nil {
			return err
		}
		return err
	}

	return tx.Commit()
}
