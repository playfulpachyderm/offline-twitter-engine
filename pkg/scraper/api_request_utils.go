package scraper

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

const API_CONVERSATION_BASE_PATH = "https://twitter.com/i/api/2/timeline/conversation/"
const API_USER_TIMELINE_BASE_PATH = "https://api.twitter.com/2/timeline/profile/"

type API struct {
	UserHandle      UserHandle
	UserID          UserID
	IsAuthenticated bool
	GuestToken      string
	Client          http.Client
	CSRFToken       string
}

// Use a global API variable since it is needed in so many utility functions (e.g.,
// tweet_trove.FillSpaceDetails, tweet_trove.FetchTombstoneUsers, etc.); this avoids having
// to inject it everywhere.
//
// Should be set by the caller (main program) depending on the session file used.
var the_api API

// Initializer for the global api variable
func InitApi(newApi API) {
	the_api = newApi
}

type api_outstruct struct {
	Cookies         []*http.Cookie
	UserID          UserID
	UserHandle      UserHandle
	IsAuthenticated bool
	GuestToken      string
	CSRFToken       string
}

var TWITTER_BASE_URL = url.URL{Scheme: "https", Host: "twitter.com"}

func (api API) MarshalJSON() ([]byte, error) {
	result, err := json.Marshal(api_outstruct{
		Cookies:         api.Client.Jar.Cookies(&TWITTER_BASE_URL),
		UserID:          api.UserID,
		UserHandle:      api.UserHandle,
		IsAuthenticated: api.IsAuthenticated,
		GuestToken:      api.GuestToken,
		CSRFToken:       api.CSRFToken,
	})
	if err != nil {
		return result, fmt.Errorf("Unable to JSONify the api:\n  %w", err)
	}
	return result, nil
}

func (api *API) UnmarshalJSON(data []byte) error {
	var in_struct api_outstruct
	err := json.Unmarshal(data, &in_struct)
	if err != nil {
		return fmt.Errorf("Unable to unmarshal:\n  %w", err)
	}
	cookie_jar, err := cookiejar.New(nil)
	if err != nil {
		panic(err)
	}
	for i := range in_struct.Cookies {
		in_struct.Cookies[i].Domain = ".twitter.com"
	}
	cookie_jar.SetCookies(&TWITTER_BASE_URL, in_struct.Cookies)
	api.IsAuthenticated = in_struct.IsAuthenticated
	api.GuestToken = in_struct.GuestToken
	api.UserID = in_struct.UserID
	api.UserHandle = in_struct.UserHandle

	api.Client = http.Client{
		Timeout: 10 * time.Second,
		Jar:     cookie_jar,
	}
	api.CSRFToken = in_struct.CSRFToken
	return nil
}

func (api API) add_authentication_headers(req *http.Request) {
	// Params for every request
	req.Header.Set("Authorization", "Bearer "+BEARER_TOKEN)
	req.Header.Set("x-twitter-client-language", "en")

	if api.IsAuthenticated {
		if api.CSRFToken == "" {
			panic("No CSRF token set!")
		}
		req.Header.Set("x-csrf-token", api.CSRFToken)
	} else {
		// Not authenticated; use guest token
		if api.GuestToken == "" {
			panic("No guest token set!")
		}
		req.Header.Set("X-Guest-Token", api.GuestToken)
	}
}

func NewGuestSession() API {
	guestAPIString, err := GetGuestToken()
	if err != nil {
		panic(err)
	}

	jar, err := cookiejar.New(nil)
	if err != nil {
		panic(err)
	}
	return API{
		IsAuthenticated: false,
		GuestToken:      guestAPIString,
		Client: http.Client{
			Timeout: 10 * time.Second,
			Jar:     jar,
		},
		CSRFToken: "",
	}
}

func (api *API) update_csrf_token() {
	dummyURL, err := url.Parse("https://twitter.com/i/api/1.1/onboarding/task.json")
	if err != nil {
		panic(err)
	}

	for _, cookie := range api.Client.Jar.Cookies(dummyURL) {
		if cookie.Name == "ct0" {
			api.CSRFToken = cookie.Value
			return
		}
	}
}

func is_timeout(err error) bool {
	var urlErr *url.Error
	if errors.As(err, &urlErr) {
		return urlErr.Timeout()
	}
	var netErr net.Error
	if errors.As(err, &netErr) {
		return netErr.Timeout()
	}
	return false
}

func (api *API) do_http_POST(remote_url string, body string, result interface{}) error {
	req, err := http.NewRequest("POST", remote_url, strings.NewReader(body))
	if err != nil {
		return fmt.Errorf("Error initializing HTTP POST request:\n  %w", err)
	}

	if len(body) == 0 || body[0] == '{' { // TODO: unclear what the content-type should be if body is empty; might not matter
		req.Header.Set("content-type", "application/json")
	} else {
		req.Header.Set("content-type", "application/x-www-form-urlencoded")
	}

	api.add_authentication_headers(req)

	log.Debugf("POST: %s\n", req.URL.String())
	for header := range req.Header {
		log.Debugf("    %s: %s\n", header, req.Header.Get(header))
	}
	log.Debug("    " + body)

	resp, err := api.Client.Do(req)
	if is_timeout(err) {
		return fmt.Errorf("POST %q:\n  %w", remote_url, ErrRequestTimeout)
	} else if err != nil {
		return fmt.Errorf("Error executing HTTP POST request:\n  %w", err)
	}
	api.update_csrf_token()

	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if is_timeout(err) {
		return fmt.Errorf("GET %q:\n  reading response body:\n  %w", remote_url, ErrRequestTimeout)
	} else if err != nil {
		panic(err)
	}

	if resp.StatusCode == 204 {
		// No Content
		return nil
	}

	if resp.StatusCode != 200 {
		responseHeaders := ""
		for header := range resp.Header {
			responseHeaders += fmt.Sprintf("    %s: %s\n", header, resp.Header.Get(header))
		}
		return fmt.Errorf("HTTP %s\n%s\n%s", resp.Status, responseHeaders, respBody)
	}

	log.Debug(string(respBody))

	err = json.Unmarshal(respBody, result)
	if err != nil {
		return fmt.Errorf("Error parsing API response:\n  %w", err)
	}
	return nil
}

func (api *API) do_http(remote_url string, cursor string, result interface{}) error {
	req, err := http.NewRequest("GET", remote_url, nil)
	if err != nil {
		return fmt.Errorf("Error initializing HTTP GET request:\n  %w", err)
	}

	if cursor != "" {
		query := req.URL.Query()
		query.Add("cursor", cursor)
		req.URL.RawQuery = query.Encode()
	}

	api.add_authentication_headers(req)

	log.Debugf("GET: %s\n", req.URL.String())
	for header := range req.Header {
		log.Debugf("    %s: %s\n", header, req.Header.Get(header))
	}

	resp, err := api.Client.Do(req)
	if is_timeout(err) {
		return fmt.Errorf("GET %q:\n  %w", remote_url, ErrRequestTimeout)
	} else if err != nil {
		return fmt.Errorf("Error executing HTTP request:\n  %w", err)
	}
	defer resp.Body.Close()

	if api.IsAuthenticated {
		// New request has been made, so the cookie will be changed; update the csrf to match
		api.update_csrf_token()
	}

	if resp.StatusCode == 429 {
		// "Too many requests" => rate limited
		reset_at := TimestampFromUnix(int64(int_or_panic(resp.Header.Get("X-Rate-Limit-Reset"))))
		return fmt.Errorf("%w (resets at %d, which is in %s)", ErrRateLimited, reset_at.Unix(), time.Until(reset_at.Time).String())
	}

	body, err := io.ReadAll(resp.Body)
	if is_timeout(err) {
		return fmt.Errorf("GET %q:\n  reading response body:\n  %w", remote_url, ErrRequestTimeout)
	} else if err != nil {
		panic(err)
	}

	if resp.StatusCode != 200 && resp.StatusCode != 403 {
		responseHeaders := ""
		for header := range resp.Header {
			responseHeaders += fmt.Sprintf("    %s: %s\n", header, resp.Header.Get(header))
		}
		return fmt.Errorf("HTTP Error.  HTTP %s\n%s\nbody: %s", resp.Status, responseHeaders, body)
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

func (api *API) GetTweet(id TweetID, cursor string) (TweetResponse, error) {
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
func (api *API) GetMoreReplies(tweet_id TweetID, response *TweetResponse, max_replies int) error {
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

func DownloadMedia(url string) ([]byte, error) {
	return the_api.DownloadMedia(url)
}

func (api *API) DownloadMedia(remote_url string) ([]byte, error) {
	fmt.Printf("Downloading: %s\n", remote_url)
	req, err := http.NewRequest("GET", remote_url, nil)
	if err != nil {
		panic(err)
	}
	// api.add_authentication_headers(req)
	// req.Header.Set("Referer", "https://twitter.com/") // DM embedded images require this header

	resp, err := api.Client.Do(req)
	if is_timeout(err) {
		return []byte{}, fmt.Errorf("GET %q:\n  waiting for headers:\n  %w", remote_url, ErrRequestTimeout)
	} else if err != nil {
		return []byte{}, fmt.Errorf("Error executing HTTP request:\n  %w", err)
	}
	defer resp.Body.Close()

	if api.IsAuthenticated {
		// New request has been made, so the cookie will be changed; update the csrf to match
		api.update_csrf_token()
	}

	body, err := io.ReadAll(resp.Body)
	if is_timeout(err) {
		return []byte{}, fmt.Errorf("GET %q:\n  reading response body:\n  %w", remote_url, ErrRequestTimeout)
	} else if err != nil {
		panic(err)
	}

	if resp.StatusCode == 403 {
		var response struct {
			Error_response string `json:"error_response"`
		}
		fmt.Println(string(body))

		err = json.Unmarshal(body, &response)
		if err != nil {
			panic(err)
		}
		if response.Error_response == "Dmcaed" {
			return body, ErrorDMCA
		}
		// Not a DCMA; fall through
	}

	if resp.StatusCode != 200 {
		url, err := url.Parse(remote_url)
		if err != nil {
			panic(err)
		}
		print_curl_cmd(*req, api.Client.Jar.Cookies(url))

		responseHeaders := ""
		for header := range resp.Header {
			responseHeaders += fmt.Sprintf("    %s: %s\n", header, resp.Header.Get(header))
		}
		log.Debug(responseHeaders)
		return body, fmt.Errorf("HTTP Error.  HTTP %s\n%s\nbody: %s", resp.Status, responseHeaders, body)
	}

	// Status code is HTTP 200
	return body, nil
}

func print_curl_cmd(r http.Request, cookies []*http.Cookie) {
	fmt.Printf("curl -X %s %q \\\n", r.Method, r.URL.String())
	for header := range r.Header {
		fmt.Printf("  -H '%s: %s' \\\n", header, r.Header.Get(header))
	}
	fmt.Printf("  -H 'Cookie: ")
	for _, c := range cookies {
		fmt.Printf("%s=%s;", c.Name, c.Value)
	}
	fmt.Printf("' \\\n")
	fmt.Printf("  --compressed\n")
}
