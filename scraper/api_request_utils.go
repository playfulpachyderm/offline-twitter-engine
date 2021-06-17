package scraper

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

const API_CONVERSATION_BASE_PATH = "https://twitter.com/i/api/2/timeline/conversation/"
const API_USER_TIMELINE_BASE_PATH = "https://api.twitter.com/2/timeline/profile/"

type API struct{}

func (api API) GetFeedFor(user_id UserID, cursor string) (TweetResponse, error) {
	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest("GET", API_USER_TIMELINE_BASE_PATH + string(user_id) + ".json", nil)
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

// Resend the request to get more tweets if necessary
func (api API) GetMoreTweets(user_id UserID, response *TweetResponse, max_tweets int) error {
	last_response := response
	for last_response.GetCursor() != "" && len(response.GlobalObjects.Tweets) < max_tweets {
		fresh_response, err := api.GetFeedFor(user_id, last_response.GetCursor())
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


func (api API) GetTweet(id string, cursor string) (TweetResponse, error) {
	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest("GET", API_CONVERSATION_BASE_PATH + id + ".json", nil)
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
		return TweetResponse{}, fmt.Errorf("HTTP %d %s: %s", resp.StatusCode, resp.Status, content)
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
func (api API) GetMoreReplies(tweet_id string, response *TweetResponse, max_replies int) error {
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

    resp, err := client.Do(req)
    if err != nil {
        return APIUser{}, err
    }
    defer resp.Body.Close()

    if !(resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusForbidden) {
        content, _ := ioutil.ReadAll(resp.Body)
        return APIUser{}, fmt.Errorf("response status %s: %s", resp.Status, content)
    }

    body, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        return APIUser{}, err
    }

    var response UserResponse
    err = json.Unmarshal(body, &response)
    return response.ConvertToAPIUser(), err
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
