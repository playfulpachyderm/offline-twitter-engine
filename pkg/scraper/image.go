package scraper

import (
	"path"
)

type ImageID int64

type Image struct {
	ID            ImageID     `db:"id"`
	TweetID       TweetID     `db:"tweet_id"`
	DMMessageID   DMMessageID `db:"chat_message_id"`
	Width         int         `db:"width"`
	Height        int         `db:"height"`
	RemoteURL     string      `db:"remote_url"`
	LocalFilename string      `db:"local_filename"`
	IsDownloaded  bool        `db:"is_downloaded"`
}

func ParseAPIMedia(apiMedia APIMedia) Image {
	local_filename := get_prefixed_path(path.Base(apiMedia.MediaURLHttps))

	return Image{
		ID:            ImageID(apiMedia.ID),
		RemoteURL:     apiMedia.MediaURLHttps,
		Width:         apiMedia.OriginalInfo.Width,
		Height:        apiMedia.OriginalInfo.Height,
		LocalFilename: local_filename,
		IsDownloaded:  false,
	}
}
