package persistence

import (
	"errors"
	"os"
	"strings"

	"offline_twitter/scraper"
)


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
