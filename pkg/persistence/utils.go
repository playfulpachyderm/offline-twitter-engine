package persistence

import (
	"errors"
	"os"
)

var ErrNotInDatabase = errors.New("not in database")

// DUPE 1
func file_exists(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	} else if errors.Is(err, os.ErrNotExist) {
		return false
	} else {
		panic(err)
	}
}
