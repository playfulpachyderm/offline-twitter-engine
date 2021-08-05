package scraper

import (
    "path"
)

type ImageID int64

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
        ID: ImageID(apiMedia.ID),
        Filename: apiMedia.MediaURLHttps,  // TODO filename
        RemoteURL: apiMedia.MediaURLHttps,
        LocalFilename: local_filename,
        IsDownloaded: false,
    }
}

func (img Image) FilenameWhenDownloaded() string {
    return path.Base(img.Filename)
}
