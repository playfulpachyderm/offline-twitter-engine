package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	log "github.com/sirupsen/logrus"
	"golang.org/x/term"
	"io/fs"
	"os"
	"strconv"
	"strings"
	"syscall"
	"time"

	"gitlab.com/offline-twitter/twitter_offline_engine/internal/webserver"
	"gitlab.com/offline-twitter/twitter_offline_engine/pkg/persistence"
	"gitlab.com/offline-twitter/twitter_offline_engine/pkg/scraper"
)

// Global variable referencing the open data profile
var profile persistence.Profile

var version_string string

var api scraper.API

func main() {
	profile_dir := flag.String("profile", ".", "")
	flag.StringVar(profile_dir, "p", ".", "")

	use_default_profile := flag.Bool("default-profile", false, "")

	show_version_flag := flag.Bool("version", false, "")
	flag.BoolVar(show_version_flag, "v", false, "")

	session_name := flag.String("session", "", "Name of session file to use")
	flag.StringVar(session_name, "s", "", "")

	how_many := flag.Int("n", 50, "")
	flag.IntVar(how_many, "number", 50, "")

	delay := flag.String("delay", "0ms", "")

	var default_log_level string
	if version_string == "" {
		default_log_level = "debug"
	} else {
		default_log_level = "info"
	}
	log_level := flag.String("log-level", default_log_level, "")

	help := flag.Bool("help", false, "")
	flag.BoolVar(help, "h", false, "")

	flag.Usage = func() {
		die("", true, 1)
	}

	flag.Parse()
	args := flag.Args()

	if *show_version_flag {
		if version_string == "" {
			fmt.Println("Development version")
		} else {
			fmt.Println("v" + version_string)
		}
		os.Exit(0)
	}

	if *help {
		die("", true, 0)
	}

	logging_level, err := log.ParseLevel(*log_level)
	if err != nil {
		die(err.Error(), false, 1)
	}
	log.SetLevel(logging_level)

	if len(args) < 2 {
		if len(args) == 1 && (args[0] == "list_followed" || args[0] == "webserver" || args[0] == "fetch_timeline" ||
			args[0] == "fetch_timeline_following_only" || args[0] == "fetch_inbox" || args[0] == "get_bookmarks" ||
			args[0] == "get_notifications") {
			// Doesn't need a target, so create a fake second arg
			args = append(args, "")
		} else {
			die("", true, 1)
		}
	}

	operation := args[0]
	target := args[1]

	if operation == "create_profile" {
		create_profile(target)
		return
	}

	if *use_default_profile {
		if *profile_dir != "." {
			die("Invalid flags: either `--profile [...]` or `--default-profile` can be used, but not both", true, 2)
		}
		*profile_dir = get_default_profile()

		// Create default profile if necessary
		fileinfo, err := os.Stat(*profile_dir)
		if errors.Is(err, fs.ErrNotExist) {
			// Doesn't exist; create it
			create_profile(*profile_dir)
		} else if err != nil {
			// Unexpected error
			die(fmt.Sprintf("Default profile path (%s) is weird: %s", *profile_dir, err.Error()), false, 2)
		} else if !fileinfo.IsDir() {
			// It exists but it's not a directory
			die(fmt.Sprintf("Default profile path (%s) already exists and is not a directory", *profile_dir), false, 2)
		}
		// Path exists and is a directory; safe to continue
	}
	profile, err = persistence.LoadProfile(*profile_dir)
	if err != nil {
		if *use_default_profile {
			create_profile(*profile_dir)
		} else {
			die(fmt.Sprintf("Could not load profile: %s", err.Error()), true, 2)
		}
	}

	if *session_name != "" {
		if strings.HasSuffix(*session_name, ".session") {
			// Lop off the ".session" suffix (allows using `--session asdf.session` which lets you tab-autocomplete at command line)
			*session_name = (*session_name)[:len(*session_name)-8]
		}
		api = profile.LoadSession(scraper.UserHandle(*session_name))
	} else {
		var err error
		api, err = scraper.NewGuestSession()
		if err != nil {
			log.Warnf("Unable to initialize guest session!  Might be a network issue")
		} // Don't exit here, some operations don't require a connection
	}
	api.Delay, err = time.ParseDuration(*delay)
	if err != nil {
		die(fmt.Sprintf("Invalid delay: %q", *delay), false, 1)
	}

	switch operation {
	case "login":
		var password string
		if len(args) == 2 {
			fmt.Printf("Password for @%s: ", target)
			bytes_password, err := term.ReadPassword(int(syscall.Stdin))
			if err != nil {
				panic(err)
			}
			fmt.Println()
			password = string(bytes_password)
		} else {
			password = args[2]
		}
		login(target, password)
	case "fetch_user":
		fetch_user(scraper.UserHandle(target))
	case "download_user_content":
		download_user_content(scraper.UserHandle(target))
	case "fetch_tweet_only":
		fetch_tweet_only(target)
	case "fetch_tweet":
		fetch_tweet_conversation(target, *how_many)
	case "get_user_tweets":
		fetch_user_feed(target, *how_many)
	case "get_user_tweets_all":
		fetch_user_feed(target, 999999999)
	case "get_user_likes":
		get_user_likes(target, *how_many)
	case "get_user_likes_all":
		get_user_likes(target, 999999999)
	case "get_followers":
		get_followers(target, *how_many)
	case "get_followees":
		get_followees(target, *how_many)
	case "get_bookmarks":
		get_bookmarks(*how_many)
	case "fetch_timeline":
		fetch_timeline(false) // TODO: *how_many
	case "fetch_timeline_following_only":
		fetch_timeline(true)
	case "get_notifications":
		get_notifications(*how_many)
	case "download_tweet_content":
		download_tweet_content(target)
	case "search":
		search(target, *how_many)
	case "follow": // TODO: update these to use Lists
		follow_user(target, true)
	case "unfollow":
		follow_user(target, false)
	case "list_followed":
		list_followed()
	case "like_tweet":
		like_tweet(target)
	case "unlike_tweet":
		unlike_tweet(target)
	case "webserver":
		fs := flag.NewFlagSet("", flag.ExitOnError)
		should_auto_open := fs.Bool("auto-open", false, "")
		addr := fs.String("addr", "localhost:1973", "port to listen on") // Random port that's probably not in use

		if err := fs.Parse(args[1:]); err != nil {
			panic(err)
		}
		start_webserver(*addr, *should_auto_open)
	case "fetch_inbox":
		fetch_inbox(*how_many)
	case "fetch_dm":
		fetch_dm(target, *how_many)
	case "send_dm":
		if len(args) == 3 {
			send_dm(target, args[2], 0)
		} else {
			val, err := strconv.Atoi(args[3])
			if err != nil {
				panic(err)
			}
			send_dm(target, args[2], val)
		}
	case "send_dm_reacc":
		if len(args) != 4 {
			die("", true, 1)
		}
		val, err := strconv.Atoi(args[2])
		if err != nil {
			panic(err)
		}
		send_dm_reacc(args[1], val, args[3]) // room, message, emoji
	default:
		die(fmt.Sprintf("Invalid operation: %s", operation), true, 3)
	}
}

// Log into twitter
//
// args:
// - username: twitter username or email address
// - password: twitter account password
func login(username string, password string) {
	api, err := scraper.NewGuestSession()
	if err != nil {
		die(fmt.Sprintf("Unable to create session: %s", err.Error()), false, 1)
	}
	challenge := api.LogIn(username, password)
	if challenge != nil {
		fmt.Printf("Secondary challenge issued:\n")
		fmt.Printf("    >>> %s\n", challenge.PrimaryText)
		fmt.Printf("    >>> %s\n", challenge.SecondaryText)
		fmt.Printf("Response: ")
		phone_number, err := bufio.NewReader(os.Stdin).ReadString('\n')
		if err != nil {
			panic(err)
		}
		api.LoginVerifyPhone(*challenge, phone_number)
	}

	profile.SaveSession(api)
	happy_exit("Logged in as "+string(api.UserHandle), nil)
}

/**
 * Create a data directory.
 *
 * args:
 * - target_dir: the location of the new data dir.
 */
func create_profile(target_dir string) {
	_, err := persistence.NewProfile(target_dir)
	if err != nil {
		panic(err)
	}
}

/**
 * Scrape a user and save it in the database.
 *
 * args:
 * - handle: e.g., "michaelmalice"
 */
func fetch_user(handle scraper.UserHandle) {
	user, err := scraper.GetUser(handle)
	if is_scrape_failure(err) {
		die(err.Error(), false, -1)
	}
	log.Debug(user)

	err = profile.SaveUser(&user)
	if err != nil {
		die(fmt.Sprintf("Error saving user: %s", err.Error()), false, 4)
	}

	download_user_content(handle)
	happy_exit("Saved the user", nil)
}

/**
 * Scrape a single tweet and save it in the database.
 *
 * args:
 * - tweet_url: e.g., "https://twitter.com/michaelmalice/status/1395882872729477131"
 */
func fetch_tweet_only(tweet_identifier string) {
	tweet_id, err := extract_id_from(tweet_identifier)
	if err != nil {
		die(err.Error(), false, -1)
	}

	tweet, err := api.GetTweet(tweet_id)
	if is_scrape_failure(err) || errors.Is(err, scraper.ErrRateLimited) {
		die(fmt.Sprintf("Error fetching tweet: %s", err.Error()), false, -1)
	}
	log.Debug(tweet)

	err2 := profile.SaveTweet(tweet)
	if err2 != nil {
		die(fmt.Sprintf("Error saving tweet: %s", err2.Error()), false, 4)
	}
	happy_exit("Saved the tweet", err)
}

/**
 * Scrape a tweet and all associated info, and save it in the database.
 *
 * args:
 * - tweet_url: e.g., "https://twitter.com/michaelmalice/status/1395882872729477131"
 */
func fetch_tweet_conversation(tweet_identifier string, how_many int) {
	tweet_id, err := extract_id_from(tweet_identifier)
	if err != nil {
		die(err.Error(), false, -1)
	}

	trove, err := api.GetTweetFullAPIV2(tweet_id, how_many)
	if is_scrape_failure(err) {
		die(err.Error(), false, -1)
	}
	profile.SaveTweetTrove(trove, true, &api)

	happy_exit(fmt.Sprintf("Saved %d tweets and %d users", len(trove.Tweets), len(trove.Users)), err)
}

/**
 * Scrape a user feed and get a big blob of tweets and retweets.  Get 50 tweets.
 *
 * args:
 * - handle: the user handle to get
 */
func fetch_user_feed(handle string, how_many int) {
	user, err := profile.GetUserByHandle(scraper.UserHandle(handle))
	if is_scrape_failure(err) {
		die(fmt.Sprintf("Error getting user: %s\n  %s", handle, err.Error()), false, -1)
	}

	trove, err := api.GetUserFeed(user.ID, how_many)
	if is_scrape_failure(err) {
		die(fmt.Sprintf("Error scraping feed: %s\n  %s", handle, err.Error()), false, -2)
	}
	profile.SaveTweetTrove(trove, true, &api)

	happy_exit(
		fmt.Sprintf("Saved %d tweets, %d retweets and %d users", len(trove.Tweets), len(trove.Retweets), len(trove.Users)),
		err,
	)
}

func get_user_likes(handle string, how_many int) {
	user, err := profile.GetUserByHandle(scraper.UserHandle(handle))
	if err != nil {
		die(fmt.Sprintf("Error getting user: %s\n  %s", handle, err.Error()), false, -1)
	}

	trove, err := api.GetUserLikes(user.ID, how_many)
	if is_scrape_failure(err) {
		die(fmt.Sprintf("Error scraping feed: %s\n  %s", handle, err.Error()), false, -2)
	}
	profile.SaveTweetTrove(trove, true, &api)

	happy_exit(
		fmt.Sprintf("Saved %d tweets, %d retweets and %d users", len(trove.Tweets), len(trove.Retweets), len(trove.Users)),
		err,
	)
}

func get_followees(handle string, how_many int) {
	user, err := profile.GetUserByHandle(scraper.UserHandle(handle))
	if err != nil {
		die(fmt.Sprintf("Error getting user: %s\n  %s", handle, err.Error()), false, -1)
	}

	trove, err := api.GetFollowees(user.ID, how_many)
	if is_scrape_failure(err) {
		die(fmt.Sprintf("Error getting followees: %s\n  %s", handle, err.Error()), false, -2)
	}
	profile.SaveTweetTrove(trove, true, &api)
	profile.SaveAsFolloweesList(user.ID, trove)

	happy_exit(fmt.Sprintf("Saved %d followees", len(trove.Users)), err)
}
func get_followers(handle string, how_many int) {
	user, err := profile.GetUserByHandle(scraper.UserHandle(handle))
	if err != nil {
		die(fmt.Sprintf("Error getting user: %s\n  %s", handle, err.Error()), false, -1)
	}
	trove, err := api.GetFollowers(user.ID, how_many)
	if is_scrape_failure(err) {
		die(fmt.Sprintf("Error getting followees: %s\n  %s", handle, err.Error()), false, -2)
	}
	profile.SaveTweetTrove(trove, true, &api)
	profile.SaveAsFollowersList(user.ID, trove)

	happy_exit(fmt.Sprintf("Saved %d followers", len(trove.Users)), err)
}
func get_bookmarks(how_many int) {
	trove, err := api.GetBookmarks(how_many)
	if is_scrape_failure(err) {
		die(fmt.Sprintf("Error scraping bookmarks:\n  %s", err.Error()), false, -2)
	}
	profile.SaveTweetTrove(trove, true, &api)

	happy_exit(fmt.Sprintf(
		"Saved %d tweets, %d retweets, %d users, and %d bookmarks",
		len(trove.Tweets), len(trove.Retweets), len(trove.Users), len(trove.Bookmarks)),
		err,
	)
}
func fetch_timeline(is_following_only bool) {
	trove, err := api.GetHomeTimeline("", is_following_only)
	if is_scrape_failure(err) {
		die(fmt.Sprintf("Error fetching timeline:\n  %s", err.Error()), false, -2)
	}
	profile.SaveTweetTrove(trove, true, &api)

	happy_exit(
		fmt.Sprintf("Saved %d tweets, %d retweets and %d users", len(trove.Tweets), len(trove.Retweets), len(trove.Users)),
		err,
	)
}

func download_tweet_content(tweet_identifier string) {
	tweet_id, err := extract_id_from(tweet_identifier)
	if err != nil {
		die(err.Error(), false, -1)
	}

	tweet, err := profile.GetTweetById(tweet_id)
	if err != nil {
		panic(fmt.Errorf("Couldn't get tweet (ID %d) from database:\n  %w", tweet_id, err))
	}
	err = profile.DownloadTweetContentFor(&tweet, &api)
	if err != nil {
		panic("Error getting content: " + err.Error())
	}
}

func download_user_content(handle scraper.UserHandle) {
	user, err := profile.GetUserByHandle(handle)
	if err != nil {
		panic("Couldn't get the user from database: " + err.Error())
	}
	err = profile.DownloadUserContentFor(&user, &api)
	if err != nil {
		panic("Error getting content: " + err.Error())
	}
}

func search(query string, how_many int) {
	trove, err := api.Search(query, how_many)
	if is_scrape_failure(err) {
		die(fmt.Sprintf("Error scraping search results: %s", err.Error()), false, -100)
	}
	profile.SaveTweetTrove(trove, true, &api)

	happy_exit(fmt.Sprintf("Saved %d tweets and %d users", len(trove.Tweets), len(trove.Users)), err)
}

func follow_user(handle string, is_followed bool) {
	user, err := profile.GetUserByHandle(scraper.UserHandle(handle))
	if err != nil {
		panic("Couldn't get the user from database: " + err.Error())
	}
	profile.SetUserFollowed(&user, is_followed)

	if is_followed {
		happy_exit("Followed user: "+handle, nil)
	} else {
		happy_exit("Unfollowed user: "+handle, nil)
	}
}

func unlike_tweet(tweet_identifier string) {
	tweet_id, err := extract_id_from(tweet_identifier)
	if err != nil {
		die(err.Error(), false, -1)
	}
	err = api.UnlikeTweet(tweet_id)
	if err != nil {
		die(err.Error(), false, -10)
	}
	happy_exit("Unliked the tweet.", nil)
}

func like_tweet(tweet_identifier string) {
	tweet_id, err := extract_id_from(tweet_identifier)
	if err != nil {
		die(err.Error(), false, -1)
	}
	like, err := api.LikeTweet(tweet_id)
	if err != nil {
		die(err.Error(), false, -10)
	}
	err = profile.SaveLike(like)
	if err != nil {
		die(err.Error(), false, -1)
	}
	happy_exit("Liked the tweet.", nil)
}

func list_followed() {
	for _, handle := range profile.GetAllFollowedUsers() {
		fmt.Println(handle)
	}
}

func start_webserver(addr string, should_auto_open bool) {
	app := webserver.NewApp(profile)
	if api.UserHandle != "" {
		err := app.SetActiveUser(api.UserHandle)
		if err != nil {
			die(err.Error(), false, -1)
		}
	}
	app.Run(addr, should_auto_open)
}

func fetch_inbox(how_many int) {
	trove, _, err := api.GetInbox(how_many)
	if err != nil {
		die(fmt.Sprintf("Failed to fetch inbox:\n  %s", err.Error()), false, 1)
	}
	profile.SaveTweetTrove(trove, true, &api)
	happy_exit(fmt.Sprintf("Saved %d messages from %d chats", len(trove.Messages), len(trove.Rooms)), nil)
}

func fetch_dm(id string, how_many int) {
	room, err := profile.GetChatRoom(scraper.DMChatRoomID(id))
	if is_scrape_failure(err) {
		panic(err)
	}
	max_id := scraper.DMMessageID(^uint(0) >> 1)
	trove, err := api.GetConversation(room.ID, max_id, how_many)
	if err != nil {
		die(fmt.Sprintf("Failed to fetch dm:\n  %s", err.Error()), false, 1)
	}
	profile.SaveTweetTrove(trove, true, &api)
	happy_exit(
		fmt.Sprintf("Saved %d messages from %d chats", len(trove.Messages), len(trove.Rooms)),
		err,
	)
}

func send_dm(room_id string, text string, in_reply_to_id int) {
	room, err := profile.GetChatRoom(scraper.DMChatRoomID(room_id))
	if err != nil {
		die(fmt.Sprintf("No such chat room: %d", in_reply_to_id), false, 1)
	}

	trove, err := api.SendDMMessage(room.ID, text, scraper.DMMessageID(in_reply_to_id))
	if err != nil {
		die(fmt.Sprintf("Failed to send dm:\n  %s", err.Error()), false, 1)
	}
	profile.SaveTweetTrove(trove, true, &api)
	happy_exit(fmt.Sprintf("Saved %d messages from %d chats", len(trove.Messages), len(trove.Rooms)), nil)
}

func send_dm_reacc(room_id string, in_reply_to_id int, reacc string) {
	room, err := profile.GetChatRoom(scraper.DMChatRoomID(room_id))
	if err != nil {
		die(fmt.Sprintf("No such chat room: %d", in_reply_to_id), false, 1)
	}
	_, err = profile.GetChatMessage(scraper.DMMessageID(in_reply_to_id))
	if err != nil {
		die(fmt.Sprintf("No such message: %d", in_reply_to_id), false, 1)
	}
	err = api.SendDMReaction(room.ID, scraper.DMMessageID(in_reply_to_id), reacc)
	if err != nil {
		die(fmt.Sprintf("Failed to react to message:\n  %s", err.Error()), false, 1)
	}

	happy_exit("Sent the reaction", nil)
}

func get_notifications(how_many int) {
	trove, _, err := api.GetNotifications(how_many)
	if err != nil && !errors.Is(err, scraper.END_OF_FEED) {
		panic(err)
	}
	to_scrape := profile.CheckNotificationScrapesNeeded(trove)
	trove, err = api.GetNotificationDetailForAll(trove, to_scrape)
	if err != nil {
		panic(err)
	}

	profile.SaveTweetTrove(trove, true, &api)
	happy_exit(fmt.Sprintf("Saved %d notifications, %d tweets and %d users",
		len(trove.Notifications), len(trove.Tweets), len(trove.Users),
	), nil)
}
