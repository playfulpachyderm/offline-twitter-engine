package scraper

import (
    "fmt"
    "sort"
)

type VideoID int64

// TODO video-source-user: extract source user information (e.g., someone shares a video
// from someone else).

type Video struct {
    ID VideoID
    TweetID TweetID
    Filename string  // TODO video-filename: delete when it all works
    RemoteURL string
    LocalFilename string
    IsDownloaded bool
}

func ParseAPIVideo(apiVideo APIExtendedMedia, tweet_id TweetID) Video {
    variants := apiVideo.VideoInfo.Variants
    sort.Sort(variants)

    local_filename := fmt.Sprintf("%d.mp4", tweet_id)

    return Video{
        ID: VideoID(apiVideo.ID),
        TweetID: tweet_id,
        Filename: variants[0].URL,
        RemoteURL: variants[0].URL,
        LocalFilename: local_filename,
        IsDownloaded: false,
    }
}

func (v Video) FilenameWhenDownloaded() string {  // TODO video-filename: delete whole method and associated test
    return fmt.Sprintf("%d.mp4", v.TweetID)
}
