package scraper

import (
	"encoding/json"
	"fmt"
	"html"
	"sort"
	"strconv"
	"strings"
	"time"
)

type APIMedia struct {
	ID            int64  `json:"id_str,string"`
	MediaURLHttps string `json:"media_url_https"`
	Type          string `json:"type"`
	URL           string `json:"url"`
	OriginalInfo  struct {
		Width  int `json:"width"`
		Height int `json:"height"`
	} `json:"original_info"`
}

type SortableVariants []struct {
	Bitrate int    `json:"bitrate,omitempty"`
	URL     string `json:"url"`
}

func (v SortableVariants) Len() int           { return len(v) }
func (v SortableVariants) Swap(i, j int)      { v[i], v[j] = v[j], v[i] }
func (v SortableVariants) Less(i, j int) bool { return v[i].Bitrate > v[j].Bitrate }

type APIExtendedMedia struct {
	ID            int64  `json:"id_str,string"`
	MediaURLHttps string `json:"media_url_https"`
	Type          string `json:"type"`
	VideoInfo     struct {
		Variants SortableVariants `json:"variants"`
		Duration int              `json:"duration_millis"`
	} `json:"video_info"`
	ExtMediaAvailability struct {
		Status string `json:"status"`
		Reason string `json:"reason"`
	} `json:"ext_media_availability"`
	OriginalInfo struct {
		Width  int `json:"width"`
		Height int `json:"height"`
	} `json:"original_info"`
	Ext struct {
		MediaStats struct {
			R interface{} `json:"r"`
		} `json:"mediaStats"`
	} `json:"ext"`
	URL string `json:"url"` // For DM videos
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
				Url    string `json:"url"`
				Width  int    `json:"width"`
				Height int    `json:"height"`
			} `json:"image_value"`
		} `json:"thumbnail_image_large"`
		PlayerImage struct {
			ImageValue struct {
				Url string `json:"url"`
			} `json:"image_value"`
		} `json:"player_image_large"`

		// For polls
		Choice1 struct {
			StringValue string `json:"string_value"`
		} `json:"choice1_label"`
		Choice2 struct {
			StringValue string `json:"string_value"`
		} `json:"choice2_label"`
		Choice3 struct {
			StringValue string `json:"string_value"`
		} `json:"choice3_label"`
		Choice4 struct {
			StringValue string `json:"string_value"`
		} `json:"choice4_label"`

		Choice1_Count struct {
			StringValue string `json:"string_value"`
		} `json:"choice1_count"`
		Choice2_Count struct {
			StringValue string `json:"string_value"`
		} `json:"choice2_count"`
		Choice3_Count struct {
			StringValue string `json:"string_value"`
		} `json:"choice3_count"`
		Choice4_Count struct {
			StringValue string `json:"string_value"`
		} `json:"choice4_count"`

		EndDatetimeUTC struct {
			StringValue string `json:"string_value"`
		} `json:"end_datetime_utc"`
		CountsAreFinal struct {
			BooleanValue bool `json:"boolean_value"`
		} `json:"counts_are_final"`
		DurationMinutes struct {
			StringValue string `json:"string_value"`
		} `json:"duration_minutes"`
		LastUpdatedAt struct {
			StringValue string `json:"string_value"`
		} `json:"last_updated_datetime_utc"`

		// For Spaces
		ID struct {
			StringValue string `json:"string_value"`
		} `json:"id"`
	} `json:"binding_values"`
}

type APITweet struct {
	ID               int64  `json:"id_str,string"`
	ConversationID   int64  `json:"conversation_id_str,string"`
	CreatedAt        string `json:"created_at"`
	FavoriteCount    int    `json:"favorite_count"`
	FullText         string `json:"full_text"`
	DisplayTextRange []int  `json:"display_text_range"`
	Entities         struct {
		Hashtags []struct {
			Text string `json:"text"`
		} `json:"hashtags"`
		Media []APIMedia `json:"media"`
		URLs  []struct {
			ExpandedURL  string `json:"expanded_url"`
			ShortenedUrl string `json:"url"`
		} `json:"urls"`
		Mentions []struct {
			UserName string `json:"screen_name"`
			UserID   int64  `json:"id_str,string"`
		} `json:"user_mentions"`
		ReplyMentions string // The leading part of the text which is cut off by "DisplayTextRange"
	} `json:"entities"`
	ExtendedEntities struct {
		Media []APIExtendedMedia `json:"media"`
	} `json:"extended_entities"`
	InReplyToStatusID     int64  `json:"in_reply_to_status_id_str,string"`
	InReplyToUserID       int64  `json:"in_reply_to_user_id_str,string"`
	InReplyToScreenName   string `json:"in_reply_to_screen_name"`
	ReplyCount            int    `json:"reply_count"`
	RetweetCount          int    `json:"retweet_count"`
	QuoteCount            int    `json:"quote_count"`
	RetweetedStatusIDStr  string `json:"retweeted_status_id_str"` // Can be empty string
	RetweetedStatusID     int64
	QuotedStatusIDStr     string `json:"quoted_status_id_str"` // Can be empty string
	QuotedStatusID        int64
	QuotedStatusPermalink struct {
		ShortURL    string `json:"url"`
		ExpandedURL string `json:"expanded"`
	} `json:"quoted_status_permalink"`
	Time          time.Time `json:"time"`
	UserID        int64     `json:"user_id_str,string"`
	UserHandle    string
	Card          APICard `json:"card"`
	TombstoneText string
	IsExpandable  bool
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

	if len(t.DisplayTextRange) == 2 {
		t.Entities.ReplyMentions = strings.TrimSpace(string([]rune(t.FullText)[0:t.DisplayTextRange[0]]))
		t.FullText = string([]rune(t.FullText)[t.DisplayTextRange[0]:t.DisplayTextRange[1]])
	}

	// Handle short links showing up at ends of tweets
	for _, url := range t.Entities.URLs {
		index := strings.Index(t.FullText, url.ShortenedUrl)
		if index < 0 {
			// It's not in the text
			continue
		}
		if index == (len(t.FullText) - len(url.ShortenedUrl)) {
			t.FullText = strings.TrimSpace(t.FullText[0:index])
		}
	}

	// Handle pasted tweet links that turn into quote tweets but still have a link in them
	// This is a separate case from above because we want it gone even if it's in the middle of the tweet
	if t.QuotedStatusID != 0 {
		for _, url := range t.Entities.URLs {
			if url.ShortenedUrl == t.QuotedStatusPermalink.ShortURL {
				t.FullText = strings.ReplaceAll(t.FullText, url.ShortenedUrl, "")
			}
		}
	}
	t.FullText = html.UnescapeString(t.FullText)
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
	PinnedTweetIdsStr    []string `json:"pinned_tweet_ids_str"` // Dunno how to type-convert an array
	ProfileBannerURL     string   `json:"profile_banner_url"`
	ProfileImageURLHTTPS string   `json:"profile_image_url_https"`
	Protected            bool     `json:"protected"`
	ScreenName           string   `json:"screen_name"`
	StatusesCount        int      `json:"statuses_count"`
	Verified             bool     `json:"verified"`
	IsBanned             bool
	DoesntExist          bool
}

type APINotification struct {
	ID          string `json:"id"`
	TimestampMs int64  `json:"timestampMs,string"`
	Message     struct {
		Text     string `json:"text"`
		Entities []struct {
			FromIndex int `json:"fromIndex"`
			ToIndex   int `json:"toIndex"`
			Ref       struct {
				User struct {
					ID int `json:"id,string"`
				} `json:"user"`
			} `json:"ref"`
		} `json:"entities"`
	} `json:"message"`
	Template struct {
		AggregateUserActionsV1 struct {
			TargetObjects []struct {
				Tweet struct {
					ID int `json:"id,string"`
				} `json:"tweet"`
			} `json:"targetObjects"`
			FromUsers []struct {
				User struct {
					ID int `json:"id,string"`
				} `json:"user"`
			} `json:"fromUsers"`
		} `json:"aggregateUserActionsV1"`
	} `json:"template"`
}

type UserResponse struct {
	Data struct {
		User struct {
			Result struct {
				ID                 int64   `json:"rest_id,string"`
				Legacy             APIUser `json:"legacy"`
				IsBlueVerified     bool    `json:"is_blue_verified"`
				UnavailableMessage struct {
					Text string `json:"text"`
				} `json:"unavailable_message"`
				Reason string `json:"reason"`
			} `json:"result"`
		} `json:"user"`
	} `json:"data"`
	Errors []struct {
		Message string `json:"message"`
		Name    string `json:"name"`
		Code    int    `json:"code"`
	} `json:"errors"`
}

func (u UserResponse) ConvertToAPIUser() APIUser {
	ret := u.Data.User.Result.Legacy
	ret.ID = u.Data.User.Result.ID
	ret.Verified = u.Data.User.Result.IsBlueVerified

	// Banned users
	for _, api_error := range u.Errors {
		if api_error.Message == "Authorization: User has been suspended. (63)" {
			ret.IsBanned = true
		} else if api_error.Name == "NotFoundError" {
			ret.DoesntExist = true
		} else {
			panic(fmt.Errorf("Unknown api error %q:\n  %w", api_error.Message, EXTERNAL_API_ERROR))
		}
	}

	// Banned users, new version
	if u.Data.User.Result.Reason == "Suspended" {
		ret.IsBanned = true
	}

	// Deleted users
	if ret.ID == 0 && ret.ScreenName == "" && u.Data.User.Result.Reason != "Suspended" {
		ret.DoesntExist = true
	}

	return ret
}

type Entry struct {
	EntryID   string `json:"entryId"`
	SortIndex int64  `json:"sortIndex,string"`
	Content   struct {
		Item struct {
			Content struct {
				Tombstone struct {
					TombstoneInfo struct {
						RichText struct {
							Text string `json:"text"`
						} `json:"richText"`
					} `json:"tombstoneInfo"`
				} `json:"tombstone"`
				Tweet struct {
					ID int64 `json:"id,string"`
				} `json:"tweet"`
				Notification struct {
					ID           string     `json:"id"`
					FromUsers    Int64Slice `json:"fromUsers"`
					TargetTweets Int64Slice `json:"targetTweets"`
				} `json:"notification"`
			} `json:"content"`
			ClientEventInfo struct {
				Element string `json:"element"`
			} `json:"clientEventInfo"`
		} `json:"item"`
		Operation struct {
			Cursor struct {
				Value string `json:"value"`
			} `json:"cursor"`
		} `json:"operation"`
	} `json:"content"`
}

func (e Entry) GetTombstoneText() string {
	return e.Content.Item.Content.Tombstone.TombstoneInfo.RichText.Text
}

type SortableEntries []Entry

func (e SortableEntries) Len() int           { return len(e) }
func (e SortableEntries) Swap(i, j int)      { e[i], e[j] = e[j], e[i] }
func (e SortableEntries) Less(i, j int) bool { return e[i].SortIndex > e[j].SortIndex }

type TweetResponse struct {
	GlobalObjects struct {
		Tweets        map[string]APITweet        `json:"tweets"`
		Users         map[string]APIUser         `json:"users"`
		Notifications map[string]APINotification `json:"notifications"`
	} `json:"globalObjects"`
	Timeline struct {
		Instructions []struct {
			AddEntries struct {
				Entries SortableEntries `json:"entries"`
			} `json:"addEntries"`
			ReplaceEntry struct {
				Entry Entry
			} `json:"replaceEntry"`
		} `json:"instructions"`
	} `json:"timeline"`
}

var tombstone_types = map[string]string{
	"This Tweet was deleted by the Tweet author. Learn more":                                                   "deleted",
	"This Tweet is from a suspended account. Learn more":                                                       "suspended",
	"You’re unable to view this Tweet because this account owner limits who can view their Tweets. Learn more": "hidden",
	"This Tweet is unavailable. Learn more":                                                                    "unavailable",
	"This Tweet violated the Twitter Rules. Learn more":                                                        "violated",
	"This Tweet is from an account that no longer exists. Learn more":                                          "no longer exists",
	"Age-restricted adult content. This content might not be appropriate for people under 18 years old. To view this media, " +
		"you’ll need to log in to Twitter. Learn more": "age-restricted",

	// New versions that use "Post" instead of "Tweet" and "X" instead of "Twitter"
	"This Post was deleted by the Post author. Learn more":                                                   "deleted",
	"This Post is from a suspended account. Learn more":                                                      "suspended",
	"You’re unable to view this Post because this account owner limits who can view their Posts. Learn more": "hidden",
	"This Post is unavailable. Learn more":                                                                   "unavailable",
	"This Post violated the X Rules. Learn more":                                                             "violated",
	"This Post is from an account that no longer exists. Learn more":                                         "no longer exists",
}

/**
 * Insert tweets into GlobalObjects for each tombstone.  Returns a list of users that need to
 * be fetched for tombstones.
 */
func (t *TweetResponse) HandleTombstones() []UserHandle {
	ret := []UserHandle{}

	// Handle tombstones in quote-tweets
	for _, api_tweet := range t.GlobalObjects.Tweets {
		// Ignore if tweet doesn't have a quoted tweet
		if api_tweet.QuotedStatusIDStr == "" {
			continue
		}
		// Ignore if quoted tweet is in the Global Objects (i.e., not a tombstone)
		if _, ok := t.GlobalObjects.Tweets[api_tweet.QuotedStatusIDStr]; ok {
			continue
		}

		user_handle, err := ParseHandleFromTweetUrl(api_tweet.QuotedStatusPermalink.ExpandedURL)
		if err != nil {
			panic(err)
		}

		var tombstoned_tweet APITweet
		tombstoned_tweet.ID = int64(int_or_panic(api_tweet.QuotedStatusIDStr))
		tombstoned_tweet.UserHandle = string(user_handle)
		tombstoned_tweet.TombstoneText = "unavailable"

		ret = append(ret, user_handle)
		fmt.Printf("Adding quoted tombstoned tweet: TweetID %d, handle %q\n", tombstoned_tweet.ID, tombstoned_tweet.UserHandle)

		t.GlobalObjects.Tweets[api_tweet.QuotedStatusIDStr] = tombstoned_tweet
	}

	// Handle tombstones in the conversation flow
	entries := t.Timeline.Instructions[0].AddEntries.Entries
	sort.Sort(entries)
	for i, entry := range entries {
		if entry.GetTombstoneText() != "" {
			// Try to reconstruct the tombstone tweet
			var tombstoned_tweet APITweet
			tombstoned_tweet.ID = int64(i) // Set a default to prevent clobbering other tombstones
			if i+1 < len(entries) && entries[i+1].Content.Item.Content.Tweet.ID != 0 {
				next_tweet_id := entries[i+1].Content.Item.Content.Tweet.ID
				api_tweet, ok := t.GlobalObjects.Tweets[fmt.Sprint(next_tweet_id)]
				if !ok {
					panic("Weird situation!")
				}
				tombstoned_tweet.ID = api_tweet.InReplyToStatusID
				tombstoned_tweet.UserID = api_tweet.InReplyToUserID
				ret = append(ret, UserHandle(api_tweet.InReplyToScreenName))
			}
			if i-1 >= 0 && entries[i-1].Content.Item.Content.Tweet.ID != 0 {
				prev_tweet_id := entries[i-1].Content.Item.Content.Tweet.ID
				_, ok := t.GlobalObjects.Tweets[fmt.Sprint(prev_tweet_id)]
				if !ok {
					panic("Weird situation 2!")
				}
				tombstoned_tweet.InReplyToStatusID = prev_tweet_id
			}

			short_text, ok := tombstone_types[entry.GetTombstoneText()]
			if !ok {
				panic(fmt.Errorf("Unknown tombstone text %q:\n  %w", entry.GetTombstoneText(), EXTERNAL_API_ERROR))
			}
			tombstoned_tweet.TombstoneText = short_text

			// Add the tombstoned tweet to GlobalObjects
			t.GlobalObjects.Tweets[fmt.Sprint(tombstoned_tweet.ID)] = tombstoned_tweet
		}
	}

	return ret
}

func (t *TweetResponse) GetCursor() string {
	// TODO: is this function used anywhere other than Notifications?
	for _, instr := range t.Timeline.Instructions {
		if len(instr.AddEntries.Entries) > 0 {
			last_entry := instr.AddEntries.Entries[len(instr.AddEntries.Entries)-1]
			if strings.Contains(last_entry.EntryID, "cursor") {
				return last_entry.Content.Operation.Cursor.Value
			}
		}
	}

	// Next, try the other format ("replaceEntry")
	instructions := t.Timeline.Instructions
	last_replace_entry := instructions[len(instructions)-1].ReplaceEntry.Entry
	if strings.Contains(last_replace_entry.EntryID, "cursor") {
		return last_replace_entry.Content.Operation.Cursor.Value
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

func (t *TweetResponse) ToTweetTrove() (TweetTrove, error) {
	ret := NewTweetTrove()

	for _, single_tweet := range t.GlobalObjects.Tweets {
		if single_tweet.RetweetedStatusIDStr == "" {
			new_tweet, err := ParseSingleTweet(single_tweet)
			if err != nil {
				return ret, err
			}
			ret.Tweets[new_tweet.ID] = new_tweet
			for _, space := range new_tweet.Spaces {
				ret.Spaces[space.ID] = space
			}
		} else {
			new_retweet, err := ParseSingleRetweet(single_tweet)
			if err != nil {
				return ret, err
			}
			ret.Retweets[new_retweet.RetweetID] = new_retweet
		}
	}

	for _, user := range t.GlobalObjects.Users {
		new_user, err := ParseSingleUser(user)
		if err != nil {
			return ret, err
		}
		ret.Users[new_user.ID] = new_user
	}
	for _, n := range t.GlobalObjects.Notifications {
		new_notification := ParseSingleNotification(n)
		ret.Notifications[new_notification.ID] = new_notification
	}
	return ret, nil
}

func idstr_to_int(idstr string) int64 {
	id, err := strconv.Atoi(idstr)
	if err != nil {
		panic(err)
	}
	return int64(id)
}
