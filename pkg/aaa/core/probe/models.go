package probe

import (
	"unsafe"

	"github.com/vincoll/vigie/pkg/aaa/core/probe/db"
)

type User struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// NewUser contains information needed to create a new User.
type NewUser struct {
	Name string `json:"name" validate:"required"`
}

// UpdateUser defines what information may be provided to modify an existing
// User. All fields are optional so clients can send just the fields they want
// changed. It uses pointer fields so we can differentiate between a field that
// was not provided and a field that was provided as explicitly blank. Normally
// we do not want to use pointers to basic types but we make exceptions around
// marshalling/unmarshalling.
type UpdateUser struct {
	Name *string `json:"name"`
}

// =============================================================================

func toUser(dbUsr db.ProbeDB) User {
	pu := (*User)(unsafe.Pointer(&dbUsr))
	return *pu
}

func toUserSlice(dbUsrs []db.ProbeDB) []User {
	users := make([]User, len(dbUsrs))
	for i, dbUsr := range dbUsrs {
		users[i] = toUser(dbUsr)
	}
	return users
}
