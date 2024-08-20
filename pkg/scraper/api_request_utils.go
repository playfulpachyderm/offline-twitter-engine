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

type API struct {
	UserHandle      UserHandle
	UserID          UserID
	IsAuthenticated bool
	GuestToken      string
	Client          http.Client
	CSRFToken       string
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

func NewGuestSession() (API, error) {
	guestAPIString, err := GetGuestTokenWithRetries(3, 1*time.Second)
	if err != nil {
		return API{}, err
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
	}, nil
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

func is_session_invalidated(respBody []byte) bool {
	var result struct {
		Errors []struct {
			Message string
			Code    int
		} `json:"errors"`
	}
	err := json.Unmarshal(respBody, &result)
	if err != nil {
		panic(err)
	}
	return len(result.Errors) == 1 &&
		(result.Errors[0].Message == "Could not authenticate you" || result.Errors[0].Code == 32)
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
		if resp.StatusCode == 401 && is_session_invalidated(respBody) {
			return ErrSessionInvalidated
		}

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
		if resp.StatusCode == 401 && is_session_invalidated(body) {
			return ErrSessionInvalidated
		}

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

	if resp.StatusCode == 404 {
		log.Debugf("Media download 404 (%s)", remote_url)
		return body, ErrMediaDownload404
	}

	if resp.StatusCode != 200 {
		url, err := url.Parse(remote_url)
		if err != nil {
			panic(err)
		}
		print_curl_cmd(*req, api.Client.Jar.Cookies(url))

		if resp.StatusCode == 401 && is_session_invalidated(body) {
			return body, ErrSessionInvalidated
		}

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
