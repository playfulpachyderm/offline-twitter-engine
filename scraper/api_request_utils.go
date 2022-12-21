package scraper

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

const API_CONVERSATION_BASE_PATH = "https://twitter.com/i/api/2/timeline/conversation/"
const API_USER_TIMELINE_BASE_PATH = "https://api.twitter.com/2/timeline/profile/"

type API struct {
	IsAuthenticated     bool
	GuestToken          string
	AuthenticationToken string
	Client              http.Client
	CSRFToken           string
}

func NewGuestSession() API {
	guestAPIString, err := GetGuestToken()
	if err != nil {
		panic(err)
	}

	return API{
		IsAuthenticated:     false,
		GuestToken:          guestAPIString,
		AuthenticationToken: "",
		Client:              http.Client{Timeout: 10 * time.Second},
		CSRFToken:           "",
	}
}

func (api *API) LogIn(username string, password string) {
	// TODO authentication: Log in and save the authentication token(s), set `IsAuthenticated = true`
	loginURL := "https://twitter.com/i/api/1.1/onboarding/task.json"

	result := make(map[string]interface{})
	err := api.do_http_POST(loginURL+"?flow_name=login", "", &result)
	if err != nil {
		panic(err)
	}

	flow_token, is_ok := result["flow_token"]
	if !is_ok {
		panic("No flow token.")
	}
	flow_token_string := flow_token.(string)

	fmt.Println(flow_token.(string))

	login_body := `{"flow_token":"` + flow_token_string + `", "subtask_inputs": [{"subtask_id": "LoginJsInstrumentationSubtask", "js_instrumentation": {"response": "{\"rf\":{\"a560cdc18ff70ce7662311eac0f2441dd3d3ed27c354f082f587e7a30d1a7d5f\":72,\"a8e890b5fec154e7af62d8f529fbec5942dfdd7ad41597245b71a3fbdc9a180d\":176,\"a3c24597ad4c862773c74b9194e675e96a7607708f6bbd0babcfdf8b109ed86d\":-161,\"af9847e2cd4e9a0ca23853da4b46bf00a2e801f98dc819ee0dd6ecc1032273fa\":-8},\"s\":\"hOai7h2KQi4RBGKSYLUhH0Y0fBm5KHIJgxD5AmNKtwP7N8gpVuAqP8o9n2FpCnNeR1d6XbB0QWkGAHiXkKao5PhaeXEZgPJU1neLcVgTnGuFzpjDnGutCUgYaxNiwUPfDX0eQkgr_q7GWmbB7yyYPt32dqSd5yt-KCpSt7MOG4aFmGf11xWE4MTpXfkefbnX4CwZeEFKQQYzJptOvmUWa7qI0A69BSOs7HZ_4Wry2TwB9k03Q_S-MDZAZ3yB_L7WoosVVb1e84YWgaLWWzqhz4C77jDy6isT8EKSWKWnVctsIcaqM_wMV8AiYa5lr0_WkN5TwK9h0vDOTS1obOZuhAAAAYTZan_3\"}", "link": "next_link"}}]}`

	err = api.do_http_POST(loginURL, login_body, &result)
	if err != nil {
		panic(err)
	}

	flow_token, is_ok = result["flow_token"]
	if !is_ok {
		panic("No flow token.")
	}
	flow_token_string = flow_token.(string)

	fmt.Println(flow_token_string)

	login_body = `{"flow_token":"` + flow_token_string + `","subtask_inputs":[{"subtask_id":"LoginEnterUserIdentifierSSO","settings_list":{"setting_responses":[{"key":"user_identifier","response_data":{"text_data":{"result":"` + username + `"}}}],"link":"next_link"}}]}`

	err = api.do_http_POST(loginURL, login_body, &result)
	if err != nil {
		panic(err)
	}

	flow_token, is_ok = result["flow_token"]
	if !is_ok {
		panic("No flow token.")
	}
	flow_token_string = flow_token.(string)

	fmt.Println(flow_token_string)

	login_body = `{"flow_token":"` + flow_token_string + `","subtask_inputs":[{"subtask_id":"LoginEnterPassword","enter_password":{"password":"` + password + `","link":"next_link"}}]}`

	err = api.do_http_POST(loginURL, login_body, &result)
	if err != nil {
		panic(err)
	}

	flow_token, is_ok = result["flow_token"]
	if !is_ok {
		panic("No flow token.")
	}
	flow_token_string = flow_token.(string)

	fmt.Println(flow_token_string)

	login_body = `{"flow_token":"` + flow_token_string + `","subtask_inputs":[{"subtask_id":"AccountDuplicationCheck","check_logged_in_account":{"link":"AccountDuplicationCheck_false"}}]}`

	err = api.do_http_POST(loginURL, login_body, &result)
	if err != nil {
		panic(err)
	}

	dummyURL, err := url.Parse(loginURL)
	if err != nil {
		panic(err)
	}

	for _, cookie := range api.Client.Jar.Cookies(dummyURL) {
		if cookie.Name == "ct0" {
			api.CSRFToken = cookie.Value
		}
	}

	if api.CSRFToken == "" {
		panic("No CSRF Token Found")
	}

	api.IsAuthenticated = true

}

func (api *API) do_http_POST(url string, body string, result interface{}) error {
	req, err := http.NewRequest("POST", url, strings.NewReader(body))
	if err != nil {
		return fmt.Errorf("Error initializing HTTP POST request:\n  %w", err)
	}

	// Params for every request
	req.Header.Set("Authorization", "Bearer "+BEARER_TOKEN)
	req.Header.Set("x-twitter-client-language", "en")
	req.Header.Set("content-type", "application/json")

	if api.IsAuthenticated {
		// TODO authentication: add authentication headers/params
	} else {
		// Not authenticated; use guest token
		if api.GuestToken == "" {
			panic("No guest token set!")
		}
		req.Header.Set("X-Guest-Token", api.GuestToken)
	}

	resp, err := api.Client.Do(req)
	if err != nil {
		return fmt.Errorf("Error executing HTTP POST request:\n  %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		content, err := io.ReadAll(resp.Body)
		if err != nil {
			panic(err)
		}

		responseHeaders := ""
		for header := range resp.Header {
			responseHeaders += fmt.Sprintf("    %s: %s\n", header, resp.Header.Get(header))
		}
		return fmt.Errorf("HTTP %s\n%s\n%s", resp.Status, responseHeaders, content)
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("Error reading response body:\n  %w", err)
	}
	log.Debug(string(respBody))

	err = json.Unmarshal(respBody, result)
	if err != nil {
		return fmt.Errorf("Error parsing API response:\n  %w", err)
	}
	return nil

}

func (api API) do_http(url string, cursor string, result interface{}) error {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return fmt.Errorf("Error initializing HTTP GET request:\n  %w", err)
	}

	if cursor != "" {
		query := req.URL.Query()
		query.Add("cursor", cursor)
		req.URL.RawQuery = query.Encode()
	}

	// Params for every request
	req.Header.Set("Authorization", "Bearer "+BEARER_TOKEN)
	req.Header.Set("x-twitter-client-language", "en")

	if api.IsAuthenticated {
		// TODO authentication: add authentication headers/params
	} else {
		// Not authenticated; use guest token
		if api.GuestToken == "" {
			panic("No guest token set!")
		}
		req.Header.Set("X-Guest-Token", api.GuestToken)
	}

	resp, err := api.Client.Do(req)
	if err != nil {
		return fmt.Errorf("Error executing HTTP request:\n  %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 && resp.StatusCode != 403 {
		content, err := io.ReadAll(resp.Body)
		if err != nil {
			panic(err)
		}

		responseHeaders := ""
		for header := range resp.Header {
			responseHeaders += fmt.Sprintf("    %s: %s\n", header, resp.Header.Get(header))
		}
		return fmt.Errorf("HTTP %s\n%s\n%s", resp.Status, responseHeaders, content)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("Error reading response body:\n  %w", err)
	}
	log.Debug(string(body))

	err = json.Unmarshal(body, result)
	if err != nil {
		return fmt.Errorf("Error parsing API response:\n  %w", err)
	}
	return nil
}

// Add the query params to get all data
func add_tweet_query_params(query *url.Values) {
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
}

func (api API) GetFeedFor(user_id UserID, cursor string) (TweetResponse, error) {
	// TODO: this function isn't actually used for anything (APIv2 is used instead)
	url, err := url.Parse(fmt.Sprintf("%s%d.json", API_USER_TIMELINE_BASE_PATH, user_id))
	if err != nil {
		panic(err)
	}
	queryParams := url.Query()
	add_tweet_query_params(&queryParams)
	url.RawQuery = queryParams.Encode()

	var result TweetResponse
	err = api.do_http(url.String(), cursor, &result)

	return result, err
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
	// TODO: break up this URL into params so it's readable
	url, err := url.Parse("https://twitter.com/i/api/graphql/Ha9BKBF0uAz9d4-lz0jnYA/AudioSpaceById?variables=%7B%22id%22%3A%22" + string(id) + "%22%2C%22isMetatagsQuery%22%3Afalse%2C%22withSuperFollowsUserFields%22%3Atrue%2C%22withDownvotePerspective%22%3Afalse%2C%22withReactionsMetadata%22%3Afalse%2C%22withReactionsPerspective%22%3Afalse%2C%22withSuperFollowsTweetFields%22%3Atrue%2C%22withReplays%22%3Atrue%7D&features=%7B%22spaces_2022_h2_clipping%22%3Atrue%2C%22spaces_2022_h2_spaces_communities%22%3Atrue%2C%22responsive_web_twitter_blue_verified_badge_is_enabled%22%3Atrue%2C%22verified_phone_label_enabled%22%3Afalse%2C%22tweetypie_unmention_optimization_enabled%22%3Atrue%2C%22responsive_web_uc_gql_enabled%22%3Atrue%2C%22vibe_api_enabled%22%3Atrue%2C%22responsive_web_edit_tweet_api_enabled%22%3Atrue%2C%22graphql_is_translatable_rweb_tweet_is_translatable_enabled%22%3Atrue%2C%22standardized_nudges_misinfo%22%3Atrue%2C%22tweet_with_visibility_results_prefer_gql_limited_actions_policy_enabled%22%3Afalse%2C%22responsive_web_graphql_timeline_navigation_enabled%22%3Atrue%2C%22interactive_text_enabled%22%3Atrue%2C%22responsive_web_text_conversations_enabled%22%3Afalse%2C%22responsive_web_enhance_cards_enabled%22%3Atrue%7D") //nolint:lll  // It's a URL, come on
	if err != nil {
		panic(err)
	}

	var result SpaceResponse
	err = api.do_http(url.String(), "", &result)
	return result, err
}

func (api API) GetTweet(id TweetID, cursor string) (TweetResponse, error) {
	url, err := url.Parse(fmt.Sprintf("%s%d.json", API_CONVERSATION_BASE_PATH, id))
	if err != nil {
		panic(err)
	}
	queryParams := url.Query()
	if cursor != "" {
		queryParams.Add("referrer", "tweet")
	}
	add_tweet_query_params(&queryParams)
	url.RawQuery = queryParams.Encode()

	var result TweetResponse
	err = api.do_http(url.String(), cursor, &result)
	return result, err
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

func (api API) GetUser(handle UserHandle) (APIUser, error) {
	// TODO: break up this URL into params so it's readable
	url, err := url.Parse("https://api.twitter.com/graphql/4S2ihIKfF3xhp-ENxvUAfQ/UserByScreenName?variables=%7B%22screen_name%22%3A%22" +
		string(handle) + "%22%2C%22withHighlightedLabel%22%3Atrue%7D")
	if err != nil {
		panic(err)
	}

	var result UserResponse
	for retries := 0; retries < 3; retries += 1 {
		result = UserResponse{} // Clear any previous result
		err = api.do_http(url.String(), "", &result)
		if err != nil {
			return APIUser{}, err
		}

		if len(result.Errors) == 0 {
			// Success; no retrying needed
			break
		}

		if result.Errors[0].Code != 50 || result.Errors[0].Name == "NotFoundError" {
			// Retry ONLY if the error is code 50 (random authentication failure)
			// Do NOT retry on real errors
			break
		}
	}

	return result.ConvertToAPIUser(), err
}

func (api API) Search(query string, cursor string) (TweetResponse, error) {
	url, err := url.Parse("https://twitter.com/i/api/2/search/adaptive.json")
	if err != nil {
		panic(err)
	}

	queryParams := url.Query()
	add_tweet_query_params(&queryParams)
	queryParams.Add("count", "50")
	queryParams.Add("spelling_corrections", "1")
	queryParams.Add("query_source", "typed_query")
	queryParams.Add("pc", "1")
	queryParams.Add("q", query)
	url.RawQuery = queryParams.Encode()
	fmt.Println(url.RawQuery)

	var result TweetResponse
	err = api.do_http(url.String(), cursor, &result)
	return result, err
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
