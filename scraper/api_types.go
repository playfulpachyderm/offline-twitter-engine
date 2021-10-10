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

type APICard struct {
	Name          string `json:"name"`
	ShortenedUrl  string `json:"url"`
	BindingValues struct {
		Domain struct {
			Value string `json:"string_value"`
		} `json:"domain"`
		Creator struct {
			UserValue struct {
				Value int64 `json:"id_str,string"`
			} `json:"user_value"`
		} `json:"creator"`
		Site struct {
			UserValue struct {
				Value int64 `json:"id_str,string"`
			} `json:"user_value"`
		} `json:"site"`
		Title struct {
			Value string `json:"string_value"`
		} `json:"title"`
		Description struct {
			Value string `json:"string_value"`
		} `json:"description"`
		Thumbnail struct {
			ImageValue struct {
				Url string `json:"url"`
			} `json:"image_value"`
		} `json:"thumbnail_image_large"`
		PlayerImage struct {
			ImageValue struct {
				Url string `json:"url"`
			} `json:"image_value"`
		} `json:"player_image_large"`
	} `json:"binding_values"`
}

type APITweet struct {
	ID                int64  `json:"id_str,string"`
	ConversationID    int64  `json:"conversation_id_str,string"`
	CreatedAt         string `json:"created_at"`
	FavoriteCount     int    `json:"favorite_count"`
	FullText          string `json:"full_text"`
	DisplayTextRange  []int  `json:"display_text_range"`
	Entities          struct {
		Hashtags []struct {
			Text string `json:"text"`
		} `json:"hashtags"`
		Media []APIMedia `json:"media"`
		URLs []struct {
			ExpandedURL  string `json:"expanded_url"`
			ShortenedUrl string `json:"url"`
		} `json:"urls"`
		Mentions []struct {
			UserName string `json:"screen_name"`
			UserID   int64  `json:"id_str,string"`
		} `json:"user_mentions"`
		ReplyMentions string  // The leading part of the text which is cut off by "DisplayTextRange"
	} `json:"entities"`
	ExtendedEntities struct {
		Media []APIExtendedMedia `json:"media"`
	} `json:"extended_entities"`
	InReplyToStatusID     int64     `json:"in_reply_to_status_id_str,string"`
	InReplyToScreenName   string    `json:"in_reply_to_screen_name"`
	ReplyCount            int       `json:"reply_count"`
	RetweetCount          int       `json:"retweet_count"`
	QuoteCount            int       `json:"quote_count"`
	RetweetedStatusIDStr  string    `json:"retweeted_status_id_str"`  // Can be empty string
	RetweetedStatusID     int64
	QuotedStatusIDStr     string    `json:"quoted_status_id_str"`     // Can be empty string
	QuotedStatusID        int64
	QuotedStatusPermalink struct {
		URL         string `json:"url"`
		ExpandedURL string `json:"expanded"`
	} `json:"quoted_status_permalink"`
	Time                  time.Time `json:"time"`
	UserID                int64     `json:"user_id_str,string"`
	Card                  APICard   `json:"card"`
}

func (t *APITweet) NormalizeContent() {
	id, err := strconv.Atoi(t.QuotedStatusIDStr)
	if err == nil {
		t.QuotedStatusID = int64(id)
	}
	id, err = strconv.Atoi(t.RetweetedStatusIDStr)
	if err == nil {
		t.RetweetedStatusID = int64(id)
	}

	if (len(t.DisplayTextRange) == 2) {
		t.Entities.ReplyMentions = strings.TrimSpace(string([]rune(t.FullText)[0:t.DisplayTextRange[0]]))
		t.FullText = string([]rune(t.FullText)[t.DisplayTextRange[0]:t.DisplayTextRange[1]])
	}

	// Handle pasted tweet links that turn into quote tweets but still have a link in them
	if t.QuotedStatusID != 0 {
		for _, url := range t.Entities.URLs {
			if url.ShortenedUrl == t.QuotedStatusPermalink.URL {
				t.FullText = strings.ReplaceAll(t.FullText, url.ShortenedUrl, "")
			}
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
