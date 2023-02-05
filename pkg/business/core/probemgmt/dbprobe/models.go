package dbprobe

import (
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

// ProbeTable represent the structure we need for moving data
// between the app and the database.
type ProbeTable struct {
	ID         uuid.UUID        `db:"id"`
	ProbeType  string           `db:"probe_type"`
	Frequency  int              `db:"frequency"`
	Interval   pgtype.Interval  `db:"interval"`
	LastRun    pgtype.Timestamp `db:"last_run"`
	Probe_data []byte           `db:"probe_data"`
	Probe_json []byte           `db:"probe_json"`
}
