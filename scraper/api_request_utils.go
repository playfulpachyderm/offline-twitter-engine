package scraper

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	log "github.com/sirupsen/logrus"
)

const API_CONVERSATION_BASE_PATH = "https://twitter.com/i/api/2/timeline/conversation/"
const API_USER_TIMELINE_BASE_PATH = "https://api.twitter.com/2/timeline/profile/"

type API struct {
	IsAuthenticated     bool
	GuestToken          string
	AuthenticationToken string
}

func NewGuestSession() API {
	// TODO api-refactor: Return a new API with its GuestToken set using `GetGuestToken()` from "scraper/guest_token.go"
	panic("TODO")
}

func (api *API) LogIn(username string, password string) {
	// TODO authentication: Log in and save the authentication token(s), set `IsAuthenticated = true`
	panic("TODO")
}

func (api API) do_http(url string, cursor string, result *interface{}) {
	if api.IsAuthenticated {
		// TODO authentication: add authentication headers/params
	} else {
		// TODO api-refactor: add guest headers / params
	}

	// TODO api-refactor: do the HTTP request and unmarshal the result into the `result` struct
	// - if `cursor != ""`, then add the cursor to the request as in `UpdateQueryCursor` before
	// executing the request.
	// - ignore `referrer=tweet` (aka the boolean param) for now
	// - ignore retries for now
}

func (api API) GetFeedFor(user_id UserID, cursor string) (TweetResponse, error) {
	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest("GET", fmt.Sprintf("%s%d.json", API_USER_TIMELINE_BASE_PATH, user_id), nil)
	if err != nil {
		return TweetResponse{}, fmt.Errorf("Error initializing HTTP request for GetFeedFor(%d):\n  %w", user_id, err)
	}

	err = ApiRequestAddTokens(req)
	if err != nil {
		return TweetResponse{}, fmt.Errorf("Error adding tokens to HTTP request:\n  %w", err)
	}

	ApiRequestAddAllParams(req)

	if cursor != "" {
		UpdateQueryCursor(req, cursor, false)
	}

	resp, err := client.Do(req)
	if err != nil {
		return TweetResponse{}, fmt.Errorf("Error executing HTTP request for GetFeedFor(%d):\n  %w", user_id, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		content, err := io.ReadAll(resp.Body)
		if err != nil {
			panic(err)
		}
		s := ""
		for header := range resp.Header {
			s += fmt.Sprintf("    %s: %s\n", header, resp.Header.Get(header))
		}
		return TweetResponse{}, fmt.Errorf("HTTP %s\n%s\n%s", resp.Status, s, content)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return TweetResponse{}, fmt.Errorf("Error reading response body for GetUserFeedFor(%d):\n  %w", user_id, err)
	}
	log.Debug(string(body))

	var response TweetResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		return response, fmt.Errorf("Error parsing API response for GetUserFeedFor(%d):\n  %w", user_id, err)
	}
	return response, nil
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

func (api API) GetSpace(id SpaceID) (SpaceResponse, error) {
	client := &http.Client{Timeout: 10 * time.Second}
	log.Debug("asdfasd")
	req, err := http.NewRequest("GET", "https://twitter.com/i/api/graphql/Ha9BKBF0uAz9d4-lz0jnYA/AudioSpaceById?variables=%7B%22id%22%3A%22"+string(id)+"%22%2C%22isMetatagsQuery%22%3Afalse%2C%22withSuperFollowsUserFields%22%3Atrue%2C%22withDownvotePerspective%22%3Afalse%2C%22withReactionsMetadata%22%3Afalse%2C%22withReactionsPerspective%22%3Afalse%2C%22withSuperFollowsTweetFields%22%3Atrue%2C%22withReplays%22%3Atrue%7D&features=%7B%22spaces_2022_h2_clipping%22%3Atrue%2C%22spaces_2022_h2_spaces_communities%22%3Atrue%2C%22responsive_web_twitter_blue_verified_badge_is_enabled%22%3Atrue%2C%22verified_phone_label_enabled%22%3Afalse%2C%22tweetypie_unmention_optimization_enabled%22%3Atrue%2C%22responsive_web_uc_gql_enabled%22%3Atrue%2C%22vibe_api_enabled%22%3Atrue%2C%22responsive_web_edit_tweet_api_enabled%22%3Atrue%2C%22graphql_is_translatable_rweb_tweet_is_translatable_enabled%22%3Atrue%2C%22standardized_nudges_misinfo%22%3Atrue%2C%22tweet_with_visibility_results_prefer_gql_limited_actions_policy_enabled%22%3Afalse%2C%22responsive_web_graphql_timeline_navigation_enabled%22%3Atrue%2C%22interactive_text_enabled%22%3Atrue%2C%22responsive_web_text_conversations_enabled%22%3Afalse%2C%22responsive_web_enhance_cards_enabled%22%3Atrue%7D", //nolint:lll  // It's a URL, come on
		nil)
	if err != nil {
		return SpaceResponse{}, fmt.Errorf("Error initializing HTTP request:\n  %w", err)
	}
	err = ApiRequestAddTokens(req)
	if err != nil {
		return SpaceResponse{}, fmt.Errorf("Error adding tokens to HTTP request:\n  %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return SpaceResponse{}, fmt.Errorf("Error executing HTTP request for GetSpace(%s):\n  %w", id, err)
	}
	defer resp.Body.Close()
	if !(resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusForbidden) {
		content, err := io.ReadAll(resp.Body)
		if err != nil {
			panic(err)
		}
		return SpaceResponse{}, fmt.Errorf("Error getting %q.  HTTP %s: %s", req.URL, resp.Status, content)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return SpaceResponse{}, fmt.Errorf("Error reading HTTP request:\n  %w", err)
	}
	log.Debug(string(body))

	var response SpaceResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		return response, fmt.Errorf("Error parsing API response for GetSpace(%s):\n  %w", id, err)
	}
	return response, nil
}

func (api API) GetTweet(id TweetID, cursor string) (TweetResponse, error) {
	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest("GET", fmt.Sprintf("%s%d.json", API_CONVERSATION_BASE_PATH, id), nil)
	if err != nil {
		return TweetResponse{}, fmt.Errorf("Error initializing HTTP request:\n  %w", err)
	}

	err = ApiRequestAddTokens(req)
	if err != nil {
		return TweetResponse{}, fmt.Errorf("Error adding tokens to HTTP request:\n  %w", err)
	}

	ApiRequestAddAllParams(req)
	if cursor != "" {
		UpdateQueryCursor(req, cursor, true)
	}

	resp, err := client.Do(req)
	if err != nil {
		return TweetResponse{}, fmt.Errorf("Error executing HTTP request:\n  %w", err)
	}
	defer resp.Body.Close()

	if !(resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusForbidden) {
		content, err := io.ReadAll(resp.Body)
		if err != nil {
			panic(err)
		}
		return TweetResponse{}, fmt.Errorf("Error getting %q.  HTTP %s: %s", req.URL, resp.Status, content)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return TweetResponse{}, fmt.Errorf("Error reading HTTP request:\n  %w", err)
	}
	log.Debug(string(body))

	var response TweetResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		return response, fmt.Errorf("Error parsing API response for GetTweet(%d):\n  %w", id, err)
	}
	return response, nil
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
	req, err := http.NewRequest(
		"GET",
		"https://api.twitter.com/graphql/4S2ihIKfF3xhp-ENxvUAfQ/UserByScreenName?variables=%7B%22screen_name%22%3A%22"+string(handle)+
			"%22%2C%22withHighlightedLabel%22%3Atrue%7D",
		nil)
	if err != nil {
		return APIUser{}, fmt.Errorf("Error initializing HTTP request:\n  %w", err)
	}
	err = ApiRequestAddTokens(req)
	if err != nil {
		return APIUser{}, fmt.Errorf("Error adding tokens to HTTP request:\n  %w", err)
	}

	var response UserResponse
	for retries := 0; retries < 3; retries += 1 {
		resp, err := client.Do(req)
		if err != nil {
			return APIUser{}, fmt.Errorf("Error executing HTTP request for GetUser(%s):\n  %w", handle, err)
		}
		defer resp.Body.Close()

		// Sometimes it randomly gives 403 Forbidden.  API's fault, not ours
		// We check for this below
		if !(resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusForbidden) {
			content, err := io.ReadAll(resp.Body)
			if err != nil {
				panic(err)
			}
			return APIUser{}, fmt.Errorf("response status %s: %s", resp.Status, content)
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return APIUser{}, fmt.Errorf("Error retrieving API response to GetUser(%s):\n  %w", handle, err)
		}
		log.Debug("GetUser(" + string(handle) + "): " + string(body))

		err = json.Unmarshal(body, &response)
		if err != nil {
			return APIUser{}, fmt.Errorf("Error parsing API response to GetUser(%s):\n  %w", handle, err)
		}

		// Retry ONLY if the error is code 50 (random authentication failure), NOT on real errors
		if len(response.Errors) == 1 && response.Errors[0].Code == 50 && response.Errors[0].Name != "NotFoundError" {
			// Reset the response (remove the Errors)
			response = UserResponse{}
			continue
		} else {
			// Do not retry on real errors
			break
		}
	}
	return response.ConvertToAPIUser(), err
}

func (api API) Search(query string, cursor string) (TweetResponse, error) {
	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest(
		"GET",
		"https://twitter.com/i/api/2/search/adaptive.json?count=50&spelling_corrections=1&query_source=typed_query&pc=1&q="+
			url.QueryEscape(query),
		nil)
	if err != nil {
		return TweetResponse{}, fmt.Errorf("Error initializing HTTP request:\n  %w", err)
	}

	err = ApiRequestAddTokens(req)
	if err != nil {
		return TweetResponse{}, fmt.Errorf("Error adding tokens to HTTP request:\n  %w", err)
	}

	ApiRequestAddAllParams(req)
	if cursor != "" {
		UpdateQueryCursor(req, cursor, false)
	}

	fmt.Println(req.URL.String())

	resp, err := client.Do(req)
	if err != nil {
		return TweetResponse{}, fmt.Errorf("Error executing HTTP request for Search(%q):\n  %w", query, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		content, err := io.ReadAll(resp.Body)
		if err != nil {
			panic(err)
		}
		return TweetResponse{}, fmt.Errorf("Error while searching for %q.  HTTP %s: %s", req.URL, resp.Status, content)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return TweetResponse{}, fmt.Errorf("Error retrieving API response for Search(%q):\n  %w", query, err)
	}
	// fmt.Println(string(body))

	var response TweetResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		return response, fmt.Errorf("Error parsing API response to Search(%q):\n  %w", query, err)
	}
	return response, nil
}

func (api API) GetMoreTweetsFromSearch(query string, response *TweetResponse, max_results int) error {
	last_response := response
	for last_response.GetCursor() != "" && len(response.GlobalObjects.Tweets) < max_results {
		fresh_response, err := api.Search(query, last_response.GetCursor())
		if err != nil {
			return err
		}
		if fresh_response.GetCursor() == last_response.GetCursor() || len(fresh_response.GlobalObjects.Tweets) == 0 {
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
	req.Header.Set("Authorization", "Bearer "+BEARER_TOKEN)
	req.Header.Set("x-twitter-client-language", "en")

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
	query.Add("include_ext_media_availability", "true")
	query.Add("send_error_codes", "true")
	query.Add("simple_quoted_tweet", "true")
	query.Add("include_tweet_replies", "true")
	query.Add("ext", "mediaStats,highlightedLabel")
	query.Add("count", "20")
	req.URL.RawQuery = query.Encode()
}
