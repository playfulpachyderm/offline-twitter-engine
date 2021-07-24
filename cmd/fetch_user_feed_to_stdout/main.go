package main

import (
	"os"
	"fmt"
	"offline_twitter/scraper"
	"log"
	"sort"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Must provide a user handle!  Exiting...")
	}
	handle := scraper.UserHandle(os.Args[1])

	user, err := scraper.GetUser(handle)
	if err != nil {
		log.Fatal("Error getting user profile: " + err.Error())
	}

	tweets, retweets, users, err := scraper.GetFeedFull(user.ID, 1)
	if err != nil {
		log.Fatal("Error getting user feed: " + err.Error())
	}

	display_feed(user, tweets, retweets, users)

	fmt.Printf("Got a total of %d tweets, %d retweets, from %d users\n", len(tweets), len(retweets), len(users))
}

func display_feed(user scraper.User, tweets []scraper.Tweet, retweets []scraper.Retweet, users []scraper.User) {
	sort.Slice(tweets, func(i, j int) bool { return !tweets[i].PostedAt.Before(tweets[j].PostedAt) })
	tweet_map := make(map[scraper.TweetID]scraper.Tweet)
	for _, t := range tweets {
		tweet_map[t.ID] = t
	}

	sort.Slice(retweets, func(i, j int) bool { return !retweets[i].RetweetedAt.Before(retweets[j].RetweetedAt) })
	users_dict := make(map[scraper.UserID]scraper.User)
	for _, u := range users {
		users_dict[u.ID] = u
	}

	i := 0
	j := 0
	for i < len(tweets) && j < len(retweets) {
		if !tweets[i].PostedAt.Before(retweets[j].RetweetedAt) {
			tweet := tweets[i]
			if tweet.UserID != user.ID {
				i += 1
				continue
			}

			user, ok := users_dict[tweet.UserID]
			if !ok {
				log.Fatalf("User not found: %q", tweet.UserID)
			}

			print_tweet(tweets[i], user)
			i += 1
		} else {
			retweet := retweets[j]
			if retweet.RetweetedBy != user.ID {
				j += 1
				continue
			}
			tweet, ok := tweet_map[retweet.TweetID]
			if !ok {
				log.Fatalf("Tweet not found: %q", retweet.TweetID)
			}
			original_poster, ok := users_dict[tweet.UserID]
			if !ok {
				log.Fatalf("User not found: %q", tweet.UserID)
			}
			retweeter, ok := users_dict[retweet.RetweetedBy]
			if !ok {
				log.Fatalf("User not found: %q", retweet.RetweetedBy)
			}
			print_retweet(retweet, tweet, original_poster, retweeter)
			j += 1
		}
	}
	for i < len(tweets) {
		tweet := tweets[i]
		if tweet.UserID != user.ID {
			i += 1
			continue
		}

		user, ok := users_dict[tweet.UserID]
		if !ok {
			log.Fatalf("User not found: %q", tweet.UserID)
		}

		print_tweet(tweets[i], user)
		i += 1
	}
	for j < len(retweets) {
		retweet := retweets[j]
		if retweet.RetweetedBy != user.ID {
			j += 1
			continue
		}
		tweet, ok := tweet_map[retweet.TweetID]
		if !ok {
			log.Fatalf("Tweet not found: %q", retweet.TweetID)
		}
		original_poster, ok := users_dict[tweet.UserID]
		if !ok {
			log.Fatalf("User not found: %q", tweet.UserID)
		}
		retweeter, ok := users_dict[retweet.RetweetedBy]
		if !ok {
			log.Fatalf("User not found: %q", retweet.RetweetedBy)
		}
		print_retweet(retweet, tweet, original_poster, retweeter)
		j += 1
	}
}

func print_tweet(tweet scraper.Tweet, user scraper.User) {
	fmt.Printf("%s => %s\n    Replies: %d  Retweets: %d  Likes: %d\n", user.DisplayName, tweet.Text, tweet.NumReplies, tweet.NumRetweets, tweet.NumLikes)
}

func print_retweet(retweet scraper.Retweet, original_tweet scraper.Tweet, original_poster scraper.User, retweeter scraper.User) {
	fmt.Printf("%s [retweet] %s => %s\n   Replies: %d  Retweets: %d  Likes: %d\n", retweeter.DisplayName, original_poster.DisplayName, original_tweet.Text, original_tweet.NumReplies, original_tweet.NumRetweets, original_tweet.NumLikes)
}
