package db

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/vincoll/vigie/internal/api/dbsqlx"
	"go.uber.org/zap"
)

// Store manages the set of APIs for user access.
type Store struct {
	log          *zap.SugaredLogger
	cdb          *dbsqlx.Client
	tr           dbsqlx.Transactor
	extContext   sqlx.ExtContext
	isWithinTran bool
}

// NewStore constructs a data for api access.
func NewStore(log *zap.SugaredLogger, db *dbsqlx.Client) Store {
	return Store{
		log:        log,
		cdb:        db,
		tr:         db.Pool,
		extContext: db.Pool,
	}
}

// WithinTran runs passed function and do commit/rollback at the end.
func (s Store) WithinTran(ctx context.Context, fn func(sqlx.ExtContext) error) error {
	if s.isWithinTran {
		return fn(s.extContext)
	}
	return s.cdb.WithinTran(ctx, s.tr, fn)
}

// Tran return new Store with transaction in it.
func (s Store) Tran(tx sqlx.ExtContext) Store {
	return Store{
		log:          s.log,
		tr:           s.tr,
		extContext:   tx,
		isWithinTran: true,
	}
}

// Create inserts a new user into the database.
func (s Store) Create(ctx context.Context, usr ProbeDB) error {
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
func (s Store) Update(ctx context.Context, usr ProbeDB) error {
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
func (s Store) Delete(ctx context.Context, userID string) error {
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
func (s Store) Query(ctx context.Context, pageNumber int, rowsPerPage int) ([]ProbeDB, error) {
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

	var usrs []ProbeDB
	if err := dbsqlx.NamedQuerySlice(ctx, s.log, s.extContext, q, data, &usrs); err != nil {
		return nil, fmt.Errorf("selecting users: %w", err)
	}

	return usrs, nil
}

// QueryByID gets the specified user from the database.
func (s Store) QueryByID(ctx context.Context, testID string) (ProbeDB, error) {
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

	var usr ProbeDB
	if err := s.cdb.NamedQueryStruct(ctx, s.extContext, q, data, &usr); err != nil {
		return ProbeDB{}, fmt.Errorf("selecting userID[%q]: %w", testID, err)
	}

	return usr, nil
}

// -------------------------------------- PGX

// XQueryByID gets the specified user from the database.
func (s Store) XQueryByID(ctx context.Context, testID string) (ProbeDB, error) {
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

	var usr ProbeDB
	if err := s.cdb.XQueryStruct(ctx, q, data, &usr); err != nil {
		return ProbeDB{}, fmt.Errorf("selecting testID[%q]: %w", testID, err)
	}

	return usr, nil
}

// XCreate inserts a new user into the database.
func (s Store) XCreate3(ctx context.Context, prb ProbeDB) error {

	// https://stackoverflow.com/questions/54619645/named-prepared-statement-in-pgx-lib-how-does-it-work

	const q = `
	INSERT INTO tests
		(id,  probe_type, frequency, interval, last_run, probe_data)
	VALUES
		($1, $2, $3, $4, $5, $6)`

	if err := s.cdb.XExecContext3(ctx, q,
		prb.ID, prb.ProbeType, prb.Frequency, prb.Interval, prb.LastRun, prb.Probe_data); err != nil {
		return fmt.Errorf("inserting user: %w", err)
	}

	return nil
}
