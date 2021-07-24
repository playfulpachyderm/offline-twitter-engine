package persistence

import (
	"fmt"
	"errors"
	"os"
	"strings"

	"offline_twitter/scraper"
)

var NotInDatabase = errors.New("Not in database")

type ErrNotInDatabase struct {
	Table string
	Value interface{}
}
func (err ErrNotInDatabase) Error() string {
	return fmt.Sprintf("Not in database: %s %q", err.Table, err.Value)
}

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


func parse_users_file(data []byte) []scraper.UserHandle {
	users := strings.Split(string(data), "\n")
	ret := []scraper.UserHandle{}
	for _, u := range users {
		if u != "" {
			ret = append(ret, scraper.UserHandle(u))
		}
	}
	return ret
}
