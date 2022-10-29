package dbpgx

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"go.uber.org/zap"
)

// Writing simple tracing wrappers for golang pgx and http.Client for Sentry
// https://anymindgroup.com/news/tech-blog/15724/

type PGConfig struct {
	Url      string `toml:"url"` // postgres://postgres:password@localhost:5432/postgres
	Host     string `toml:"host"`
	Port     string `toml:"port"`
	User     string `toml:"user"`
	Password string `toml:"password"`
	DbName   string `toml:"dbname"`
	Disable  string `toml:"disable"`
}

type Client struct {
	Pool   *sqlx.DB
	Poolx  *pgxpool.Pool
	status []string
	logger *zap.SugaredLogger
}

var (
	ErrDBNotFound            = errors.New("not found")
	ErrDBDuplicatedEntry     = errors.New("duplicated entry")
	ErrNotFound              = errors.New("user not found")
	ErrInvalidID             = errors.New("ID is not in its proper form")
	ErrInvalidEmail          = errors.New("email is not valid")
	ErrUniqueEmail           = errors.New("email is not unique")
	ErrAuthenticationFailure = errors.New("authentication failed")
)

// lib/pq errorCodeNames
// https://github.com/lib/pq/blob/master/error.go#L178
const uniqueViolation = "23505"

func NewDBPool(ctx context.Context, pgConfig PGConfig, logger *zap.SugaredLogger) (*Client, error) {

	_, span := otel.Tracer("vigie-boot").Start(ctx, "db-init")
	defer span.End()

	logger.Infof("Connection to DB on %s", pgConfig.Host)
	c := Client{
		status: []string{"Trying to connect to the DB"},
		logger: logger,
	}

	err := c.connect(pgConfig)

	if err != nil {
		logger.Errorf("Unable to connect to database: %v\n", err)
		span.SetStatus(codes.Error, fmt.Sprintf("Unable to connect to database: %v", err))
		return nil, err
	}

	span.SetStatus(codes.Ok, "DB succesfully connected")

	return &c, nil
}

func (c *Client) connect(pgConfig PGConfig) error {

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	pgxConfig, err := pgxpool.ParseConfig(pgConfig.Url)
	if err != nil {
		c.logger.Errorw("PG Config is invalid.",
			"error", err.Error())
		return err
	}

	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s password=%q dbname=%s sslmode=disable",
		pgConfig.Host, pgConfig.Port, pgConfig.User, pgConfig.Password, pgConfig.DbName)

	// Get query client

	// Send basic query
	success := false
	retryDelay := 500 * time.Millisecond

	for success == false {

		poolx, err := pgxpool.NewWithConfig(ctx, pgxConfig)

		pool, err := sqlx.Connect("postgres", psqlInfo)
		if err != nil {

			host := strings.Split(pgConfig.Url, "//")[1]
			_, tcperr := net.Dial("tcp", host)

			if tcperr != nil {

				c.logger.Errorw(fmt.Sprintf("Fail to reach DB through TCP! Next try : %s", retryDelay),
					"err", err.Error(),
					"component", "pg")

				c.status = []string{"Trying to connect to the DB"}

			} else {

				c.logger.Errorw(fmt.Sprintf("cannot reach InfluxDB. Next try : %s", retryDelay),
					"err", err.Error(),
					"component", "pg")
			}
			time.Sleep(retryDelay)
			// Multiplicative wait
			retryDelay = retryDelay * 2

		} else {
			success = true
			c.Pool = pool
			c.Poolx = poolx

		}

	}

	return nil
}

func (c *Client) StatusCheck(ctx context.Context) error {
	return nil
}

func (c *Client) Health() bool {

	//c.status = "ok"
	return true
}

func (c *Client) GracefulShutdown() error {

	//c.status = "ok"
	return nil
}

func (c *Client) Exit() {
	c.status = []string{"Shutdown"}
	defer c.Pool.Close()
}

//// ---------------------

// Transactor interface needed to begin transaction.
type Transactor interface {
	Beginx() (*sqlx.Tx, error)
}

// WithinTran runs passed function and do commit/rollback at the end.
func (c *Client) WithinTran(ctx context.Context, db Transactor, fn func(sqlx.ExtContext) error) error {
	//	traceID := web.GetTraceID(ctx)

	// Begin the transaction.
	c.logger.Infow("begin tran", "traceid", "traceID")
	tx, err := db.Beginx()
	if err != nil {
		return fmt.Errorf("begin tran: %w", err)
	}

	// Mark to the defer function a rollback is required.
	mustRollback := true

	// Set up a defer function for rolling back the transaction. If
	// mustRollback is true it means the call to fn failed, and we
	// need to roll back the transaction.
	defer func() {
		if mustRollback {
			c.logger.Infow("rollback tran", "traceid", "traceID")
			if err := tx.Rollback(); err != nil {
				c.logger.Errorw("unable to rollback tran", "traceid", "traceID", "ERROR", err)
			}
		}
	}()

	// Execute the code inside the transaction. If the function
	// fails, return the error and the defer function will roll back.
	if err := fn(tx); err != nil {

		// Checks if the error is of code 23505 (unique_violation).
		if pqerr, ok := err.(*pq.Error); ok && pqerr.Code == uniqueViolation {
			return err //ErrDBDuplicatedEntry
		}
		return fmt.Errorf("exec tran: %w", err)
	}

	// Disarm the deferred rollback.
	mustRollback = false

	// Commit the transaction.
	c.logger.Infow("commit tran", "traceid", "traceID")
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit tran: %w", err)
	}

	return nil
}

// NamedExecContext is a helper function to execute a CUD operation with
// logging and tracing.
func (c *Client) NamedExecContext(ctx context.Context, db sqlx.ExtContext, query string, data any) error {
	q := queryString(query, data)
	c.logger.Infow("database.NamedExecContext", "traceid", "web.GetTraceID(ctx)", "query", q)

	if _, err := sqlx.NamedExecContext(ctx, db, query, data); err != nil {

		// Checks if the error is of code 23505 (unique_violation).
		if pqerr, ok := err.(*pq.Error); ok && pqerr.Code == uniqueViolation {
			return err //ErrDBDuplicatedEntry
		}
		return err
	}

	return nil
}

// -> Cannot have generics with func( c Client) ....
// NamedQuerySlice is a helper function for executing queries that return a
// collection of data to be unmarshalled into a slice.
func (c *Client) NamedQuerySlice(ctx context.Context, dbsqlx sqlx.ExtContext, query string, data any, dest *any) error {
	q := queryString(query, data)
	c.logger.Infow("database.NamedQuerySlice", "traceid", "web.GetTraceID(ctx)", "query", q)

	rows, err := sqlx.NamedQueryContext(ctx, dbsqlx, query, data)
	if err != nil {
		return err
	}
	defer rows.Close()

	var slice []any
	for rows.Next() {
		v := new(any)
		if err := rows.StructScan(v); err != nil {
			return err
		}
		slice = append(slice, *v)
	}
	*dest = slice

	return nil
}

// NamedQueryStruct is a helper function for executing queries that return a
// single value to be unmarshalled into a struct type.
func (c *Client) NamedQueryStruct(ctx context.Context, db sqlx.ExtContext, query string, data any, dest any) error {
	q := queryString(query, data)
	c.logger.Infow("database.NamedQueryStruct", "traceid", "web.GetTraceID(ctx)", "query", q)

	rows, err := sqlx.NamedQueryContext(ctx, db, query, data)
	if err != nil {
		return err
	}
	defer rows.Close()

	if !rows.Next() {
		return err //ErrDBNotFound
	}

	if err := rows.StructScan(dest); err != nil {
		return err
	}

	return nil
}

// queryString provides a pretty print version of the query and parameters.
func queryString(query string, args ...any) string {
	query, params, err := sqlx.Named(query, args)
	if err != nil {
		return err.Error()
	}

	for _, param := range params {
		var value string
		switch v := param.(type) {
		case string:
			value = fmt.Sprintf("%q", v)
		case []byte:
			value = fmt.Sprintf("%q", string(v))
		default:
			value = fmt.Sprintf("%v", v)
		}
		query = strings.Replace(query, "?", value, 1)
	}

	query = strings.ReplaceAll(query, "\t", "")
	query = strings.ReplaceAll(query, "\n", " ")

	return strings.Trim(query, " ")
}

func NamedQuerySlice[T any](ctx context.Context, log *zap.SugaredLogger, db sqlx.ExtContext, query string, data any, dest *[]T) error {
	q := queryString(query, data)
	log.Infow("database.NamedQuerySlice", "traceid", "web.GetTraceID(ctx)", "query", q)

	rows, err := sqlx.NamedQueryContext(ctx, db, query, data)
	if err != nil {
		return err
	}
	defer rows.Close()

	var slice []T
	for rows.Next() {
		v := new(T)
		if err := rows.StructScan(v); err != nil {
			return err
		}
		slice = append(slice, *v)
	}
	*dest = slice

	return nil
}

///// -------------------------------------------
/////  PGX
/////- ------------------------------------------

// queryString provides a pretty print version of the query and parameters.
func queryStringX(query string, args ...any) string {
	query, params, err := sqlx.Named(query, args)
	if err != nil {
		return err.Error()
	}

	for _, param := range params {
		var value string
		switch v := param.(type) {
		case string:
			value = fmt.Sprintf("'%s'", v)
		case pgtype.Timestamp:
			value = fmt.Sprintf("'%v'", v.Time.Format(time.RFC3339))
		case pgtype.Interval:
			value = fmt.Sprintf("'%v'", v)
		case []byte:
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

func (c *Client) XExecContext(ctx context.Context, query string, data any) error {
	q := queryStringX(query, data)
	c.logger.Infow("database.NamedExecContext", "traceid", "web.GetTraceID(ctx)", "query", q)

	r, err := c.Poolx.Exec(ctx, q)
	if err != nil {

	}
	fmt.Printf(r.String())

	return nil
}

func (c *Client) XExecContext3(ctx context.Context, query string, args ...interface{}) error {

	c.logger.Infow("database.NamedExecContext", "traceid", "web.GetTraceID(ctx)", "query", "querryyyy")

	r, err := c.Poolx.Exec(ctx, query, args...)
	if err != nil {
		return err
	}
	fmt.Printf(r.String())

	return nil
}

func (c *Client) XExecContext2(ctx context.Context, query string, data []interface{}) error {

	c.logger.Infow("database.NamedExecContext", "traceid", "web.GetTraceID(ctx)", "query", "querryyyy")

	r, err := c.Poolx.Exec(ctx, query, data...)
	if err != nil {

	}
	fmt.Printf(r.String())

	return nil
}

func (c *Client) XQueryStruct(ctx context.Context, query string, data any, dest any) error {
	q := queryStringX(query, data)
	c.logger.Infow("database.NamedQueryStruct", "traceid", "web.GetTraceID(ctx)", "query", q)

	// https://github.com/georgysavva/scany

	rows, err := c.Poolx.Query(ctx, q)
	if err != nil {
		return err
	}
	defer rows.Close()

	if !rows.Next() {
		return err //ErrDBNotFound
	}

	if err := rows.Scan(dest); err != nil {
		return err
	}

	return nil
}

// Transactor interface needed to begin transaction.
type XTransactor interface {
	Beginx() (*sqlx.Tx, error)
}

// XWithinTran runs passed function and do commit/rollback at the end.
func (c *Client) XWithinTran(ctx context.Context, db Transactor, fn func(sqlx.ExtContext) error) error {
	//	traceID := web.GetTraceID(ctx)
	/*
		// Begin the transaction.
		c.logger.Infow("begin tran", "traceid", "traceID")
		tx, err := c.Poolx.Begin(ctx)
		if err != nil {
			return err
		}

		if err != nil {
			return fmt.Errorf("begin tran: %w", err)
		}

		// Mark to the defer function a rollback is required.
		mustRollback := true

		// Set up a defer function for rolling back the transaction. If
		// mustRollback is true it means the call to fn failed, and we
		// need to roll back the transaction.
		defer func() {
			if mustRollback {
				c.logger.Infow("rollback tran", "traceid", "traceID")
				if err := tx.Rollback(ctx); err != nil {
					c.logger.Errorw("unable to rollback tran", "traceid", "traceID", "ERROR", err)
				}
			}
		}()

		// Execute the code inside the transaction. If the function
		// fails, return the error and the defer function will roll back.
		if err := fn(tx); err != nil {

			// Checks if the error is of code 23505 (unique_violation).
			if pqerr, ok := err.(*pq.Error); ok && pqerr.Code == uniqueViolation {
				return err //ErrDBDuplicatedEntry
			}
			return fmt.Errorf("exec tran: %w", err)
		}

		// Disarm the deferred rollback.
		mustRollback = false

		// Commit the transaction.
		c.logger.Infow("commit tran", "traceid", "traceID")
		if err := tx.Commit(ctx); err != nil {
			return fmt.Errorf("commit tran: %w", err)
		}
	*/
	return nil
}
