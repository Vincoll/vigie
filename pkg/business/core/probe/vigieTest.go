// Package probe provides an example of a core business API. Right now these
// calls are just wrapping the data/data layer. But at some point you will
// want to audit or something that isn't specific to the data/store layer.
package probe

import (
	"context"
	"errors"
	"time"

	"github.com/vincoll/vigie/internal/api/dbpgx"
	"github.com/vincoll/vigie/pkg/business/core/probe/dbprobe"
	"github.com/vincoll/vigie/pkg/probe"
	"github.com/vincoll/vigie/pkg/probe/assertion"
	"go.uber.org/zap"
)

// VigieTest is the expected struct received by the REST API
type VigieTest struct {
	Metadata   probe.Metadata         `json:"metadata"`
	Spec       interface{}            `json:"spec"`
	Assertions []*assertion.Assertion `json:"assertions"`
}

type Core struct {
	store dbprobe.ProbeDB
}

func NewCore(log *zap.SugaredLogger, db *dbpgx.Client) Core {
	return Core{
		store: dbprobe.NewProbeDB(log, db),
	}
}

// Set of error variables for CRUD operations.
var (
	ErrNotFoundProbe = errors.New("probe not found")
	ErrInvalidProbe  = errors.New("probe is not valid")
)

// Create inserts a new probe into the database.
func (c Core) Create(ctx context.Context, nt VigieTest, time time.Time) (VigieTest, error) {

	return nt, nil
}
