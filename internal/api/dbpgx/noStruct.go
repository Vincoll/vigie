package dbpgx

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

// StatusCheck returns nil if it can successfully talk to the database. It
// returns a non-nil error otherwise.
func StatusCheck(ctx context.Context, db *sqlx.DB) error {
	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, time.Second)
		defer cancel()
	}

	var pingError error
	for attempts := 1; ; attempts++ {
		pingError = db.Ping()
		if pingError == nil {
			break
		}
		time.Sleep(time.Duration(attempts) * 100 * time.Millisecond)
		if ctx.Err() != nil {
			return ctx.Err()
		}
	}

	if ctx.Err() != nil {
		return ctx.Err()
	}

	// Run a simple query to determine connectivity.
	// Running this query forces a round trip through the database.
	const q = `SELECT true`
	var tmp bool
	return db.QueryRowContext(ctx, q).Scan(&tmp)
}

// WithinTran runs passed function and do commit/rollback at the end.
func WithinTran(ctx context.Context, log *zap.SugaredLogger, db *sqlx.DB, fn func(*sqlx.Tx) error) error {
	log.Debug(ctx, "begin tran")
	tx, err := db.Beginx()
	if err != nil {
		return fmt.Errorf("begin tran: %w", err)
	}

	// We can defer the rollback since the code checks if the transaction
	// has already been committed.
	defer func() {
		if err := tx.Rollback(); err != nil {
			if errors.Is(err, sql.ErrTxDone) {
				return
			}
			log.Error(ctx, "unable to rollback tran", "msg", err)
		}
		log.Debug(ctx, "rollback tran")
	}()

	if err := fn(tx); err != nil {
		if pqerr, ok := err.(*pgconn.PgError); ok && pqerr.Code == uniqueViolation {
			return ErrDBDuplicatedEntry
		}
		return fmt.Errorf("exec tran: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit tran: %w", err)
	}
	log.Debug(ctx, "commit tran")

	return nil
}

// ExecContext is a helper function to execute a CUD operation with
// logging and tracing.
func ExecContext(ctx context.Context, log *zap.SugaredLogger, db *pgxpool.Pool, query string) error {
	return NamedExecContext(ctx, log, db, query, struct{}{})
}

// NamedExecContext is a helper function to execute a CUD operation with
// logging and tracing where field replacement is necessary.
func NamedExecContext(ctx context.Context, log *zap.SugaredLogger, db *pgxpool.Pool, query string, data any) error {
	q := queryString(query, data)

	log.Debugw(fmt.Sprintf("database.NamedExecContext"), "component", "pg", "query", q)

	/*
		ctx, span := web.AddSpan(ctx, "business.sys.database.exec", attribute.String("query", q))
		defer span.End()
	*/
	//if _, err := sqlx.NamedExecContext(ctx, db, query, data); err != nil {
	if _, err := db.Exec(ctx, query, data); err != nil {

		if pqerr, ok := err.(*pgconn.PgError); ok {
			switch pqerr.Code {
			case undefinedTable:
				return ErrUndefinedTable
			case uniqueViolation:
				return ErrDBDuplicatedEntry
			}
		}
		return err
	}

	return nil
}

// QuerySlice is a helper function for executing queries that return a
// collection of data to be unmarshalled into a slice.
func QuerySlice[T any](ctx context.Context, log *zap.SugaredLogger, db *pgxpool.Pool, query string, dest *[]T) error {
	return namedQuerySlice(ctx, log, db, query, struct{}{}, dest, false)
}

// NamedQuerySlice is a helper function for executing queries that return a
// collection of data to be unmarshalled into a slice where field replacement is
// necessary.
func NamedQuerySlice[T any](ctx context.Context, log *zap.SugaredLogger, db *pgxpool.Pool, query string, data any, dest *[]T) error {
	return namedQuerySlice(ctx, log, db, query, data, dest, false)
}

// NamedQuerySliceUsingIn is a helper function for executing queries that return
// a collection of data to be unmarshalled into a slice where field replacement
// is necessary. Use this if the query has an IN clause.
func NamedQuerySliceUsingIn[T any](ctx context.Context, log *zap.SugaredLogger, db *pgxpool.Pool, query string, data any, dest *[]T) error {
	return namedQuerySlice(ctx, log, db, query, data, dest, true)
}

func namedQuerySlice[T any](ctx context.Context, log *zap.SugaredLogger, db *pgxpool.Pool, query string, data any, dest *[]T, withIn bool) error {
	q := queryString(query, data)

	log.Debugw(fmt.Sprintf("database.NamedQuerySlice"), "component", "pg", "query", q)

	/*
		ctx, span := web.AddSpan(ctx, "business.sys.database.queryslice", attribute.String("query", q))
		defer span.End()
	*/
	var rows pgx.Rows
	var err error

	switch withIn {
	/*
		case true:
			rows, err = func() (*sqlx.Rows, error) {
				named, args, err := sqlx.Named(query, data)
				if err != nil {
					return nil, err
				}

				query, args, err := sqlx.In(named, args...)
				if err != nil {
					return nil, err
				}

				query = db.Rebind(query)
				return db.QueryxContext(ctx, query, args...)
			}()
	*/
	default:
		rows, err = db.Query(ctx, query, data)
		//rows, err = sqlx.NamedQueryContext(ctx, db, query, data)
	}

	if err != nil {
		if pqerr, ok := err.(*pgconn.PgError); ok && pqerr.Code == undefinedTable {
			return ErrUndefinedTable
		}
		return err
	}
	defer rows.Close()

	var slice []T
	/*
		for rows.Next() {
			v := new(T)
			if err := rows.Scan(v); err != nil {
				return err
			}
			slice = append(slice, *v)
		}
		*dest = slice
	*/
	if err := pgxscan.ScanAll(&slice, rows); err != nil {
		return err
	}

	for _, e := range slice {

		// put e into dest
		*dest = append(*dest, e)
	}
	/*
		// Using scanny instead of sqlx for scan
		// https://github.com/georgysavva/scany#features
		err = pgxscan.ScanRow(dest, rows)
		if err != nil {
			return nil
		}


			if err := rows.Scan(dest); err != nil {
				return err
			}
	*/

	return nil
}

// QueryStruct is a helper function for executing queries that return a
// single value to be unmarshalled into a struct type where field replacement is necessary.
func QueryStruct(ctx context.Context, log *zap.SugaredLogger, db *pgxpool.Pool, query string, dest any) error {
	return namedQueryStruct(ctx, log, db, query, struct{}{}, dest, false)
}

// NamedQueryStruct is a helper function for executing queries that return a
// single value to be unmarshalled into a struct type where field replacement is necessary.
func NamedQueryStruct(ctx context.Context, log *zap.SugaredLogger, db *pgxpool.Pool, query string, data any, dest any) error {
	return namedQueryStruct(ctx, log, db, query, data, dest, false)
}

// NamedQueryStructUsingIn is a helper function for executing queries that return
// a single value to be unmarshalled into a struct type where field replacement
// is necessary. Use this if the query has an IN clause.
func NamedQueryStructUsingIn(ctx context.Context, log *zap.SugaredLogger, db *pgxpool.Pool, query string, data any, dest any) error {
	return namedQueryStruct(ctx, log, db, query, data, dest, true)
}

func _namedQueryStruct(ctx context.Context, log *zap.SugaredLogger, db sqlx.ExtContext, query string, data any, dest any, withIn bool) error {
	q := queryString(query, data)

	log.Debugw(fmt.Sprintf("database.NamedQueryStruct"), "component", "pg", "query", q)

	/*
		ctx, span := web.AddSpan(ctx, "business.sys.database.query", attribute.String("query", q))
		defer span.End()
	*/
	var rows *sqlx.Rows
	var err error

	switch withIn {
	case true:
		rows, err = func() (*sqlx.Rows, error) {
			named, args, err := sqlx.Named(query, data)
			if err != nil {
				return nil, err
			}

			query, args, err := sqlx.In(named, args...)
			if err != nil {
				return nil, err
			}

			query = db.Rebind(query)
			return db.QueryxContext(ctx, query, args...)
		}()

	default:
		rows, err = sqlx.NamedQueryContext(ctx, db, query, data)
	}

	if err != nil {
		if pqerr, ok := err.(*pgconn.PgError); ok && pqerr.Code == undefinedTable {
			return ErrUndefinedTable
		}
		return err
	}
	defer rows.Close()

	if !rows.Next() {
		return ErrDBNotFound
	}

	if err := rows.StructScan(dest); err != nil {
		return err
	}

	return nil
}

func namedQueryStruct(ctx context.Context, log *zap.SugaredLogger, db *pgxpool.Pool, query string, data any, dest any, withIn bool) error {
	q := queryString(query, data)

	log.Debugw(fmt.Sprintf("database.NamedQueryStruct"), "component", "pg", "query", q)

	/*
		ctx, span := web.AddSpan(ctx, "business.sys.database.query", attribute.String("query", q))
		defer span.End()
	*/
	var rows pgx.Rows
	var err error

	switch withIn {
	case true:
		/*
			rows, err = func() (*sqlx.Rows, error) {
				named, args, err := sqlx.Named(query, data)
				if err != nil {
					return nil, err
				}

				query, args, err := sqlx.In(named, args...)
				if err != nil {
					return nil, err
				}

				query = db.Rebind(query)
				return db.QueryxContext(ctx, query, args...)
			}()
		*/
	default:
		rows, err = db.Query(ctx, query, data)
	}

	if err != nil {
		if pqerr, ok := err.(*pgconn.PgError); ok && pqerr.Code == undefinedTable {
			return ErrUndefinedTable
		}
		return err
	}
	defer rows.Close()

	if !rows.Next() {
		return ErrDBNotFound
	}

	// Using scanny instead of sqlx for scan
	// https://github.com/georgysavva/scany#features
	err = pgxscan.ScanRow(dest, rows)
	if err != nil {
		return nil
	}

	/*
		if err := rows.Scan(dest); err != nil {
			return err
		}
	*/

	return nil
}

// queryString provides a pretty print version of the query and parameters.
func queryString(query string, args any) string {
	query, params, err := sqlx.Named(query, args)
	if err != nil {
		return err.Error()
	}

	for _, param := range params {
		var value string
		switch v := param.(type) {
		case string:
			value = fmt.Sprintf("'%s'", v)
		case []byte:
			value = fmt.Sprintf("'%s'", string(v))
		case pgtype.Timestamp:
			value = fmt.Sprintf("'%s'", v.Time.Format(time.RFC3339Nano))
		case uuid.UUID:
			value = fmt.Sprintf("'%s'", v)
		default:
			value = fmt.Sprintf("%v", v)
		}
		query = strings.Replace(query, "?", value, 1)
	}

	query = strings.ReplaceAll(query, "\t", "")
	query = strings.ReplaceAll(query, "\n", " ")

	return strings.Trim(query, " ")
}
