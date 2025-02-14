package persistence

import (
	"errors"
)

// Downloader errors
var (
	ErrorDMCA           = errors.New("video is DMCAed, unable to download (HTTP 403 Forbidden)")
	ErrMediaDownload404 = errors.New("media download HTTP 404")

	// TODO: this DEFINITELY does not belong here
	ErrRequestTimeout = errors.New("request timed out")
)
