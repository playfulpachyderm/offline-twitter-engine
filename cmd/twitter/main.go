package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	log "github.com/sirupsen/logrus" // TODO: remove eventually
	"golang.org/x/term"
	"io/fs"
	"os"
	"strconv"
	"strings"
	"syscall"
	"time"

	. "gitlab.com/offline-twitter/twitter_offline_engine/pkg/persistence"
	"gitlab.com/offline-twitter/twitter_offline_engine/pkg/scraper"
	"gitlab.com/offline-twitter/twitter_offline_engine/pkg/webserver"
)

// Global variable referencing the open data profile
var profile Profile

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
		if len(args) == 1 && (args[0] == "webserver" || args[0] == "fetch_timeline" ||
			args[0] == "fetch_timeline_following_only" || args[0] == "fetch_inbox" || args[0] == "get_bookmarks" ||
			args[0] == "get_notifications" || args[0] == "mark_notifications_as_read") {
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
	profile, err = LoadProfile(*profile_dir)
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
		profile.LoadSession(UserHandle(*session_name), &api)
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
		fetch_user(UserHandle(target))
	case "fetch_user_by_id":
		id, err := strconv.Atoi(target)
		if err != nil {
			panic(err)
		}
		fetch_user_by_id(UserID(id))
	case "download_user_content":
		download_user_content(UserHandle(target))
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
	case "get_followers_you_know":
		get_followers_you_know(target, *how_many)
	case "get_bookmarks":
		get_bookmarks(*how_many)
	case "fetch_timeline":
		fetch_timeline(false) // TODO: *how_many
	case "fetch_timeline_following_only":
		fetch_timeline(true)
	case "get_notifications":
		get_notifications(*how_many)
	case "mark_notifications_as_read":
		mark_notification_as_read()
	case "download_tweet_content":
		download_tweet_content(target)
	case "search":
		search(target, *how_many)
	case "follow":
		follow_user(target, true)
	case "unfollow":
		follow_user(target, false)
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

	profile.SaveSession(api.UserHandle, api.MustMarshalJSON())
	happy_exit("Logged in as "+string(api.UserHandle), nil)
}

/**
 * Create a data directory.
 *
 * args:
 * - target_dir: the location of the new data dir.
 */
func create_profile(target_dir string) {
	_, err := NewProfile(target_dir)
	if err != nil {
		panic(err)
	}
}

func _fetch_user_by_id(id UserID) error {
	user, err := scraper.GetUserByID(id)
	if errors.Is(err, scraper.ErrDoesntExist) {
		// Mark them as deleted.
		// Handle and display name won't be updated if the user exists.
		user = User{ID: id, DisplayName: "<Unknown User>", Handle: "<UNKNOWN USER>", IsDeleted: true}
	} else if err != nil {
		return fmt.Errorf("scraping error on user ID %d: %w", id, err)
	}
	log.Debugf("%#v\n", user)

	err = profile.SaveUser(&user)
	var conflict_err ErrConflictingUserHandle
	if errors.As(err, &conflict_err) {
		log.Warnf(
			"Conflicting user handle found (ID %d); old user has been marked deleted.  Rescraping them",
			conflict_err.ConflictingUserID,
		)
		if err := _fetch_user_by_id(conflict_err.ConflictingUserID); err != nil {
			return fmt.Errorf("error scraping conflicting user (ID %d): %w", conflict_err.ConflictingUserID, err)
		}
	} else if err != nil {
		return fmt.Errorf("error saving user: %w", err)
	}

	user, err = profile.GetUserByID(user.ID)
	if err != nil {
		panic(fmt.Sprintf("User not found for some reason: %s", err.Error()))
	}
	download_user_content(user.Handle)
	return nil
}

func fetch_user(handle UserHandle) {
	user, err := api.GetUser(handle)
	if errors.Is(err, scraper.ErrDoesntExist) {
		// There's several reasons we could get a ErrDoesntExist:
		//   1. account never existed (user made a CLI typo)
		//   2. user changed their handle
		//   3. user deleted their account
		// In case (1), we should just report the error; in case (2) and (3), it would be nice to rescrape by ID,
		// but that feels kind of too complicated to do here.  So just report the error and let the user decide
		die(fmt.Sprintf("User with handle %q doesn't exist.  Check spelling, or try scraping with the ID instead", handle), false, -1)
	} else if is_scrape_failure(err) {
		die(err.Error(), false, -1)
	}
	log.Debugf("%#v\n", user)

	err = profile.SaveUser(&user)
	var conflict_err ErrConflictingUserHandle
	if errors.As(err, &conflict_err) {
		log.Warnf(
			"Conflicting user handle found (ID %d); old user has been marked deleted.  Rescraping them",
			conflict_err.ConflictingUserID,
		)
		if err := _fetch_user_by_id(conflict_err.ConflictingUserID); err != nil {
			die(fmt.Sprintf("error scraping conflicting user (ID %d): %s", conflict_err.ConflictingUserID, err.Error()), false, 4)
		}
	} else if err != nil {
		die(fmt.Sprintf("error saving user: %s", err.Error()), false, 4)
	}

	download_user_content(handle)
	happy_exit("Saved the user", nil)
}

func fetch_user_by_id(id UserID) {
	err := _fetch_user_by_id(id)
	if err != nil {
		die(err.Error(), false, -1)
	}
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

	trove, err := api.GetTweetFullAPIV2(tweet_id, 1)
	if err != nil {
		die(fmt.Sprintf("Error fetching tweet: %s", err.Error()), false, -1)
	}

	// Find the main tweet and update its "is_conversation_downloaded" and "last_scraped_at"
	tweet, ok := trove.Tweets[tweet_id]
	if !ok {
		panic("Trove didn't contain its own tweet!")
	}
	tweet.LastScrapedAt = Timestamp{time.Now()}
	tweet.IsConversationScraped = true

	log.Debug(tweet)

	err2 := profile.SaveTweet(tweet)
	if err2 != nil {
		die(fmt.Sprintf("Error saving tweet: %s", err2.Error()), false, 4)
	}
	happy_exit("Saved the tweet", nil)
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
	full_save_tweet_trove(trove)

	happy_exit(fmt.Sprintf("Saved %d tweets and %d users", len(trove.Tweets), len(trove.Users)), err)
}

/**
 * Scrape a user feed and get a big blob of tweets and retweets.  Get 50 tweets.
 *
 * args:
 * - handle: the user handle to get
 */
func fetch_user_feed(handle string, how_many int) {
	user, err := profile.GetUserByHandle(UserHandle(handle))
	if is_scrape_failure(err) {
		die(fmt.Sprintf("Error getting user: %s\n  %s", handle, err.Error()), false, -1)
	}

	trove, err := api.GetUserFeed(user.ID, how_many)
	if is_scrape_failure(err) {
		die(fmt.Sprintf("Error scraping feed: %s\n  %s", handle, err.Error()), false, -2)
	}
	full_save_tweet_trove(trove)

	happy_exit(
		fmt.Sprintf("Saved %d tweets, %d retweets and %d users", len(trove.Tweets), len(trove.Retweets), len(trove.Users)),
		err,
	)
}

func get_user_likes(handle string, how_many int) {
	user, err := profile.GetUserByHandle(UserHandle(handle))
	if err != nil {
		die(fmt.Sprintf("Error getting user: %s\n  %s", handle, err.Error()), false, -1)
	}

	trove, err := api.GetUserLikes(user.ID, how_many)
	if is_scrape_failure(err) {
		die(fmt.Sprintf("Error scraping feed: %s\n  %s", handle, err.Error()), false, -2)
	}
	full_save_tweet_trove(trove)

	happy_exit(
		fmt.Sprintf("Saved %d tweets, %d retweets and %d users", len(trove.Tweets), len(trove.Retweets), len(trove.Users)),
		err,
	)
}

func get_followees(handle string, how_many int) {
	user, err := profile.GetUserByHandle(UserHandle(handle))
	if err != nil {
		die(fmt.Sprintf("Error getting user: %s\n  %s", handle, err.Error()), false, -1)
	}

	trove, err := api.GetFollowees(user.ID, how_many)
	if is_scrape_failure(err) {
		die(fmt.Sprintf("Error getting followees: %s\n  %s", handle, err.Error()), false, -2)
	}
	full_save_tweet_trove(trove)
	profile.SaveAsFolloweesList(user.ID, trove)

	happy_exit(fmt.Sprintf("Saved %d followees", len(trove.Users)), err)
}
func get_followers(handle string, how_many int) {
	user, err := profile.GetUserByHandle(UserHandle(handle))
	if err != nil {
		die(fmt.Sprintf("Error getting user: %s\n  %s", handle, err.Error()), false, -1)
	}
	trove, err := api.GetFollowers(user.ID, how_many)
	if is_scrape_failure(err) {
		die(fmt.Sprintf("Error getting followees: %s\n  %s", handle, err.Error()), false, -2)
	}
	full_save_tweet_trove(trove)
	profile.SaveAsFollowersList(user.ID, trove)

	happy_exit(fmt.Sprintf("Saved %d followers", len(trove.Users)), err)
}
func get_followers_you_know(handle string, how_many int) {
	user, err := profile.GetUserByHandle(UserHandle(handle))
	if err != nil {
		die(fmt.Sprintf("Error getting user: %s\n  %s", handle, err.Error()), false, -1)
	}
	trove, err := api.GetFollowersYouKnow(user.ID, how_many)
	if is_scrape_failure(err) {
		die(fmt.Sprintf("Error getting followees: %s\n  %s", handle, err.Error()), false, -2)
	}
	full_save_tweet_trove(trove)

	// You follow everyone in this list
	profile.SaveAsFolloweesList(api.UserID, trove)

	// ...and they follow the specified user
	profile.SaveAsFollowersList(user.ID, trove)

	happy_exit(fmt.Sprintf("Saved %d followers-you-know", len(trove.Users)), err)
}
func get_bookmarks(how_many int) {
	trove, err := api.GetBookmarks(how_many)
	if is_scrape_failure(err) {
		die(fmt.Sprintf("Error scraping bookmarks:\n  %s", err.Error()), false, -2)
	}
	full_save_tweet_trove(trove)

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
	full_save_tweet_trove(trove)

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
	err = profile.DownloadTweetContentFor(&tweet, api.DownloadMedia)
	if err != nil {
		panic("Error getting content: " + err.Error())
	}
}

func download_user_content(handle UserHandle) {
	user, err := profile.GetUserByHandle(handle)
	if err != nil {
		panic("Couldn't get the user from database: " + err.Error())
	}
	err = profile.DownloadUserContentFor(&user, api.DownloadMedia)
	if err != nil {
		panic("Error getting content: " + err.Error())
	}
}

func search(query string, how_many int) {
	trove, err := api.Search(query, how_many)
	if is_scrape_failure(err) {
		die(fmt.Sprintf("Error scraping search results: %s", err.Error()), false, -100)
	}
	full_save_tweet_trove(trove)

	happy_exit(fmt.Sprintf("Saved %d tweets and %d users", len(trove.Tweets), len(trove.Users)), err)
}

func follow_user(handle string, is_followed bool) {
	user, err := profile.GetUserByHandle(UserHandle(handle))
	if err != nil {
		panic("Couldn't get the user from database: " + err.Error())
	}
	if is_followed {
		err := api.FollowUser(user.ID)
		if err != nil {
			die(fmt.Sprintf("Failed to follow user:\n  %s", err.Error()), false, 1)
		}
		profile.SaveFollow(api.UserID, user.ID)
		happy_exit("Followed user: "+handle, nil)
	} else {
		err := api.UnfollowUser(user.ID)
		if err != nil {
			die(fmt.Sprintf("Failed to unfollow user:\n  %s", err.Error()), false, 1)
		}
		profile.DeleteFollow(api.UserID, user.ID)
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
	full_save_tweet_trove(trove)
	happy_exit(fmt.Sprintf("Saved %d messages from %d chats", len(trove.Messages), len(trove.Rooms)), nil)
}

func fetch_dm(id string, how_many int) {
	room, err := profile.GetChatRoom(DMChatRoomID(id))
	if is_scrape_failure(err) {
		panic(err)
	}
	max_id := DMMessageID(^uint(0) >> 1)
	trove, err := api.GetConversation(room.ID, max_id, how_many)
	if err != nil {
		die(fmt.Sprintf("Failed to fetch dm:\n  %s", err.Error()), false, 1)
	}
	full_save_tweet_trove(trove)
	happy_exit(
		fmt.Sprintf("Saved %d messages from %d chats", len(trove.Messages), len(trove.Rooms)),
		err,
	)
}

func send_dm(room_id string, text string, in_reply_to_id int) {
	room, err := profile.GetChatRoom(DMChatRoomID(room_id))
	if err != nil {
		die(fmt.Sprintf("No such chat room: %d", in_reply_to_id), false, 1)
	}

	trove, err := api.SendDMMessage(room.ID, text, DMMessageID(in_reply_to_id))
	if err != nil {
		die(fmt.Sprintf("Failed to send dm:\n  %s", err.Error()), false, 1)
	}
	full_save_tweet_trove(trove)
	happy_exit(fmt.Sprintf("Saved %d messages from %d chats", len(trove.Messages), len(trove.Rooms)), nil)
}

func send_dm_reacc(room_id string, in_reply_to_id int, reacc string) {
	room, err := profile.GetChatRoom(DMChatRoomID(room_id))
	if err != nil {
		die(fmt.Sprintf("No such chat room: %d", in_reply_to_id), false, 1)
	}
	_, err = profile.GetChatMessage(DMMessageID(in_reply_to_id))
	if err != nil {
		die(fmt.Sprintf("No such message: %d", in_reply_to_id), false, 1)
	}
	err = api.SendDMReaction(room.ID, DMMessageID(in_reply_to_id), reacc)
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
	trove, err = api.GetNotificationDetailForAll(trove, profile.CheckNotificationScrapesNeeded(trove))
	if err != nil {
		panic(err)
	}

	full_save_tweet_trove(trove)
	happy_exit(fmt.Sprintf("Saved %d notifications, %d tweets and %d users",
		len(trove.Notifications), len(trove.Tweets), len(trove.Users),
	), nil)
}

func mark_notification_as_read() {
	if err := api.MarkNotificationsAsRead(); err != nil {
		panic(err)
	}
	happy_exit("Notifications marked as read", nil)
}
