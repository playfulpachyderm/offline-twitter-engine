package scraper

import (
	"errors"
	"fmt"
	"strings"
	"net/url"

	. "gitlab.com/offline-twitter/twitter_offline_engine/pkg/persistence"
)

var AlreadyLikedThisTweet error = errors.New("already liked this tweet")
var HaventLikedThisTweet error = errors.New("Haven't liked this tweet")

func (api API) LikeTweet(id TweetID) (Like, error) {
	if !api.IsAuthenticated {
		return Like{}, ErrLoginRequired
	}
	type LikeResponse struct {
		Data struct {
			FavoriteTweet string `json:"favorite_tweet"`
		} `json:"data"`
		Errors []struct {
			Message string `json:"message"`
			Code    int    `json:"code"`
			Kind    string `json:"kind"`
			Name    string `json:"name"`
		} `json:"errors"`
	}
	var result LikeResponse
	err := api.do_http_POST(
		"https://twitter.com/i/api/graphql/lI07N6Otwv1PhnEgXILM7A/FavoriteTweet",
		"{\"variables\":{\"tweet_id\":\""+fmt.Sprint(id)+"\"},\"queryId\":\"lI07N6Otwv1PhnEgXILM7A\"}",
		&result,
	)
	if err != nil {
		return Like{}, fmt.Errorf("Error executing the HTTP POST request:\n  %w", err)
	}
	if len(result.Errors) > 0 {
		if strings.Contains(result.Errors[0].Message, "has already favorited tweet") {
			return Like{
				UserID:  api.UserID,
				TweetID: id,
				SortID:  -1,
			}, AlreadyLikedThisTweet
		}
	}
	if result.Data.FavoriteTweet != "Done" {
		panic(fmt.Sprintf("Dunno why but it failed with value %q", result.Data.FavoriteTweet))
	}
	return Like{
		UserID:  api.UserID,
		TweetID: id,
		SortID:  -1,
	}, nil
}

func (api API) UnlikeTweet(id TweetID) error {
	if !api.IsAuthenticated {
		return ErrLoginRequired
	}
	type UnlikeResponse struct {
		Data struct {
			UnfavoriteTweet string `json:"unfavorite_tweet"`
		} `json:"data"`
		Errors []struct {
			Message string `json:"message"`
			Code    int    `json:"code"`
			Kind    string `json:"kind"`
			Name    string `json:"name"`
		} `json:"errors"`
	}
	var result UnlikeResponse
	err := api.do_http_POST(
		"https://twitter.com/i/api/graphql/ZYKSe-w7KEslx3JhSIk5LA/UnfavoriteTweet",
		"{\"variables\":{\"tweet_id\":\""+fmt.Sprint(id)+"\"},\"queryId\":\"ZYKSe-w7KEslx3JhSIk5LA\"}",
		&result,
	)
	if err != nil {
		return fmt.Errorf("Error executing the HTTP POST request:\n  %w", err)
	}
	if len(result.Errors) > 0 {
		if strings.Contains(result.Errors[0].Message, "not found in actor's") {
			return HaventLikedThisTweet
		}
	}
	if result.Data.UnfavoriteTweet != "Done" {
		panic(fmt.Sprintf("Dunno why but it failed with value %q", result.Data.UnfavoriteTweet))
	}
	return nil
}

// Follow the given user
// INFO: manual testing only
func (api API) FollowUser(u_id UserID) error {
	url, err := url.Parse("https://twitter.com/i/api/1.1/friendships/create.json")
	if err != nil {
		panic(err)
	}

	query := url.Query()
	add_tweet_query_params(&query)
	query.Add("user_id", fmt.Sprint(u_id))
	url.RawQuery = query.Encode()

	var result APIUser
	return api.do_http_POST(url.String(), "", &result)
}

// Unfollow the given user
// INFO: manual testing only
func (api API) UnfollowUser(u_id UserID) error {
	url, err := url.Parse("https://twitter.com/i/api/1.1/friendships/destroy.json")
	if err != nil {
		panic(err)
	}

	query := url.Query()
	add_tweet_query_params(&query)
	query.Add("user_id", fmt.Sprint(u_id))
	url.RawQuery = query.Encode()

	var result APIUser
	return api.do_http_POST(url.String(), "", &result)
}
