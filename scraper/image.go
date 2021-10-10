package scraper

import (
    "path"
)

type ImageID int64

type Image struct {
    ID ImageID
    TweetID TweetID
    Width int
    Height int
    RemoteURL string
    LocalFilename string
    IsDownloaded bool
}

func ParseAPIMedia(apiMedia APIMedia) Image {
    local_filename := path.Base(apiMedia.MediaURLHttps)
    return Image{
        ID: ImageID(apiMedia.ID),
        RemoteURL: apiMedia.MediaURLHttps,
        Width: apiMedia.OriginalInfo.Width,
        Height: apiMedia.OriginalInfo.Height,
        LocalFilename: local_filename,
        IsDownloaded: false,
    }
}
