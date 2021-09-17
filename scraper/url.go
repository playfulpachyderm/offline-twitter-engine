package scraper

import (
	"fmt"
	"path"
	"net/url"
)

type Url struct {
	Domain string
	Text string
	Title string
	Description string
	ThumbnailRemoteUrl string
	ThumbnailLocalPath string
	CreatorID UserID
	SiteID UserID

	IsContentDownloaded bool
}

func ParseAPIUrlCard(apiCard APICard) Url {
	values := apiCard.BindingValues
	return Url{
		Domain: values.Domain.Value,
		Title: values.Title.Value,
		Description: values.Description.Value,
		ThumbnailRemoteUrl: values.Thumbnail.ImageValue.Url,
		ThumbnailLocalPath: get_thumbnail_local_path(values.Thumbnail.ImageValue.Url),
		CreatorID: UserID(values.Creator.UserValue.Value),
		SiteID: UserID(values.Site.UserValue.Value),
		IsContentDownloaded: false,
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
