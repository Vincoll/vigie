package dbpgx

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jmoiron/sqlx"
	"github.com/vincoll/vigie/foundation/web"
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

// lib/pq errorCodeNames
// https://github.com/lib/pq/blob/master/error.go#L178
const (
	uniqueViolation = "23505"
	undefinedTable  = "42P01"
)

// Set of error variables for CRUD operations.
var (
	ErrDBNotFound        = sql.ErrNoRows
	ErrDBDuplicatedEntry = errors.New("duplicated entry")
	ErrUndefinedTable    = errors.New("undefined table")
)

func NewDBPool(ctx context.Context, pgConfig PGConfig, logger *zap.SugaredLogger) (*Client, error) {

	_, span := otel.Tracer("vigie-boot").Start(ctx, "db-init")
	defer span.End()

	logger.Infow(fmt.Sprintf("Connection to DB on %s/%s with %s", pgConfig.Host, pgConfig.DbName, pgConfig.User), "component", "pg")
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

	// Parsing PG Config
	pgxConfig, err := pgxpool.ParseConfig(pgConfig.Url)
	if err != nil {
		c.logger.Errorw("PG Config is invalid.",
			"error", err.Error())
		return err
	}

	// Prepare to connect and test if connection is valid
	success := false
	retryDelay := 500 * time.Millisecond

	for success == false {

		poolx, err := pgxpool.NewWithConfig(ctx, pgxConfig)
		ctx, ctxCancel := context.WithTimeout(context.Background(), time.Second*5)
		defer ctxCancel()

		// Executes an empty sql statement against the DB
		err = poolx.Ping(ctx)
		// If Err we want to know if this is caused by a network issue, or a PG Issue.
		if err != nil {

			host := strings.Split(pgConfig.Url, "//")[1]
			_, tcperr := net.Dial("tcp", host)

			if tcperr != nil {

				c.logger.Errorw(fmt.Sprintf("Fail to reach DB through TCP! Next try : %s", retryDelay),
					"err", err.Error(),
					"component", "pg")

				c.status = []string{"Trying to connect to the DB"}

			} else {

				c.logger.Errorw(fmt.Sprintf("cannot establish a TCP connection to PG. Next try : %s", retryDelay),
					"err", err.Error(),
					"component", "pg")
			}
			time.Sleep(retryDelay)
			// Multiplicative wait
			retryDelay = retryDelay * 2

		} else {
			success = true
			c.Poolx = poolx

			c.logger.Infow(fmt.Sprintf("PG connection pool (%d to %d) established to: %s/%s with %s", pgxConfig.MinConns, pgxConfig.MaxConns, pgConfig.Host, pgConfig.DbName, pgConfig.User),
				"component", "pg")

		}

	}

	return nil
}

// StatusCheck return an error if the DB cannot be queried with a empty query in less than 5 sec.
func (c *Client) StatusCheck(ctx context.Context) error {
	ctx, ctxCancel := context.WithTimeout(context.Background(), time.Second*5)
	defer ctxCancel()

	// Executes an empty sql statement against the DB
	err := c.Poolx.Ping(ctx)
	if err != nil {
		return err
	}
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
	defer func(Pool *sqlx.DB) {
		err := Pool.Close()
		if err != nil {
			c.logger.Warnf(fmt.Sprintf("PG connection pool have been ask to shutdown, but have not been able to do gracefully."),
				"component", "pg")
		}
	}(c.Pool)
}

//// ---------------------

// Transactor interface needed to begin transaction.
type Transactor interface {
	Beginx() (*sqlx.Tx, error)
}

// WithinTran runs passed function and do commit/rollback at the end.
func (c *Client) WithinTran(ctx context.Context, db Transactor, fn func(sqlx.ExtContext) error) error {
	traceID := web.GetTraceID(ctx)

	c.logger.Infow("begin tran")
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
			c.logger.Errorw("unable to rollback tran", "trace_id", traceID, "ERROR", err)
		}
		c.logger.Infow("rollback tran", "trace_id", traceID)
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
	c.logger.Infow("commit tran", "trace_id", traceID)

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
