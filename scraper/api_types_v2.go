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

type APIV2Tweet struct {
	APITweet
	RetweetedStatusResult struct {
		Result struct {
			ID int `json:"rest_id,string"`
			Legacy APITweet `json:"legacy"`
			Core struct {
				UserResults struct {
					Result struct {
						ID     int64   `json:"rest_id,string"`
						Legacy APIUser `json:"legacy"`
					} `json:"result"`
				} `json:"user_results"`
			} `json:"core"`
			QuotedStatusResult struct {
				Result struct {
					ID int64 `json:"rest_id,string"`
					Legacy APITweet `json:"legacy"`
					Core struct {
						UserResults struct {
							Result struct {
								ID     int64   `json:"rest_id,string"`
								Legacy APIUser `json:"legacy"`
							} `json:"result"`
						} `json:"user_results"`
					} `json:"core"`
				} `json:"result"`
			} `json:"quoted_status_result"`
		} `json:"result"`
	} `json:"retweeted_status_result"`
}

type APIV2Response struct {
	Data struct {
		User struct {
			Result struct {
				Timeline struct {
					Timeline struct {
						Instructions []struct {
							Type string `json:"type"`
							Entries []struct {
								EntryID string `json:"entryId"`
								SortIndex int64 `json:"sortIndex,string"`
								Content struct {
									ItemContent struct {
										EntryType string `json:"entryType"`
										TweetResults struct {
											Result struct {
												Legacy APIV2Tweet `json:"legacy"`
												Core struct {
													UserResults struct {
														Result struct {
															ID     int64   `json:"rest_id,string"`
															Legacy APIUser `json:"legacy"`
														} `json:"result"`
													} `json:"user_results"`
												} `json:"core"`
												QuotedStatusResult struct {  // Same as "Result"
													Result struct {
														ID int64 `json:"rest_id,string"`
														Legacy APIV2Tweet `json:"legacy"`
														Core struct {
															UserResults struct {
																Result struct {
																	ID     int64   `json:"rest_id,string"`
																	Legacy APIUser `json:"legacy"`
																} `json:"result"`
															} `json:"user_results"`
														} `json:"core"`
													} `json:"result"`
												} `json:"quoted_status_result"`
											} `json:"result"`
										} `json:"tweet_results"`
									} `json:"itemContent"`

									// Cursors
									EntryType string `json:"entryType"`
									Value string `json:"value"`
									CursorType string `json:"cursorType"`

								} `json:"content"`
							} `json:"entries"`
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
			// println(entry.EntryID)
			continue
		}

		result := entry.Content.ItemContent.TweetResults.Result
		apiv2_tweet := result.Legacy
		apiv2_user_result := result.Core.UserResults.Result
		apiv2_retweeted_tweet_result := apiv2_tweet.RetweetedStatusResult.Result
		apiv2_retweeted_tweet_user := apiv2_retweeted_tweet_result.Core.UserResults.Result
		apiv2_retweeted_quoted_result := apiv2_retweeted_tweet_result.QuotedStatusResult.Result
		apiv2_retweeted_quoted_user := apiv2_retweeted_quoted_result.Core.UserResults.Result
		apiv2_quoted_tweet_result := result.QuotedStatusResult.Result
		apiv2_quoted_user_result := apiv2_quoted_tweet_result.Core.UserResults.Result

		// Handle case of retweet (main tweet doesn't get parsed other than retweeted_at)
		if apiv2_retweeted_tweet_result.ID != 0 {
			orig_tweet, err := ParseSingleTweet(apiv2_retweeted_tweet_result.Legacy)
			if err != nil {
				return TweetTrove{}, err
			}
			ret.Tweets[orig_tweet.ID] = orig_tweet

			orig_user, err := ParseSingleUser(apiv2_retweeted_tweet_user.Legacy)
			if err != nil {
				return TweetTrove{}, err
			}
			orig_user.ID = UserID(apiv2_retweeted_tweet_user.ID)
			ret.Users[orig_user.ID] = orig_user

			retweeting_user, err := ParseSingleUser(apiv2_user_result.Legacy)
			if err != nil {
				return TweetTrove{}, err
			}
			retweeting_user.ID = UserID(apiv2_user_result.ID)
			ret.Users[retweeting_user.ID] = retweeting_user

			retweet := Retweet{}
			retweet.RetweetID = TweetID(apiv2_tweet.ID)
			retweet.TweetID = TweetID(orig_tweet.ID)
			retweet.RetweetedByID = retweeting_user.ID
			retweet.RetweetedAt, err = time.Parse(time.RubyDate, apiv2_tweet.CreatedAt)
			if err != nil {
				fmt.Printf("%v\n", apiv2_tweet)
				panic(err)
			}
			ret.Retweets[retweet.RetweetID] = retweet

			// Handle quoted tweet
			if apiv2_retweeted_quoted_result.ID != 0 {
				quoted_tweet, err := ParseSingleTweet(apiv2_retweeted_quoted_result.Legacy)
				if err != nil {
					return TweetTrove{}, err
				}
				ret.Tweets[quoted_tweet.ID] = quoted_tweet

				quoted_user, err := ParseSingleUser(apiv2_retweeted_quoted_user.Legacy)
				if err != nil {
					return TweetTrove{}, err
				}
				quoted_user.ID = UserID(apiv2_retweeted_quoted_user.ID)
				ret.Users[quoted_user.ID] = quoted_user
			}

			continue
		}

		// The main tweet
		tweet, err := ParseSingleTweet(apiv2_tweet.APITweet)
		if err != nil {
			return TweetTrove{}, err
		}
		ret.Tweets[tweet.ID] = tweet

		user, err := ParseSingleUser(apiv2_user_result.Legacy)
		if err != nil {
			return TweetTrove{}, err
		}
		user.ID = UserID(apiv2_user_result.ID)
		ret.Users[user.ID] = user

		// Handle quoted tweet
		if apiv2_quoted_tweet_result.ID != 0 {
			quoted_tweet, err := ParseSingleTweet(apiv2_quoted_tweet_result.Legacy.APITweet)
			if err != nil {
				return TweetTrove{}, err
			}
			ret.Tweets[quoted_tweet.ID] = quoted_tweet

			quoted_user, err := ParseSingleUser(apiv2_quoted_user_result.Legacy)
			if err != nil {
				return TweetTrove{}, err
			}
			quoted_user.ID = UserID(apiv2_quoted_user_result.ID)
			ret.Users[quoted_user.ID] = quoted_user
		}
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
