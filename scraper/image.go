package scraper

type ImageID int

type Image struct {
    ID ImageID
    TweetID TweetID
    Filename string
    IsDownloaded bool
}
