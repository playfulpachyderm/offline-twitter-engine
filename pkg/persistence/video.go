package persistence

type VideoID int64

type Video struct {
	ID            VideoID     `db:"id"`
	TweetID       TweetID     `db:"tweet_id"`
	DMMessageID   DMMessageID `db:"chat_message_id"`
	Width         int         `db:"width"`
	Height        int         `db:"height"`
	RemoteURL     string      `db:"remote_url"`
	LocalFilename string      `db:"local_filename"`

	ThumbnailRemoteUrl string `db:"thumbnail_remote_url"`
	ThumbnailLocalPath string `db:"thumbnail_local_filename"`
	Duration           int    `db:"duration"` // milliseconds
	ViewCount          int    `db:"view_count"`

	IsDownloaded    bool `db:"is_downloaded"`
	IsBlockedByDMCA bool `db:"is_blocked_by_dmca"`
	IsGeoblocked    bool `db:"is_geoblocked"`
	IsGif           bool `db:"is_gif"`
}
