package scraper

import (
	"errors"
)

var (
	END_OF_FEED        = errors.New("End of feed")
	ErrDoesntExist     = errors.New("Doesn't exist")
	EXTERNAL_API_ERROR = errors.New("Unexpected result from external API")
	ErrorIsTombstone   = errors.New("tweet is a tombstone")
	ErrRateLimited     = errors.New("rate limited")
	ErrorDMCA          = errors.New("video is DMCAed, unable to download (HTTP 403 Forbidden)")
	ErrRequestTimeout  = errors.New("request timed out")
)
