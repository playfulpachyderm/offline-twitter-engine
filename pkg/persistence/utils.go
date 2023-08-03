package persistence

import (
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"
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

// https://stackoverflow.com/questions/56616196/how-to-convert-camel-case-string-to-snake-case#56616250
func ToSnakeCase(str string) string {
	snake := regexp.MustCompile("(.)_?([A-Z][a-z]+)").ReplaceAllString(str, "${1}_${2}")
	snake = regexp.MustCompile("([a-z0-9])_?([A-Z])").ReplaceAllString(snake, "${1}_${2}")
	return strings.ToLower(snake)
}
