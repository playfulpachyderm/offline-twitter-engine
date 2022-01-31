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
												Legacy struct {
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
												} `json:"legacy"`
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
														Legacy struct {
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
														} `json:"legacy"`
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
