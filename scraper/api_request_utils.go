package scraper

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"
)

const API_CONVERSATION_BASE_PATH = "https://twitter.com/i/api/2/timeline/conversation/"
const API_USER_TIMELINE_BASE_PATH = "https://api.twitter.com/2/timeline/profile/"

type APIError string
func (e APIError) Error() string {
	return string(e)
}

const END_OF_FEED = APIError("End of feed")

type API struct{}

func (api API) GetFeedFor(user_id UserID, cursor string) (TweetResponse, error) {
	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest("GET", fmt.Sprintf("%s%d.json", API_USER_TIMELINE_BASE_PATH, user_id), nil)
	if err != nil {
		return TweetResponse{}, err
	}

	err = ApiRequestAddTokens(req)
	if err != nil {
		return TweetResponse{}, err
	}

	ApiRequestAddAllParams(req)

	if cursor != "" {
		UpdateQueryCursor(req, cursor, false)
	}

	resp, err := client.Do(req)
	if err != nil {
		return TweetResponse{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		content, _ := ioutil.ReadAll(resp.Body)
		return TweetResponse{}, fmt.Errorf("HTTP %s: %s", resp.Status, content)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return TweetResponse{}, err
	}

	var response TweetResponse
	err = json.Unmarshal(body, &response)
	return response, err
}

/**
 * Resend the request to get more tweets if necessary
 *
 * args:
 * - user_id: the user's UserID
 * - response: an "out" parameter; the TweetResponse that tweets, RTs and users will be appended to
 * - min_tweets: the desired minimum amount of tweets to get
 */
func (api API) GetMoreTweetsFromFeed(user_id UserID, response *TweetResponse, min_tweets int) error {
	// TODO user-feed-infinite-fetch: what if you reach the end of the user's timeline?  Might loop
	// forever getting no new tweets
	last_response := response
	for last_response.GetCursor() != "" && len(response.GlobalObjects.Tweets) < min_tweets {
		fresh_response, err := api.GetFeedFor(user_id, last_response.GetCursor())
		if err != nil {
			return err
		}

		if fresh_response.GetCursor() == last_response.GetCursor() && len(fresh_response.GlobalObjects.Tweets) == 0 {
			// Empty response, cursor same as previous: end of feed has been reached
			return END_OF_FEED
		}
		if fresh_response.IsEndOfFeed() {
			// Response has a pinned tweet, but no other content: end of feed has been reached
			return END_OF_FEED
		}

		last_response = &fresh_response

		// Copy over the tweets and the users
		for id, tweet := range last_response.GlobalObjects.Tweets {
			response.GlobalObjects.Tweets[id] = tweet
		}
		for id, user := range last_response.GlobalObjects.Users {
			response.GlobalObjects.Users[id] = user
		}
		fmt.Printf("Have %d tweets, and %d users so far\n", len(response.GlobalObjects.Tweets), len(response.GlobalObjects.Users))
	}
	return nil
}


func (api API) GetTweet(id TweetID, cursor string) (TweetResponse, error) {
	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest("GET", fmt.Sprintf("%s%d.json", API_CONVERSATION_BASE_PATH, id), nil)
	if err != nil {
		return TweetResponse{}, err
	}

	err = ApiRequestAddTokens(req)
	if err != nil {
		return TweetResponse{}, err
	}

	ApiRequestAddAllParams(req)
	if cursor != "" {
		UpdateQueryCursor(req, cursor, true)
	}

	resp, err := client.Do(req)
	if err != nil {
		return TweetResponse{}, err
	}
	defer resp.Body.Close()

	if !(resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusForbidden) {
		content, _ := ioutil.ReadAll(resp.Body)
		return TweetResponse{}, fmt.Errorf("Error getting %q.  HTTP %s: %s", req.URL, resp.Status, content)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return TweetResponse{}, err
	}

	var response TweetResponse
	err = json.Unmarshal(body, &response)
	return response, err
}

// Resend the request to get more replies if necessary
func (api API) GetMoreReplies(tweet_id TweetID, response *TweetResponse, max_replies int) error {
	last_response := response
	for last_response.GetCursor() != "" && len(response.GlobalObjects.Tweets) < max_replies {
		fresh_response, err := api.GetTweet(tweet_id, last_response.GetCursor())
		if err != nil {
			return err
		}

		last_response = &fresh_response

		// Copy over the tweets and the users
		for id, tweet := range last_response.GlobalObjects.Tweets {
			response.GlobalObjects.Tweets[id] = tweet
		}
		for id, user := range last_response.GlobalObjects.Users {
			response.GlobalObjects.Users[id] = user
		}
	}
	return nil
}

func UpdateQueryCursor(req *http.Request, new_cursor string, is_tweet bool) {
	query := req.URL.Query()
	query.Add("cursor", new_cursor)
	if is_tweet {
		query.Add("referrer", "tweet")
	}
	req.URL.RawQuery = query.Encode()
}


func (api API) GetUser(handle UserHandle) (APIUser, error) {
    client := &http.Client{Timeout: 10 * time.Second}
    req, err := http.NewRequest("GET", "https://api.twitter.com/graphql/4S2ihIKfF3xhp-ENxvUAfQ/UserByScreenName?variables=%7B%22screen_name%22%3A%22" + string(handle) + "%22%2C%22withHighlightedLabel%22%3Atrue%7D", nil)
    if err != nil {
        return APIUser{}, err
    }
	err = ApiRequestAddTokens(req)
	if err != nil {
		return APIUser{}, err
	}

    var response UserResponse
	for retries := 0; retries < 3; retries += 1 {
	    resp, err := client.Do(req)
	    if err != nil {
	        return APIUser{}, err
	    }
	    defer resp.Body.Close()

	    // Sometimes it randomly gives 403 Forbidden.  API's fault, not ours
	    // We check for this below
	    if !(resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusForbidden) {
	        content, _ := ioutil.ReadAll(resp.Body)
	        return APIUser{}, fmt.Errorf("response status %s: %s", resp.Status, content)
	    }

	    body, err := ioutil.ReadAll(resp.Body)
	    if err != nil {
	        return APIUser{}, err
	    }

	    err = json.Unmarshal(body, &response)
	    if err != nil {
	        return APIUser{}, err
	    }

	    if len(response.Errors) == 0 {
	        break
	    }

	    // Reset the response (remove the Errors)
	    response = UserResponse{}
    }
    return response.ConvertToAPIUser(), err
}

func (api API) Search(query string, cursor string) (TweetResponse, error) {
	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest("GET", "https://twitter.com/i/api/2/search/adaptive.json?count=50&spelling_corrections=1&query_source=typed_query&pc=1&q=" + url.QueryEscape(query), nil)
	if err != nil {
		return TweetResponse{}, err
	}

	err = ApiRequestAddTokens(req)
	if err != nil {
		return TweetResponse{}, err
	}

	ApiRequestAddAllParams(req)
	if cursor != "" {
		UpdateQueryCursor(req, cursor, false)
	}

	fmt.Println(req.URL.String())

	resp, err := client.Do(req)
	if err != nil {
		return TweetResponse{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		content, _ := ioutil.ReadAll(resp.Body)
		return TweetResponse{}, fmt.Errorf("Error while searching for %q.  HTTP %s: %s", req.URL, resp.Status, content)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return TweetResponse{}, err
	}
	// fmt.Println(string(body))

	var response TweetResponse
	err = json.Unmarshal(body, &response)
	return response, err
}

func (api API) GetMoreTweetsFromSearch(query string, response *TweetResponse, max_results int) error {
	last_response := response
	for last_response.GetCursor() != "" && len(response.GlobalObjects.Tweets) < max_results {
		fresh_response, err := api.Search(query, last_response.GetCursor())
		if err != nil {
			return err
		}
		if fresh_response.GetCursor() == last_response.GetCursor() && len(fresh_response.GlobalObjects.Tweets) == 0 {
			// Empty response, cursor same as previous: end of feed has been reached
			return END_OF_FEED
		}

		last_response = &fresh_response

		// Copy the results over
		for id, tweet := range last_response.GlobalObjects.Tweets {
			response.GlobalObjects.Tweets[id] = tweet
		}
		for id, user := range last_response.GlobalObjects.Users {
			response.GlobalObjects.Users[id] = user
		}
		fmt.Printf("Have %d tweets\n", len(response.GlobalObjects.Tweets))
		// fmt.Printf("Cursor: %s\n", last_response.GetCursor())
	}
	return nil
}


// Add Bearer token and guest token
func ApiRequestAddTokens(req *http.Request) error {
	req.Header.Set("Authorization", "Bearer " + BEARER_TOKEN)

	guestToken, err := GetGuestToken()
	if err != nil {
		return err
	}
	req.Header.Set("X-Guest-Token", guestToken)
	return nil
}

// Add the query params to get all data
func ApiRequestAddAllParams(req *http.Request) {
	query := req.URL.Query()
	query.Add("include_profile_interstitial_type", "1")
	query.Add("include_blocking", "1")
	query.Add("include_blocked_by", "1")
	query.Add("include_followed_by", "1")
	query.Add("include_want_retweets", "1")
	query.Add("include_mute_edge", "1")
	query.Add("include_can_dm", "1")
	query.Add("include_can_media_tag", "1")
	query.Add("skip_status", "1")
	query.Add("cards_platform", "Web-12")
	query.Add("include_cards", "1")
	query.Add("include_ext_alt_text", "true")
	query.Add("include_quote_count", "true")
	query.Add("include_reply_count", "1")
	query.Add("tweet_mode", "extended")
	query.Add("include_entities", "true")
	query.Add("include_user_entities", "true")
	query.Add("include_ext_media_color", "true")
	query.Add("include_ext_media_availability", "true")
	query.Add("send_error_codes", "true")
	query.Add("simple_quoted_tweet", "true")
	query.Add("include_tweet_replies", "true")
	query.Add("ext", "mediaStats,highlightedLabel")
	query.Add("count", "20")
	req.URL.RawQuery = query.Encode()
}
