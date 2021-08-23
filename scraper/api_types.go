package scraper

import (
	"time"
	"strings"
	"encoding/json"
	"strconv"
)


type APIMedia struct {
	ID            int64  `json:"id_str,string"`
	MediaURLHttps string `json:"media_url_https"`
	Type          string `json:"type"`
	URL           string `json:"url"`
}

type SortableVariants []struct {
	Bitrate     int    `json:"bitrate,omitempty"`
	URL         string `json:"url"`
}
func (v SortableVariants) Len() int { return len(v) }
func (v SortableVariants) Swap(i, j int) { v[i], v[j] = v[j], v[i] }
func (v SortableVariants) Less(i, j int) bool { return v[i].Bitrate > v[j].Bitrate }

type APIExtendedMedia struct {
	ID            int64  `json:"id_str,string"`
	MediaURLHttps string `json:"media_url_https"`
	Type          string `json:"type"`
	VideoInfo     struct {
		Variants  SortableVariants `json:"variants"`
	} `json:"video_info"`
}

type APITweet struct {
	ID                int64  `json:"id_str,string"`
	ConversationID    int64  `json:"conversation_id_str,string"`
	CreatedAt         string `json:"created_at"`
	FavoriteCount     int    `json:"favorite_count"`
	FullText          string `json:"full_text"`
	Entities          struct {
		Hashtags []struct {
			Text string `json:"text"`
		} `json:"hashtags"`
		Media []APIMedia `json:"media"`
		URLs []struct {
			ExpandedURL string `json:"expanded_url"`
			URL         string `json:"url"`
		} `json:"urls"`
		Mentions []struct {
			UserName string `json:"screen_name"`
			UserID   int64  `json:"id_str,string"`
		} `json:"user_mentions"`
	} `json:"entities"`
	ExtendedEntities struct {
		Media []APIExtendedMedia `json:"media"`
	} `json:"extended_entities"`
	InReplyToStatusID    int64     `json:"in_reply_to_status_id_str,string"`
	InReplyToScreenName  string    `json:"in_reply_to_screen_name"`
	ReplyCount           int       `json:"reply_count"`
	RetweetCount         int       `json:"retweet_count"`
	QuoteCount           int       `json:"quote_count"`
	RetweetedStatusIDStr string    `json:"retweeted_status_id_str"`  // Can be empty string
	RetweetedStatusID    int64
	QuotedStatusIDStr    string    `json:"quoted_status_id_str"`     // Can be empty string
	QuotedStatusID       int64
	Time                 time.Time `json:"time"`
	UserID               int64     `json:"user_id_str,string"`
}

func (t *APITweet) NormalizeContent() {
	// Remove embedded links at the end of the text
	if len(t.Entities.URLs) == 1 {  // TODO: should this be `>= 1`, like below?
		url := t.Entities.URLs[0].URL
		if strings.Index(t.FullText, url) == len(t.FullText) - len(url) {
			t.FullText = t.FullText[0:len(t.FullText) - len(url)]  // Also strip the newline
		}
	}
	if len(t.Entities.Media) >= 1 {
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

	id, err := strconv.Atoi(t.QuotedStatusIDStr)
	if err == nil {
		t.QuotedStatusID = int64(id)
	}
	id, err = strconv.Atoi(t.RetweetedStatusIDStr)
	if err == nil {
		t.RetweetedStatusID = int64(id)
	}
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
	ID                   int64    `json:"id_str,string"`
	ListedCount          int      `json:"listed_count"`
	Name                 string   `json:"name"`
	Location             string   `json:"location"`
	PinnedTweetIdsStr    []string `json:"pinned_tweet_ids_str"`  // Dunno how to type-convert an array
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
			ID     int64   `json:"rest_id,string"`
			Legacy APIUser `json:"legacy"`
		} `json:"user"`
	} `json:"data"`
	Errors []struct {
		Message string  `json:"message"`
		Code    int     `json:"code"`
	} `json:"errors"`
}
func (u UserResponse) ConvertToAPIUser() APIUser {
	ret := u.Data.User.Legacy
	ret.ID = u.Data.User.ID
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

/**
 * Test for one case of end-of-feed.  Cursor increments on each request for some reason, but
 * there's no new content.  This seems to happen when there's a pinned tweet.
 *
 * In this case, we look for an "entries" object that has only cursors in it, and no tweets.
 */
func (t *TweetResponse) IsEndOfFeed() bool {
	entries := t.Timeline.Instructions[0].AddEntries.Entries
	if len(entries) > 2 {
		return false
	}
	for _, e := range entries {
		if !strings.Contains(e.EntryID, "cursor") {
			return false
		}
	}
	return true
}


func idstr_to_int(idstr string) int64 {
	id, err := strconv.Atoi(idstr)
	if err != nil {
		panic(err)
	}
	return int64(id)
}
