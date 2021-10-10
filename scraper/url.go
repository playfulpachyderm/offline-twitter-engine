package scraper

import (
	"fmt"
	"path"
	"net/url"
)

type Url struct {
	TweetID TweetID

	Domain string
	Text string
	Title string
	Description string
	ThumbnailWidth int
	ThumbnailHeight int
	ThumbnailRemoteUrl string
	ThumbnailLocalPath string
	CreatorID UserID
	SiteID UserID

	HasCard bool
	HasThumbnail bool
	IsContentDownloaded bool
}

func ParseAPIUrlCard(apiCard APICard) Url {
	values := apiCard.BindingValues
	ret := Url{}
	ret.HasCard = true

	ret.Domain = values.Domain.Value
	ret.Title = values.Title.Value
	ret.Description = values.Description.Value
	ret.IsContentDownloaded = false
	ret.CreatorID = UserID(values.Creator.UserValue.Value)
	ret.SiteID = UserID(values.Site.UserValue.Value)

	var thumbnail_url string

	if apiCard.Name == "summary_large_image" || apiCard.Name == "summary" {
		thumbnail_url = values.Thumbnail.ImageValue.Url
	} else if apiCard.Name == "player" {
		thumbnail_url = values.PlayerImage.ImageValue.Url
	} else {
		panic("Unknown card type: " + apiCard.Name)
	}

	if thumbnail_url != "" {
		ret.HasThumbnail = true
		ret.ThumbnailRemoteUrl = thumbnail_url
		ret.ThumbnailLocalPath = get_thumbnail_local_path(thumbnail_url)
		ret.ThumbnailWidth = values.Thumbnail.ImageValue.Width
		ret.ThumbnailHeight = values.Thumbnail.ImageValue.Height
	}

	return ret
}

func get_thumbnail_local_path(remote_url string) string {
    u, err := url.Parse(remote_url)
    if err != nil {
        panic(err)
    }
    query_params, err := url.ParseQuery(u.RawQuery)
    if err != nil {
        panic(err)
    }

    return fmt.Sprintf("%s_%s.%s", path.Base(u.Path), query_params["name"][0], query_params["format"][0])
}
