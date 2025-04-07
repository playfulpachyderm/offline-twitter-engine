package scraper

import (
	"errors"
)

var (
	END_OF_FEED           = errors.New("End of feed") //nolint:staticcheck // This isn't really an error
	ErrDoesntExist        = errors.New("Doesn't exist")
	ErrUserIsBanned       = errors.New("user is banned")
	ErrExternalApiError   = errors.New("Unexpected result from external API")
	ErrorIsTombstone      = errors.New("tweet is a tombstone")
	ErrRateLimited        = errors.New("rate limited")
	ErrLoginRequired      = errors.New("login required; please provide `--session <user>` flag")
	ErrSessionInvalidated = errors.New("session invalidated by Twitter")

	// These are not API errors, but network errors generally
	ErrNoInternet = errors.New("no internet connection")
)
