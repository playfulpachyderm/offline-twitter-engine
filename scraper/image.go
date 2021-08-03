package scraper

import (
    "path"
)

type ImageID int

type Image struct {
    ID ImageID
    TweetID TweetID
    Filename string
    IsDownloaded bool
}

func (img Image) FilenameWhenDownloaded() string {
    return path.Base(img.Filename)
}
