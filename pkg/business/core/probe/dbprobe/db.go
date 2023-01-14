package dbprobe

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"github.com/vincoll/vigie/internal/api/dbpgx"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.uber.org/zap"
)

var (
	ErrDBNotFound            = errors.New("not found")
	ErrDBDuplicatedEntry     = errors.New("duplicated entry")
	ErrNotFound              = errors.New("user not found")
	ErrInvalidID             = errors.New("ID is not in its proper form")
	ErrInvalidEmail          = errors.New("email is not valid")
	ErrUniqueEmail           = errors.New("email is not unique")
	ErrAuthenticationFailure = errors.New("authentication failed")
)

// ProbeDB manages the entry of a probe in the DB
type ProbeDB struct {
	log          *zap.SugaredLogger
	cdb          *dbpgx.Client
	tr           dbpgx.Transactor
	extContext   sqlx.ExtContext
	isWithinTran bool
}

// NewProbeDB constructs a data for api access.
func NewProbeDB(log *zap.SugaredLogger, db *dbpgx.Client) ProbeDB {
	return ProbeDB{
		log:        log,
		cdb:        db,
		tr:         db.Pool,
		extContext: db.Pool,
	}
}

// WithinTran runs passed function and do commit/rollback at the end.
func (s ProbeDB) WithinTran(ctx context.Context, fn func(sqlx.ExtContext) error) error {
	if s.isWithinTran {
		return fn(s.extContext)
	}
	return s.cdb.WithinTran(ctx, s.tr, fn)
}

// Tran return new ProbeDB with transaction in it.
func (s ProbeDB) Tran(tx sqlx.ExtContext) ProbeDB {
	return ProbeDB{
		log:          s.log,
		tr:           s.tr,
		extContext:   tx,
		isWithinTran: true,
	}
}

// Create inserts a new user into the database.
func (s ProbeDB) Create(ctx context.Context, usr ProbeTable) error {
	const q = `
	INSERT INTO tests
		(id,  probe_type, frequency, last_run, probe_data)
	VALUES
		(:id, :probe_type, :frequency, :last_run, :probe_data)`

	if err := s.cdb.NamedExecContext(ctx, s.extContext, q, usr); err != nil {
		return fmt.Errorf("inserting user: %w", err)
	}

	return nil
}

// Update replaces a user document in the database.
func (s ProbeDB) Update(ctx context.Context, usr ProbeTable) error {
	const q = `
	UPDATE
		users
	SET 
		"name" = :name,
		"email" = :email,
		"roles" = :roles,
		"password_hash" = :password_hash,
		"date_updated" = :date_updated
	WHERE
		user_id = :user_id`

	if err := s.cdb.NamedExecContext(ctx, s.extContext, q, usr); err != nil {
		return fmt.Errorf("updating userID[%s]: %w", usr.ID, err)
	}

	return nil
}

// Delete removes a user from the s.cdb.
func (s ProbeDB) Delete(ctx context.Context, userID string) error {
	data := struct {
		UserID string `dbsqlx:"user_id"`
	}{
		UserID: userID,
	}

	const q = `
	DELETE FROM
		users
	WHERE
		user_id = :user_id`

	if err := s.cdb.NamedExecContext(ctx, s.extContext, q, data); err != nil {
		return fmt.Errorf("deleting userID[%s]: %w", userID, err)
	}

	return nil
}

// Query retrieves a list of existing users from the database.
func (s ProbeDB) Query(ctx context.Context, pageNumber int, rowsPerPage int) ([]ProbeTable, error) {
	data := struct {
		Offset      int `db:"offset"`
		RowsPerPage int `db:"rows_per_page"`
	}{
		Offset:      (pageNumber - 1) * rowsPerPage,
		RowsPerPage: rowsPerPage,
	}

	const q = `
	SELECT
		*
	FROM
		test
	ORDER BY
		id
	OFFSET :offset ROWS FETCH NEXT :rows_per_page ROWS ONLY`

	var usrs []ProbeTable
	if err := dbpgx.NamedQuerySlice(ctx, s.log, s.extContext, q, data, &usrs); err != nil {
		return nil, fmt.Errorf("selecting users: %w", err)
	}

	return usrs, nil
}

// QueryByID gets the specified user from the database.
func (s ProbeDB) QueryByID(ctx context.Context, testID string) (ProbeTable, error) {
	data := struct {
		UserID string `db:"id"`
	}{
		UserID: testID,
	}

	const q = `
	SELECT
		*
	FROM
		tests
	WHERE 
		id = :id`

	var usr ProbeTable
	if err := s.cdb.NamedQueryStruct(ctx, s.extContext, q, data, &usr); err != nil {
		return ProbeTable{}, fmt.Errorf("selecting userID[%q]: %w", testID, err)
	}

	return usr, nil
}

// -------------------------------------- PGX

// XQueryByID gets the specified user from the database.
func (s ProbeDB) XQueryByID(ctx context.Context, testID string) (ProbeTable, error) {
	data := struct {
		UserID string `db:"id"`
	}{
		UserID: testID,
	}

	const q = `
	SELECT
		*
	FROM
		public.tests
	WHERE 
		id = :id`

	var usr ProbeTable
	if err := s.cdb.XQueryStruct(ctx, q, data, &usr); err != nil {
		return ProbeTable{}, fmt.Errorf("selecting testID[%q]: %w", testID, err)
	}

	return usr, nil
}

// XCreate probe a new user into the database.
func (s ProbeDB) XCreate3(ctx context.Context, prb ProbeTable) error {

	// Start to Trace the boot of vigie-agi
	tracer := otel.Tracer("db-insert")
	_, span := tracer.Start(ctx, "db-insert")
	defer span.End()
	// https://stackoverflow.com/questions/54619645/named-prepared-statement-in-pgx-lib-how-does-it-work

	id, _ := uuid.NewRandom()

	const q = `
	INSERT INTO tests
		(id,  probe_type, interval, last_run, probe_data, probe_json)
	VALUES
		($1, $2, $3, $4, $5, $6)`

	span.SetAttributes(attribute.String("query", q), attribute.String("type", "probe"))

	if err := s.cdb.XExecContext3(ctx, q,
		id, prb.ProbeType, prb.Interval, prb.LastRun, prb.Probe_data, prb.Probe_json); err != nil {

		span.SetStatus(codes.Error, err.Error())

		s.log.Errorw("Fail to insert into DB",
			"component", "pgx",
			"error", err,
			"query", q,
		)

		return fmt.Errorf("inserting probe: %w", err)
	}

	span.SetStatus(codes.Ok, "Insert test with Success")
	return nil
}

// WithinTran runs passed function and do commit/rollback at the end.
func (s ProbeDB) XWithinTran(ctx context.Context, fn func(sqlx.ExtContext) error) error {
	if s.isWithinTran {
		return fn(s.extContext)
	}
	return s.cdb.WithinTran(ctx, s.tr, fn)
}

// Tran return new ProbeDB with transaction in it.
func (s ProbeDB) XTran(tx sqlx.ExtContext) ProbeDB {
	return ProbeDB{
		log:          s.log,
		tr:           s.tr,
		extContext:   tx,
		isWithinTran: true,
	}
}

/*
func nonrepeatableRead(conn1, conn2 *pgx.Conn, isolationLevel string) {
	tx, err := conn1.Begin(ctx)
	if err != nil {
		panic(err)
	}
	tx.Exec(ctx, "SET TRANSACTION ISOLATION LEVEL "+isolationLevel)

	row := tx.QueryRow(ctx, "SELECT balance FROM users WHERE name='Bob'")
	var balance int
	row.Scan(&balance)
	fmt.Printf("Bob balance at the beginning of transaction: %d\n", balance)

	fmt.Printf("Updating Bob balance to 1000 from connection 2\n")
	_, err = conn2.Exec(ctx, "UPDATE users SET balance = 1000 WHERE name='Bob'")
	if err != nil {
		fmt.Printf("Failed to update Bob balance from conn2  %e", err)
	}

	_, err = tx.Exec(ctx, "UPDATE users SET balance = $1 WHERE name='Bob'", balance+10)
	if err != nil {
		fmt.Printf("Failed to update Bob balance in tx: %v\n", err)
	}

	if err := tx.Commit(ctx); err != nil {
		fmt.Printf("Failed to commit: %v\n", err)
	}
}
*/