package scraper

import (
	"net/url"
	"path"
	"sort"
)

type VideoID int64

// TODO video-source-user: extract source user information (e.g., someone shares a video
// from someone else).

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

func get_filename(remote_url string) string {
	u, err := url.Parse(remote_url)
	if err != nil {
		panic(err)
	}
	return path.Base(u.Path)
}

func ParseAPIVideo(apiVideo APIExtendedMedia) Video {
	variants := apiVideo.VideoInfo.Variants
	sort.Sort(variants)
	video_remote_url := variants[0].URL

	var view_count int

	r := apiVideo.Ext.MediaStats.R

	switch r.(type) {
	case string:
		view_count = 0
	case map[string]interface{}:
		OK_entry, ok := r.(map[string]interface{})["ok"]
		if !ok {
			panic("No 'ok' value found in the R!")
		}
		view_count_str, ok := OK_entry.(map[string]interface{})["viewCount"]
		view_count = int_or_panic(view_count_str.(string))
		if !ok {
			panic("No 'viewCount' value found in the OK!")
		}
	}

	local_filename := get_prefixed_path(get_filename(video_remote_url))

	return Video{
		ID:            VideoID(apiVideo.ID),
		Width:         apiVideo.OriginalInfo.Width,
		Height:        apiVideo.OriginalInfo.Height,
		RemoteURL:     video_remote_url,
		LocalFilename: local_filename,

		ThumbnailRemoteUrl: apiVideo.MediaURLHttps,
		ThumbnailLocalPath: get_prefixed_path(path.Base(apiVideo.MediaURLHttps)),
		Duration:           apiVideo.VideoInfo.Duration,
		ViewCount:          view_count,

		IsDownloaded:    false,
		IsBlockedByDMCA: false,
		IsGeoblocked:    apiVideo.ExtMediaAvailability.Reason == "Geoblocked",
		IsGif:           apiVideo.Type == "animated_gif",
	}
}
