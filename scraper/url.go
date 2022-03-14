package scraper

import (
	"fmt"
	"net/url"
	"path"
	"regexp"
)

type Url struct {
	TweetID TweetID

	Domain             string
	Text               string
	ShortText          string
	Title              string
	Description        string
	ThumbnailWidth     int
	ThumbnailHeight    int
	ThumbnailRemoteUrl string
	ThumbnailLocalPath string
	CreatorID          UserID
	SiteID             UserID

	HasCard             bool
	HasThumbnail        bool
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
	if u.RawQuery == "" {
		return path.Base(u.Path)
	}
	query_params, err := url.ParseQuery(u.RawQuery)
	if err != nil {
		panic(err)
	}

	return fmt.Sprintf("%s_%s.%s", path.Base(u.Path), query_params["name"][0], query_params["format"][0])
}

/**
 * Given an URL, try to parse it as a tweet url.
 * The bool is an `is_ok` value; true if the parse was successful, false if it didn't match
 */
func TryParseTweetUrl(url string) (UserHandle, TweetID, bool) {
	r := regexp.MustCompile(`^https://twitter.com/(\w+)/status/(\d+)(?:\?.*)?$`)
	matches := r.FindStringSubmatch(url)
	if matches == nil {
		return UserHandle(""), TweetID(0), false
	}
	if len(matches) != 3 { // matches[0] is the full string
		panic(matches)
	}
	return UserHandle(matches[1]), TweetID(int_or_panic(matches[2])), true
}

/**
 * Given a tweet URL, return the corresponding user handle.
 * If tweet url is not valid, return an error.
 */
func ParseHandleFromTweetUrl(tweet_url string) (UserHandle, error) {
	short_url_regex := regexp.MustCompile(`^https://t.co/\w{5,20}$`)
	if short_url_regex.MatchString(tweet_url) {
		tweet_url = ExpandShortUrl(tweet_url)
	}

	ret, _, is_ok := TryParseTweetUrl(tweet_url)
	if !is_ok {
		return "", fmt.Errorf("Invalid tweet url: %s", tweet_url)
	}
	return ret, nil
}
