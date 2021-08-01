package scraper

type VideoID int

type Video struct {
    ID VideoID
    TweetID TweetID
    Filename string
    IsDownloaded bool
}
