package scraper

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"strings"
	"time"

	"offline_twitter/terminal_utils"
)

type BroadcastID string

type Broadcast struct {
	ID                 BroadcastID
	TweetID            TweetID
	State              string
	Height             int
	Width              int
	Title              string
	ThumbnailWidth     int
	ThumbnailHeight    int
	ThumbnailRemoteUrl string
	ThumbnailLocalPath string
	BroadcasterID      UserID
	URL                string
	ShortURL           string
	Orientation        string
	Source             string
	MediaID            string
	MediaKey           string
}

func ParseAsBroadcast(card APICard) Broadcast {
	binding_values := card.BindingValues
	return Broadcast{
		ID: BroadcastID(binding_values.BroadcastID.StringValue),
		// TweetID: TweetID(bindi)
		State:              binding_values.BroadcastState.StringValue,
		Height:             int_or_panic(binding_values.BroadcastHeight.StringValue),
		Width:              int_or_panic(binding_values.BroadcastWidth.StringValue),
		Title:              binding_values.BroadcastTitle.StringValue,
		ThumbnailWidth:     binding_values.BroadcastThumbnailOriginal.ImageValue.Width,
		ThumbnailHeight:    binding_values.BroadcastThumbnailOriginal.ImageValue.Height,
		ThumbnailRemoteUrl: binding_values.BroadcastThumbnailOriginal.ImageValue.Url,
		ThumbnailLocalPath: get_prefixed_path(get_filename(binding_values.BroadcastThumbnailOriginal.ImageValue.Url)),
		URL:                binding_values.BroadcastURL.StringValue,
		ShortURL:           binding_values.BroadcastShortURL.StringValue,
		Orientation:        binding_values.BroadcastOrientation.StringValue,
		Source:             binding_values.BroadcastSource.StringValue,
		MediaID:            binding_values.BroadcastMediaID.StringValue,
		MediaKey:           binding_values.BroadcastMediaKey.StringValue,
	}
}
