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
	ThumbnailRemoteUrl string
	ThumbnailLocalPath string
	CreatorID UserID
	SiteID UserID

	HasCard bool
	IsContentDownloaded bool
}

func ParseAPIUrlCard(apiCard APICard) Url {
	values := apiCard.BindingValues
	if apiCard.Name == "summary_large_image" || apiCard.Name == "summary" {
		return Url{
			Domain: values.Domain.Value,
			Title: values.Title.Value,
			Description: values.Description.Value,
			ThumbnailRemoteUrl: values.Thumbnail.ImageValue.Url,
			ThumbnailLocalPath: get_thumbnail_local_path(values.Thumbnail.ImageValue.Url),
			CreatorID: UserID(values.Creator.UserValue.Value),
			SiteID: UserID(values.Site.UserValue.Value),
			HasCard: true,
			IsContentDownloaded: false,
		}
	} else if apiCard.Name == "player" {
		return Url{
			Domain: values.Domain.Value,
			Title: values.Title.Value,
			Description: values.Description.Value,
			ThumbnailRemoteUrl: values.PlayerImage.ImageValue.Url,
			ThumbnailLocalPath: get_thumbnail_local_path(values.PlayerImage.ImageValue.Url),
			CreatorID: UserID(values.Creator.UserValue.Value),
			SiteID: UserID(values.Site.UserValue.Value),
			HasCard: true,
			IsContentDownloaded: false,
		}
	} else {
		panic("Unknown card type: " + apiCard.Name)
	}
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
