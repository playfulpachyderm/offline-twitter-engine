package scraper

import (
    "fmt"
)

type VideoID int

type Video struct {
    ID VideoID
    TweetID TweetID
    Filename string
    IsDownloaded bool
}

func (v Video) FilenameWhenDownloaded() string {
    return fmt.Sprintf("%s.mp4", v.TweetID)
}