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
	ID            VideoID
	TweetID       TweetID
	Width         int
	Height        int
	RemoteURL     string
	LocalFilename string

	ThumbnailRemoteUrl string
	ThumbnailLocalPath string `db:"thumbnail_local_filename"`
	Duration           int    // milliseconds
	Bitrate            int
	BitratesAvailable  []int
	ViewCount          int

	IsDownloaded    bool
	IsBlockedByDMCA bool
	IsGif           bool
}

func get_filename(remote_url string) string {
	u, err := url.Parse(remote_url)
	if err != nil {
		panic(err)
	}
	return path.Base(u.Path)
}

func ParseAPIVideo(apiVideo APIExtendedMedia, tweet_id TweetID) Video {
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

	bitrates := []int{}
	for _, v := range variants {
		if v.Bitrate == 0 {
			// Skip the .m3u8 one
			continue
		}
		bitrates = append([]int{v.Bitrate}, bitrates...)
	}

	return Video{
		ID:            VideoID(apiVideo.ID),
		TweetID:       tweet_id,
		Width:         apiVideo.OriginalInfo.Width,
		Height:        apiVideo.OriginalInfo.Height,
		RemoteURL:     video_remote_url,
		LocalFilename: local_filename,

		ThumbnailRemoteUrl: apiVideo.MediaURLHttps,
		ThumbnailLocalPath: get_prefixed_path(path.Base(apiVideo.MediaURLHttps)),
		Duration:           apiVideo.VideoInfo.Duration,
		BitratesAvailable:  bitrates,
		ViewCount:          view_count,

		IsDownloaded:    false,
		IsBlockedByDMCA: false,
		IsGif:           apiVideo.Type == "animated_gif",
	}
}
