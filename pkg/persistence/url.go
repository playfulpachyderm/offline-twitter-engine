package persistence

import (
	"net/url"
)

type Url struct {
	TweetID            TweetID     `db:"tweet_id"`
	DMMessageID        DMMessageID `db:"chat_message_id"`
	Domain             string      `db:"domain"`
	Text               string      `db:"text"`
	ShortText          string      `db:"short_text"`
	Title              string      `db:"title"`
	Description        string      `db:"description"`
	ThumbnailWidth     int         `db:"thumbnail_width"`
	ThumbnailHeight    int         `db:"thumbnail_height"`
	ThumbnailRemoteUrl string      `db:"thumbnail_remote_url"`
	ThumbnailLocalPath string      `db:"thumbnail_local_path"`
	CreatorID          UserID      `db:"creator_id"`
	SiteID             UserID      `db:"site_id"`

	HasCard             bool `db:"has_card"`
	HasThumbnail        bool `db:"has_thumbnail"`
	IsContentDownloaded bool `db:"is_content_downloaded"`
}

// TODO: view-layer
// - view helpers should go in a view layer

func (u Url) GetDomain() string {
	if u.Domain != "" {
		return u.Domain
	}
	urlstruct, err := url.Parse(u.Text)
	if err != nil {
		panic(err)
	}
	return urlstruct.Host
}
