package scraper

import (
    "path"
)

type ImageID int

type Image struct {
    ID ImageID
    TweetID TweetID
    Filename string
    RemoteURL string
    LocalFilename string
    IsDownloaded bool
}

func ParseAPIMedia(apiMedia APIMedia) Image {
    local_filename := path.Base(apiMedia.MediaURLHttps)
    return Image{
        Filename: apiMedia.MediaURLHttps,  // XXX filename
        RemoteURL: apiMedia.MediaURLHttps,
        LocalFilename: local_filename,
        IsDownloaded: false,
    }
}

func (img Image) FilenameWhenDownloaded() string {
    return path.Base(img.Filename)
}
