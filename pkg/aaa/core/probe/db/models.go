package db

import (
	"github.com/jackc/pgtype"
)

// ProbeDB represent the structure we need for moving data
// between the app and the database.
type ProbeDB struct {
	ID         int              `db:"id"`
	ProbeType  string           `db:"probe_type"`
	Frequency  int              `db:"frequency"`
	Interval   pgtype.Interval  `db:"interval"`
	LastRun    pgtype.Timestamp `db:"last_run"`
	Probe_data []byte           `db:"probe_data"`
}
