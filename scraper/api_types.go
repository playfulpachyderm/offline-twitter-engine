package scraper

import (
	"time"
	"strings"
	"encoding/json"
)

type APITweet struct {
	ID                string `json:"id_str"`
	ConversationIDStr string `json:"conversation_id_str"`
	CreatedAt         string `json:"created_at"`
	FavoriteCount     int    `json:"favorite_count"`
	FullText          string `json:"full_text"`
	Entities          struct {
		Hashtags []struct {
			Text string `json:"text"`
		} `json:"hashtags"`
		Media []struct {
			MediaURLHttps string `json:"media_url_https"`
			Type          string `json:"type"`
			URL           string `json:"url"`
		} `json:"media"`
		URLs []struct {
			ExpandedURL string `json:"expanded_url"`
			URL         string `json:"url"`
		} `json:"urls"`
		Mentions []struct {
			UserName string `json:"screen_name"`
			UserID   string `json:"id_str"`
		} `json:"user_mentions"`
	} `json:"entities"`
	ExtendedEntities struct {
		Media []struct {
			IDStr         string `json:"id_str"`
			MediaURLHttps string `json:"media_url_https"`
			Type          string `json:"type"`
			VideoInfo     struct {
				Variants []struct {
					Bitrate int    `json:"bitrate,omitempty"`
					URL     string `json:"url"`
				} `json:"variants"`
			} `json:"video_info"`
		} `json:"media"`
	} `json:"extended_entities"`
	InReplyToStatusIDStr string    `json:"in_reply_to_status_id_str"`
	InReplyToScreenName  string    `json:"in_reply_to_screen_name"`
	ReplyCount           int       `json:"reply_count"`
	RetweetCount         int       `json:"retweet_count"`
	QuoteCount           int       `json:"quote_count"`
	RetweetedStatusIDStr string    `json:"retweeted_status_id_str"`
	QuotedStatusIDStr    string    `json:"quoted_status_id_str"`
	Time                 time.Time `json:"time"`
	UserIDStr            string    `json:"user_id_str"`
}

func (t *APITweet) NormalizeContent() {
	// Remove embedded links at the end of the text
	if len(t.Entities.URLs) == 1 {
		url := t.Entities.URLs[0].URL
		if strings.Index(t.FullText, url) == len(t.FullText) - len(url) {
			t.FullText = t.FullText[0:len(t.FullText) - len(url)]  // Also strip the newline
		}
	}
	if len(t.Entities.Media) == 1 {
		url := t.Entities.Media[0].URL
		if strings.Index(t.FullText, url) == len(t.FullText) - len(url) {
			t.FullText = t.FullText[0:len(t.FullText) - len(url)]  // Also strip the trailing space
		}
	}
	// Remove leading `@username` for replies
	if t.InReplyToScreenName != "" {
		if strings.Index(t.FullText, "@" + t.InReplyToScreenName) == 0 {
			t.FullText = t.FullText[len(t.InReplyToScreenName) + 1:]  // `@`, username, space
		}
	}
	t.FullText = strings.TrimSpace(t.FullText)
}

func (t APITweet) String() string {
	data, err := json.Marshal(t)
	if err != nil {
		panic(err)
	}
	return string(data)
}


type APIUser struct {
	CreatedAt   string `json:"created_at"`
	Description string `json:"description"`
	Entities    struct {
		URL struct {
			Urls []struct {
				ExpandedURL string `json:"expanded_url"`
			} `json:"urls"`
		} `json:"url"`
	} `json:"entities"`
	FavouritesCount      int      `json:"favourites_count"`
	FollowersCount       int      `json:"followers_count"`
	FriendsCount         int      `json:"friends_count"`
	IDStr                string   `json:"id_str"`
	ListedCount          int      `json:"listed_count"`
	Name                 string   `json:"name"`
	Location             string   `json:"location"`
	PinnedTweetIdsStr    []string `json:"pinned_tweet_ids_str"`
	ProfileBannerURL     string   `json:"profile_banner_url"`
	ProfileImageURLHTTPS string   `json:"profile_image_url_https"`
	Protected            bool     `json:"protected"`
	ScreenName           string   `json:"screen_name"`
	StatusesCount        int      `json:"statuses_count"`
	Verified             bool     `json:"verified"`
}


type UserResponse struct {
	Data struct {
		User struct {
			ID     string  `json:"rest_id"`
			Legacy APIUser `json:"legacy"`
		} `json:"user"`
	} `json:"data"`
}
func (u UserResponse) ConvertToAPIUser() APIUser {
	ret := u.Data.User.Legacy
	ret.IDStr = u.Data.User.ID
	return ret
}

type TweetResponse struct {
	GlobalObjects struct {
		Tweets map[string]APITweet `json:"tweets"`
		Users  map[string]APIUser  `json:"users"`
	} `json:"globalObjects"`
	Timeline struct {
		Instructions []struct {
			AddEntries struct {
				Entries []struct {
					EntryID string `json:"entryId"`
					Content struct {
						Operation struct {
							Cursor struct {
								Value string `json:"value"`
							} `json:"cursor"`
						} `json:"operation"`
					} `json:"content"`
				} `json:"entries"`
			} `json:"addEntries"`
		} `json:"instructions"`
	} `json:"timeline"`
}

func (t *TweetResponse) GetCursor() string {
	entries := t.Timeline.Instructions[0].AddEntries.Entries
	last_entry := entries[len(entries) - 1]
	if strings.Contains(last_entry.EntryID, "cursor") {
		return last_entry.Content.Operation.Cursor.Value
	}
	return ""
}
