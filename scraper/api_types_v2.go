package scraper

import (
	"fmt"
	"net/http"
	"net/url"
	"io/ioutil"
	"time"
	"encoding/json"
	"strings"
)

type CardValue struct {
	Type string `json:"type"`
	StringValue string `json:"string_value"`
	ImageValue struct {
		AltText string `json:"alt"`
		Height int `json:"height"`
		Width int `json:"width"`
		Url string `json:"url"`
	} `json:"image_value"`
	UserValue struct {
		ID int64 `json:"id_str,string"`
	} `json:"user_value"`
	BooleanValue bool `json:"boolean_value"`
}

type APIV2Card struct {
	Legacy struct {
		BindingValues []struct {
			Key string `json:"key"`
			Value CardValue `json:"value"`
		} `json:"binding_values"`
		Name string `json:"name"`
		Url string `json:"url"`
	} `json:"legacy"`
}
func (card APIV2Card) ParseAsUrl() Url {
	values := make(map[string]CardValue)
	for _, obj := range card.Legacy.BindingValues {
		values[obj.Key] = obj.Value
	}

	ret := Url{}
	ret.HasCard = true

	ret.ShortText = card.Legacy.Url
	ret.Domain = values["domain"].StringValue
	ret.Title = values["title"].StringValue
	ret.Description = values["description"].StringValue
	ret.IsContentDownloaded = false
	ret.CreatorID = UserID(values["creator"].UserValue.ID)
	ret.SiteID = UserID(values["site"].UserValue.ID)

	var thumbnail_url string
	if card.Legacy.Name == "summary_large_image" || card.Legacy.Name == "summary" {
		thumbnail_url = values["thumbnail_image_large"].ImageValue.Url
	} else if card.Legacy.Name == "player" {
		thumbnail_url = values["player_image_large"].ImageValue.Url
	} else {
		panic("TODO unknown card type")
	}

	if thumbnail_url != "" {
		ret.HasThumbnail = true
		ret.ThumbnailRemoteUrl = thumbnail_url
		ret.ThumbnailLocalPath = get_thumbnail_local_path(thumbnail_url)
		ret.ThumbnailWidth = values["thumbnail_image_large"].ImageValue.Width
		ret.ThumbnailHeight = values["thumbnail_image_large"].ImageValue.Height
	}
	return ret
}

type APIV2UserResult struct {
	UserResults struct {
		Result struct {
			ID     int64   `json:"rest_id,string"`
			Legacy APIUser `json:"legacy"`
		} `json:"result"`
	} `json:"user_results"`
}
func (u APIV2UserResult) ToUser() User {
	user, err := ParseSingleUser(u.UserResults.Result.Legacy)
	if err != nil {
		panic(err)
	}
	user.ID = UserID(u.UserResults.Result.ID)
	return user
}

type APIV2Result struct {
	Result struct {
		ID int64 `json:"rest_id,string"`
		Legacy APIV2Tweet `json:"legacy"`
		Tombstone *struct {
			Text struct {
				Text string `json:"text"`
			} `json:"text"`
		} `json:"tombstone"`
		Core *APIV2UserResult `json:"core"`
		Card APIV2Card `json:"card"`
		QuotedStatusResult *APIV2Result `json:"quoted_status_result"`
	} `json:"result"`
}
func (api_result APIV2Result) ToTweetTrove() TweetTrove {
	ret := NewTweetTrove()

	if api_result.Result.Core != nil {
		main_user := api_result.Result.Core.ToUser()
		ret.Users[main_user.ID] = main_user
	} /*else {
		// TODO
	}*/

	main_tweet_trove := api_result.Result.Legacy.ToTweetTrove()
	ret.MergeWith(main_tweet_trove)

	// Handle quoted tweet
	if api_result.Result.QuotedStatusResult != nil {
		quoted_api_result := api_result.Result.QuotedStatusResult

		// Quoted tweets might be tombstones!
		if quoted_api_result.Result.Tombstone != nil {
			tombstoned_tweet := &quoted_api_result.Result.Legacy.APITweet
			tombstoned_tweet.TombstoneText = quoted_api_result.Result.Tombstone.Text.Text
			tombstoned_tweet.ID = int64(int_or_panic(api_result.Result.Legacy.APITweet.QuotedStatusIDStr))
			handle, err := ParseHandleFromTweetUrl(api_result.Result.Legacy.APITweet.QuotedStatusPermalink.ExpandedURL)
			if err != nil {
				panic(err)
			}
			tombstoned_tweet.UserHandle = string(handle)
			ret.TombstoneUsers = append(ret.TombstoneUsers, handle)
		}

		quoted_trove := api_result.Result.QuotedStatusResult.ToTweetTrove()
		ret.MergeWith(quoted_trove)
	}

	// Handle URL cards
	if api_result.Result.Card.Legacy.Name == "summary_large_image" || api_result.Result.Card.Legacy.Name == "player" {
		url := api_result.Result.Card.ParseAsUrl()

		main_tweet := ret.Tweets[TweetID(api_result.Result.Legacy.ID)]
		found := false
		for i := range main_tweet.Urls {
			if main_tweet.Urls[i].ShortText != url.ShortText {
				continue
			}
			found = true
			url.Text = main_tweet.Urls[i].Text  // Copy the expanded URL over, since the card doesn't have it in the new API
			main_tweet.Urls[i] = url
		}
		if !found {
			panic("Tweet trove doesn't contain its own main tweet")
		}
	}

	return ret
}

type APIV2Tweet struct {
	// For some reason, retweets are nested *inside* the Legacy tweet, whereas
	// quoted-tweets are next to it, as their own tweet
	RetweetedStatusResult *APIV2Result `json:"retweeted_status_result"`
	APITweet
}
func (api_v2_tweet APIV2Tweet) ToTweetTrove() TweetTrove {
	ret := NewTweetTrove()

	// If there's a retweet, we ignore the main tweet except for posted_at and retweeting UserID
	if api_v2_tweet.RetweetedStatusResult != nil {
		orig_tweet_trove := api_v2_tweet.RetweetedStatusResult.ToTweetTrove()
		ret.MergeWith(orig_tweet_trove)


		retweet := Retweet{}
		var err error
		retweet.RetweetID = TweetID(api_v2_tweet.ID)
		retweet.TweetID = TweetID(api_v2_tweet.RetweetedStatusResult.Result.ID)
		retweet.RetweetedByID = UserID(api_v2_tweet.APITweet.UserID)
		retweet.RetweetedAt, err = time.Parse(time.RubyDate, api_v2_tweet.APITweet.CreatedAt)
		if err != nil {
			fmt.Printf("%v\n", api_v2_tweet)
			panic(err)
		}
		ret.Retweets[retweet.RetweetID] = retweet
	} else {
		main_tweet, err := ParseSingleTweet(api_v2_tweet.APITweet)
		if err != nil {
			panic(err)
		}
		ret.Tweets[main_tweet.ID] = main_tweet
	}

	return ret
}

type APIV2Entry struct {
	EntryID string `json:"entryId"`
	SortIndex int64 `json:"sortIndex,string"`
	Content struct {
		ItemContent struct {
			EntryType string `json:"entryType"`
			TweetResults APIV2Result `json:"tweet_results"`
		} `json:"itemContent"`

		// Cursors
		EntryType string `json:"entryType"`
		Value string `json:"value"`
		CursorType string `json:"cursorType"`

	} `json:"content"`
}

type APIV2Response struct {
	Data struct {
		User struct {
			Result struct {
				Timeline struct {
					Timeline struct {
						Instructions []struct {
							Type string `json:"type"`
							Entries []APIV2Entry`json:"entries"`
						} `json:"instructions"`
					} `json:"timeline"`
				} `json:"timeline"`
			} `json:"result"`
		} `json:"user"`
	} `json:"data"`
}

func (api_response APIV2Response) GetCursorBottom() string {
	entries := api_response.Data.User.Result.Timeline.Timeline.Instructions[0].Entries
	last_entry := entries[len(entries) - 1]
	if last_entry.Content.CursorType != "Bottom" {
		panic("No bottom cursor found")
	}

	return last_entry.Content.Value
}

/**
 * Parse the collected API response and turn it into a TweetTrove
 */
func (api_response APIV2Response) ToTweetTrove() (TweetTrove, error) {
	ret := NewTweetTrove()
	for _, entry := range api_response.Data.User.Result.Timeline.Timeline.Instructions[0].Entries {  // TODO: the second Instruction is the pinned tweet
		if !strings.HasPrefix(entry.EntryID, "tweet-") {
			continue
		}

		result := entry.Content.ItemContent.TweetResults

		main_trove := result.ToTweetTrove()
		ret.MergeWith(main_trove)
	}

	return ret, nil
}


func get_graphql_user_timeline_url(user_id UserID, cursor string) string {
	if cursor != "" {
		return "https://twitter.com/i/api/graphql/CwLU7qTfeu0doqhSr6tW4A/UserTweetsAndReplies?variables=%7B%22userId%22%3A%22" + fmt.Sprint(user_id) + "%22%2C%22count%22%3A40%2C%22cursor%22%3A%22" + url.QueryEscape(cursor) + "%22%2C%22includePromotedContent%22%3Atrue%2C%22withCommunity%22%3Atrue%2C%22withSuperFollowsUserFields%22%3Atrue%2C%22withBirdwatchPivots%22%3Afalse%2C%22withDownvotePerspective%22%3Afalse%2C%22withReactionsMetadata%22%3Afalse%2C%22withReactionsPerspective%22%3Afalse%2C%22withSuperFollowsTweetFields%22%3Atrue%2C%22withVoice%22%3Atrue%2C%22withV2Timeline%22%3Afalse%2C%22__fs_interactive_text%22%3Afalse%2C%22__fs_responsive_web_uc_gql_enabled%22%3Afalse%2C%22__fs_dont_mention_me_view_api_enabled%22%3Afalse%7D"
	}
	return "https://twitter.com/i/api/graphql/CwLU7qTfeu0doqhSr6tW4A/UserTweetsAndReplies?variables=%7B%22userId%22%3A%22" + fmt.Sprint(user_id) + "%22%2C%22count%22%3A40%2C%22includePromotedContent%22%3Afalse%2C%22withCommunity%22%3Atrue%2C%22withSuperFollowsUserFields%22%3Atrue%2C%22withBirdwatchPivots%22%3Afalse%2C%22withDownvotePerspective%22%3Afalse%2C%22withReactionsMetadata%22%3Afalse%2C%22withReactionsPerspective%22%3Afalse%2C%22withSuperFollowsTweetFields%22%3Atrue%2C%22withVoice%22%3Atrue%2C%22withV2Timeline%22%3Afalse%2C%22__fs_interactive_text%22%3Afalse%2C%22__fs_dont_mention_me_view_api_enabled%22%3Afalse%7D"
}

/**
 * Get a User feed using the new GraphQL twitter api
 */
func (api API) GetGraphqlFeedFor(user_id UserID, cursor string) (APIV2Response, error) {
	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest("GET", get_graphql_user_timeline_url(user_id, cursor), nil)
	if err != nil {
		return APIV2Response{}, err
	}

	err = ApiRequestAddTokens(req)
	if err != nil {
		return APIV2Response{}, err
	}

	if cursor != "" {
		UpdateQueryCursor(req, cursor, false)
	}

	resp, err := client.Do(req)
	if err != nil {
		return APIV2Response{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		content, _ := ioutil.ReadAll(resp.Body)
		s := ""
		for header := range resp.Header {
			s += fmt.Sprintf("    %s: %s\n", header, resp.Header.Get(header))
		}
		return APIV2Response{}, fmt.Errorf("HTTP %s\n%s\n%s", resp.Status, s, content)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return APIV2Response{}, err
	}
	fmt.Println(string(body))

	var response APIV2Response
	err = json.Unmarshal(body, &response)
	return response, err
}

/**
 * Resend the request to get more tweets if necessary
 *
 * args:
 * - user_id: the user's UserID
 * - response: an "out" parameter; the APIV2Response that tweets, RTs and users will be appended to
 * - min_tweets: the desired minimum amount of tweets to get
 */
func (api API) GetMoreTweetsFromGraphqlFeed(user_id UserID, response *APIV2Response, min_tweets int) error {
	// TODO user-feed-infinite-fetch: what if you reach the end of the user's timeline?  Might loop
	// forever getting no new tweets
	last_response := response
	for last_response.GetCursorBottom() != "" && len(response.Data.User.Result.Timeline.Timeline.Instructions[0].Entries) < min_tweets {
		fresh_response, err := api.GetGraphqlFeedFor(user_id, last_response.GetCursorBottom())
		if err != nil {
			return err
		}

		if fresh_response.GetCursorBottom() == last_response.GetCursorBottom() && len(fresh_response.Data.User.Result.Timeline.Timeline.Instructions[0].Entries) == 0 {
			// Empty response, cursor same as previous: end of feed has been reached
			return END_OF_FEED
		}
		if len(fresh_response.Data.User.Result.Timeline.Timeline.Instructions[0].Entries) == 0 {
			// Response has a pinned tweet, but no other content: end of feed has been reached
			return END_OF_FEED  // TODO: check that there actually is a pinned tweet and the request didn't just fail lol
		}

		last_response = &fresh_response

		// Copy over the entries
		response.Data.User.Result.Timeline.Timeline.Instructions[0].Entries = append(
			response.Data.User.Result.Timeline.Timeline.Instructions[0].Entries,
			last_response.Data.User.Result.Timeline.Timeline.Instructions[0].Entries...)

		fmt.Printf("Have %d entries so far\n", len(response.Data.User.Result.Timeline.Timeline.Instructions[0].Entries))
	}
	return nil
}
