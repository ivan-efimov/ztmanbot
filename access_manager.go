package main

import (
	"encoding/json"
	"errors"
	"io/ioutil"
)

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

func ValidLevelToSet(level int) bool {
	return AccessLevelBanned <= level && level < AccessLevelAdmin
}

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

type AccessManagerWithFileStorage struct {
	adminId   int64
	accessMap map[int64]int
	filepath  string
}

// Returns nil if an error occurred
func NewAccessManagerWithFileStorage(adminId int64, filepath string) (*AccessManagerWithFileStorage, error) {
	fileData, err := ioutil.ReadFile(filepath)
	if err != nil {
		return nil, err
	}
	m := make(map[int64]int, 0)
	err = json.Unmarshal(fileData, &m)
	if err != nil {
		return nil, err
	}
	for _, level := range m {
		if !ValidLevelToSet(level) {
			return nil, errors.New("file corrupted")
		}
	}
	delete(m, adminId)
	return &AccessManagerWithFileStorage{
		adminId:   adminId,
		accessMap: m,
		filepath:  filepath,
	}, nil
}

func (a AccessManagerWithFileStorage) GetAccessLevel(id int64) int {
	if id == a.adminId {
		return AccessLevelAdmin
	}
	level, found := a.accessMap[id]
	if !found {
		return AccessLevelGuest
	}
	return level
}

func (a AccessManagerWithFileStorage) SetAccessLevel(id int64, accessLevel int) error {
	if id == a.adminId {
		return AdminMutationError
	}
	if accessLevel < AccessLevelBanned || accessLevel >= AccessLevelAdmin {
		return InvalidAccessLevelError
	}
	// AccessLevelGuest is default value
	if accessLevel == AccessLevelGuest {
		delete(a.accessMap, id)
	} else {
		a.accessMap[id] = accessLevel
	}
	return a.commit()
}

func (a AccessManagerWithFileStorage) commit() error {
	fileData, err := json.Marshal(a.accessMap)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(a.filepath, fileData, 664)
}
