package probemgmt

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"github.com/vincoll/vigie/internal/api/dbpgx"
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
	extContext   sqlx.ExtContext
	isWithinTran bool
}

// NewProbeDB constructs a data for api access.
func NewProbeDB(log *zap.SugaredLogger, db *dbpgx.Client) ProbeDB {
	return ProbeDB{
		log:        log,
		cdb:        db,
		extContext: db.Pool,
	}
}

// Tran return new ProbeDB with transaction in it.
func (s ProbeDB) Tran(tx sqlx.ExtContext) ProbeDB {
	return ProbeDB{
		log:          s.log,
		extContext:   tx,
		isWithinTran: true,
	}
}

// Create inserts a new user into the database.
func (s ProbeDB) Create(ctx context.Context, usr ProbeTable) error {
	const q = `
	INSERT INTO tests
		(id,  probe_type, interval, last_run, probe_data)
	VALUES
		(@id, @probe_type, @interval, @last_run, @probe_data)`

	data := pgx.NamedArgs{
		"id":         usr.ID,
		"probe_type": usr.ProbeType,
		"interval":   usr.Interval,
		"last_run":   usr.LastRun,
		"probe_data": usr.Probe_data,
	}

	if err := dbpgx.NamedExecContext(ctx, s.log, s.cdb.Poolx, q, data); err != nil {
		return fmt.Errorf("inserting test: %w", err)
	}

	return nil
}

// Update replaces a user document in the database.
func (s ProbeDB) Update(ctx context.Context, usr ProbeTable) error {
	const q = `
	UPDATE
		tests
	SET 
		"name" = :name,
		"email" = :email,
		"roles" = :roles,
		"password_hash" = :password_hash,
		"date_updated" = :date_updated
	WHERE
		user_id = :user_id`

	if err := dbpgx.NamedExecContext(ctx, s.log, s.cdb.Poolx, q, ""); err != nil {
		return fmt.Errorf("updating userID[%s]: %w", usr.ID, err)
	}

	return nil
}

// Delete removes a user from the s.cdb.
func (s ProbeDB) Delete(ctx context.Context, testID string) error {
	data := pgx.NamedArgs{
		"id": testID,
	}

	const q = `
	DELETE FROM
		users
	WHERE
		user_id = :user_id`

	if err := dbpgx.NamedExecContext(ctx, s.log, s.cdb.Poolx, q, data); err != nil {
		return fmt.Errorf("deleting userID[%s]: %w", testID, err)
	}

	return nil
}

// Query retrieves a list of existing users from the database.
func (s ProbeDB) Query(ctx context.Context, pageNumber int, rowsPerPage int) ([]ProbeTable, error) {

	data := pgx.NamedArgs{
		"Offset":      (pageNumber - 1) * rowsPerPage,
		"RowsPerPage": rowsPerPage,
	}
	const q = `
	SELECT
		*
	FROM
		test
	ORDER BY
		id
	OFFSET :offset ROWS FETCH NEXT :rows_per_page ROWS ONLY`

	var pts []ProbeTable

	if err := dbpgx.NamedQuerySlice(ctx, s.log, s.cdb.Poolx, q, data, &pts); err != nil {
		return nil, fmt.Errorf("selecting tests: %w", err)
	}

	/*
		if err := s.cdb.NamedQuerySlice(ctx, s.extContext, q, data, &pts); err != nil {
			return nil, fmt.Errorf("selecting users: %w", err)
		}
	*/
	return pts, nil
}

// QueryByID gets the specified user from the database.
func (s ProbeDB) QueryByID(ctx context.Context, testID string) (ProbeTable, error) {
	data := pgx.NamedArgs{
		"id": testID,
	}

	const q = `
	SELECT
		*
	FROM
		tests
	WHERE 
		id = @id`

	var pt ProbeTable
	if err := dbpgx.NamedQueryStruct(ctx, s.log, s.cdb.Poolx, q, data, &pt); err != nil {
		return ProbeTable{}, fmt.Errorf("selecting tests: %w", err)
	}
	/*
		if err := s.cdb.XQueryStruct(ctx, q, data, &pt); err != nil {
			return ProbeTable{}, fmt.Errorf("selecting testID[%s]: %w", testID, err)
		}*/

	return pt, nil
}

// QueryByID gets the specified user from the database.
func (s ProbeDB) QueryByType(ctx context.Context, probeType string) ([]ProbeTable, error) {
	data := pgx.NamedArgs{
		"probeType": probeType,
	}

	const q = `
	SELECT
		*
	FROM
		tests
	WHERE 
		probe_type = @probeType`

	var pt []ProbeTable
	if err := dbpgx.NamedQuerySlice(ctx, s.log, s.cdb.Poolx, q, data, &pt); err != nil {
		return nil, fmt.Errorf("selecting tests: %w", err)
	}
	/*
		if err := s.cdb.XQueryStruct(ctx, q, data, &pt); err != nil {
			return ProbeTable{}, fmt.Errorf("selecting testID[%s]: %w", testID, err)
		}*/

	return pt, nil
}

// QueryByID gets the specified user from the database.
func (s ProbeDB) QueryPastInterval(ctx context.Context, probeType string, interval time.Duration) ([]ProbeTable, error) {
	data := pgx.NamedArgs{
		"probeType": probeType,
	}

	sqlWhereProbe := ""
	if probeType != "" {
		sqlWhereProbe = " AND probe_type = @probeType"
	}

	var q = `
	SELECT
		*
	FROM
		tests
	WHERE 
		COALESCE(last_run + interval, NOW()) < NOW()` + sqlWhereProbe

	var pt []ProbeTable
	if err := dbpgx.NamedQuerySlice(ctx, s.log, s.cdb.Poolx, q, data, &pt); err != nil {
		return nil, fmt.Errorf("selecting tests: %w", err)
	}
	/*
		if err := s.cdb.XQueryStruct(ctx, q, data, &pt); err != nil {
			return ProbeTable{}, fmt.Errorf("selecting testID[%s]: %w", testID, err)
		}*/

	return pt, nil
}
