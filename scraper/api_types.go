package scraper

import "time"

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
		}
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

type TweetResponse struct {
	GlobalObjects struct {
		Tweets map[string]APITweet `json:"tweets"`
		Users  map[string]struct {
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
		} `json:"users"`
	} `json:"globalObjects"`
}

type UserResponse struct {
	Data struct {
		User struct {
			ID     string `json:"rest_id"`
			Legacy struct {
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
			} `json:"legacy"`
		} `json:"user"`
	} `json:"data"`
}
