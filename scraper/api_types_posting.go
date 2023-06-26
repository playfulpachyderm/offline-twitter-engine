package scraper

import (
	"errors"
	"fmt"
	"strings"

	log "github.com/sirupsen/logrus"
)

var AlreadyLikedThisTweet error = errors.New("already liked this tweet")
var HaventLikedThisTweet error = errors.New("Haven't liked this tweet")

func (api API) LikeTweet(id TweetID) error {
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
		return fmt.Errorf("Error executing the HTTP POST request:\n  %w", err)
	}
	if len(result.Errors) > 0 {
		if strings.Contains(result.Errors[0].Message, "has already favorited tweet") {
			return AlreadyLikedThisTweet
		}
	}
	if result.Data.FavoriteTweet != "Done" {
		panic(fmt.Sprintf("Dunno why but it failed with value %q", result.Data.FavoriteTweet))
	}
	return nil
}

func (api API) UnlikeTweet(id TweetID) error {
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

func LikeTweet(id TweetID) error {
	if !the_api.IsAuthenticated {
		log.Fatalf("Must be authenticated!")
	}
	return the_api.LikeTweet(id)
}
func UnlikeTweet(id TweetID) error {
	if !the_api.IsAuthenticated {
		log.Fatalf("Must be authenticated!")
	}
	return the_api.UnlikeTweet(id)
}
