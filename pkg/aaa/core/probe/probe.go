package probe

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/vincoll/vigie/internal/api/dbsqlx"
	"github.com/vincoll/vigie/pkg/aaa/core/probe/db"
	"go.uber.org/zap"
)

// Package user provides an example of a core business API. Right now these
// calls are just wrapping the data/data layer. But at some point you will
// want auditing or something that isn't specific to the data/store layer.

// Set of error variables for CRUD operations.

// Core manages the set of APIs for user access.
type Core struct {
	store db.Store
}

// NewCore constructs a core for user api access.
func NewCore(log *zap.SugaredLogger, sqlxDB *sqlx.DB) Core {
	return Core{
		store: db.NewStore(log, sqlxDB),
	}
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

// Create inserts a new user into the database.
func (c Core) Create(ctx context.Context, nu NewUser, now time.Time) (User, error) {

	/*
		if err := validate.Check(nu); err != nil {
			return User{}, fmt.Errorf("validating data: %w", err)
		}
	*/
	rand := rand.Intn(100000000000000)

	dbUsr := db.ProbeDB{
		ID:   string(rand),
		Name: nu.Name,
	}

	// This provides an example of how to execute a transaction if required.
	tran := func(tx sqlx.ExtContext) error {
		if err := c.store.Tran(tx).Create(ctx, dbUsr); err != nil {
			if errors.Is(err, dbsqlx.ErrDBDuplicatedEntry) {
				return fmt.Errorf("create: %w", ErrUniqueEmail)
			}
			return fmt.Errorf("create: %w", err)
		}
		return nil
	}

	if err := c.store.WithinTran(ctx, tran); err != nil {
		return User{}, fmt.Errorf("tran: %w", err)
	}

	return toUser(dbUsr), nil
}

// Update replaces a user document in the database.
func (c Core) Update(ctx context.Context, userID string, uu UpdateUser, now time.Time) error {

	/*
		if err := validate.CheckID(userID); err != nil {
			return ErrInvalidID
		}

	*/

	dbUsr, err := c.store.QueryByID(ctx, userID)
	if err != nil {
		if errors.Is(err, dbsqlx.ErrDBNotFound) {
			return ErrNotFound
		}
		return fmt.Errorf("updating user userID[%s]: %w", userID, err)
	}

	if err := c.store.Update(ctx, dbUsr); err != nil {
		if errors.Is(err, dbsqlx.ErrDBDuplicatedEntry) {
			return fmt.Errorf("updating user userID[%s]: %w", userID, ErrUniqueEmail)
		}
		return fmt.Errorf("update: %w", err)
	}

	return nil
}

// Delete removes a user from the database.
func (c Core) Delete(ctx context.Context, userID string) error {
	/*
		if err := validate.CheckID(userID); err != nil {
			return ErrInvalidID
		}*/

	if err := c.store.Delete(ctx, userID); err != nil {
		return fmt.Errorf("delete: %w", err)
	}

	return nil
}

// Query retrieves a list of existing users from the database.
func (c Core) Query(ctx context.Context, pageNumber int, rowsPerPage int) ([]User, error) {
	dbUsers, err := c.store.Query(ctx, pageNumber, rowsPerPage)
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}

	return toUserSlice(dbUsers), nil
}

// QueryByID gets the specified user from the database.
func (c Core) QueryByID(ctx context.Context, userID string) (User, error) {
	/*
		if err := validate.CheckID(userID); err != nil {
			return User{}, ErrInvalidID
		}
	*/
	dbUsr, err := c.store.QueryByID(ctx, userID)
	if err != nil {
		if errors.Is(err, dbsqlx.ErrDBNotFound) {
			return User{}, ErrNotFound
		}
		return User{}, fmt.Errorf("query: %w", err)
	}

	return toUser(dbUsr), nil
}

/*
// QueryByEmail gets the specified user from the database by email.
func (c Core) QueryByEmail(ctx context.Context, email string) (User, error) {

	// Email Validate function in validate.
	if !validate.CheckEmail(email) {
		return User{}, ErrInvalidEmail
	}

	dbUsr, err := c.store.QueryByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, database.ErrDBNotFound) {
			return User{}, ErrNotFound
		}
		return User{}, fmt.Errorf("query: %w", err)
	}

	return toUser(dbUsr), nil
}
*/
