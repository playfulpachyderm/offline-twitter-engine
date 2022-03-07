package scraper

import (
    "fmt"
    "sort"
    "path"
)

type VideoID int64

// TODO video-source-user: extract source user information (e.g., someone shares a video
// from someone else).

type Video struct {
    ID VideoID
    TweetID TweetID
    Width int
    Height int
    RemoteURL string
    LocalFilename string

    ThumbnailRemoteUrl string
    ThumbnailLocalPath string `db:"thumbnail_local_filename"`
    Duration int  // milliseconds
    ViewCount int

    IsDownloaded bool
    IsGif  bool
}

func ParseAPIVideo(apiVideo APIExtendedMedia, tweet_id TweetID) Video {
    variants := apiVideo.VideoInfo.Variants
    sort.Sort(variants)

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

    local_filename := fmt.Sprintf("%d.mp4", tweet_id)

    return Video{
        ID: VideoID(apiVideo.ID),
        TweetID: tweet_id,
        Width: apiVideo.OriginalInfo.Width,
        Height: apiVideo.OriginalInfo.Height,
        RemoteURL: variants[0].URL,
        LocalFilename: local_filename,

        ThumbnailRemoteUrl: apiVideo.MediaURLHttps,
        ThumbnailLocalPath: path.Base(apiVideo.MediaURLHttps),
        Duration: apiVideo.VideoInfo.Duration,
        ViewCount: view_count,

        IsDownloaded: false,
        IsGif: apiVideo.Type == "animated_gif",
    }
}
