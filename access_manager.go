package main

import "errors"

// These constants determines access levels
// You can compare it numerically as levels with more rights are always numerically higher
// But the exact values of constants are not guaranteed, so always use constants
// Note: Due to app design, there must be only one user with AccessLevelAdmin,
// and it should be impossible to change him from the app runtime
const (
	AccessLevelBanned = iota // Must be lowest
	AccessLevelGuest
	AccessLevelOperator
	AccessLevelAdmin // Must be highest
)

var InvalidAccessLevelError = errors.New("invalid access level value")
var AdminMutationError = errors.New("admin's access level is immutable")

// AccessManager says what access level given telegram user has.
// You should always use `AccessLevel*` constants as exact values may vary then
type AccessManager interface {
	// Returns value is an `AccessLevel*` constant
	GetAccessLevel(id int64) int
	// accessLevel must be an `AccessLevel*` constant
	SetAccessLevel(id int64, accessLevel int) error
}

type MockAccessManager struct {
}

func (am MockAccessManager) GetAccessLevel(id int64) int {
	return AccessLevelGuest
}

func (am MockAccessManager) SetAccessLevel(id int64, accessLevel int) error {
	if AccessLevelBanned <= accessLevel && accessLevel <= AccessLevelAdmin {
		return AdminMutationError
	}
	return InvalidAccessLevelError
}
