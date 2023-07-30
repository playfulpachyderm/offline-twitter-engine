package scraper

import (
	"fmt"
	"net/url"
	"path"
	"regexp"
)

type Url struct {
	TweetID            TweetID `db:"tweet_id"`
	Domain             string  `db:"domain"`
	Text               string  `db:"text"`
	ShortText          string  `db:"short_text"`
	Title              string  `db:"title"`
	Description        string  `db:"description"`
	ThumbnailWidth     int     `db:"thumbnail_width"`
	ThumbnailHeight    int     `db:"thumbnail_height"`
	ThumbnailRemoteUrl string  `db:"thumbnail_remote_url"`
	ThumbnailLocalPath string  `db:"thumbnail_local_path"`
	CreatorID          UserID  `db:"creator_id"`
	SiteID             UserID  `db:"site_id"`

	HasCard             bool `db:"has_card"`
	HasThumbnail        bool `db:"has_thumbnail"`
	IsContentDownloaded bool `db:"is_content_downloaded"`
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

func get_prefixed_path(p string) string {
	local_prefix_regex := regexp.MustCompile(`^[\w-]{2}`)
	local_prefix := local_prefix_regex.FindString(p)
	if len(local_prefix) != 2 {
		panic(fmt.Sprintf("Unable to extract a 2-letter prefix for filename %s", p))
	}
	return path.Join(local_prefix, p)
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

	return get_prefixed_path(
		fmt.Sprintf("%s_%s.%s", path.Base(u.Path), query_params["name"][0], query_params["format"][0]),
	)
}

/**
 * Given an URL, try to parse it as a tweet url.
 * The bool is an `is_ok` value; true if the parse was successful, false if it didn't match
 */
func TryParseTweetUrl(s string) (UserHandle, TweetID, bool) {
	parsed_url, err := url.Parse(s)
	if err != nil {
		return UserHandle(""), TweetID(0), false
	}

	if parsed_url.Host != "twitter.com" && parsed_url.Host != "mobile.twitter.com" {
		return UserHandle(""), TweetID(0), false
	}

	r := regexp.MustCompile(`^/(\w+)/status/(\d+)$`)
	matches := r.FindStringSubmatch(parsed_url.Path)
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
