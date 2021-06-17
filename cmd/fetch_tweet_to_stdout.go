package main

import (
	"os"
	"fmt"
	"offline_twitter/scraper"
	// "time"
	"log"
	"strings"
)

const INCLUDE_REPLIES = true;

// input: e.g., "https://twitter.com/michaelmalice/status/1395882872729477131"
func parse_tweet(url string) (string, error) {
	parts := strings.Split(url, "/")
	if len(parts) != 6 {
		return "", fmt.Errorf("Tweet format isn't right (%d)", len(parts))
	}
	if parts[0] != "https:" || parts[1] != "" || parts[2] != "twitter.com" || parts[4] != "status" {
		return "", fmt.Errorf("Tweet format isn't right")
	}
	return parts[5], nil
}

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Must provide tweet!  Exiting...")
	}

	tweet_id, err := parse_tweet(os.Args[1])
	if err != nil {
		log.Fatal(err.Error())
	}

	if INCLUDE_REPLIES {
		tweets, retweets, users, err := scraper.GetTweetFull(tweet_id)
		if err != nil {
			log.Fatal(err.Error())
		}
		for _, t := range tweets {
			fmt.Printf("%v\n", t)
		}
		for _, t := range retweets {
			fmt.Printf("%v\n", t)
		}
		for _, u := range users {
			fmt.Printf("%v\n", u)
		}
	} else {
		tweet, err := scraper.GetTweet(tweet_id)
		if err != nil {
			log.Fatal(err.Error())
		}
		fmt.Printf("%v\n", tweet)
	}
}
