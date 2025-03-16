package persistence_test

import (
	"fmt"
	"math/rand"
	"time"

	. "gitlab.com/offline-twitter/twitter_offline_engine/pkg/persistence"
)

// Load a test profile, or create it if it doesn't exist.
func create_or_load_profile(profile_path string) Profile {
	var profile Profile
	var err error

	if !file_exists(profile_path) {
		profile, err = NewProfile(profile_path)
		if err != nil {
			panic(err)
		}
		u := create_stable_user()
		err = profile.SaveUser(&u)
		if err != nil {
			panic(err)
		}
		err = profile.SaveTweet(create_stable_tweet())
		if err != nil {
			panic(err)
		}
		err = profile.SaveRetweet(create_stable_retweet())
		if err != nil {
			panic(err)
		}
		err = profile.SaveChatRoom(create_stable_chat_room())
	} else {
		profile, err = LoadProfile(profile_path)
	}
	if err != nil {
		panic(err)
	}
	return profile
}

// Create a stable user with a fixed ID and handle
func create_stable_user() User {
	return User{
		ID:                    UserID(-1),
		DisplayName:           "stable display name",
		Handle:                UserHandle("handle stable"),
		Bio:                   "stable bio",
		FollowersCount:        10,
		FollowingCount:        2000,
		Location:              "stable location",
		Website:               "stable website",
		JoinDate:              TimestampFromUnix(10000000),
		IsVerified:            true,
		IsPrivate:             false,
		ProfileImageUrl:       "stable profile image url",
		ProfileImageLocalPath: "stable profile image local path",
		BannerImageUrl:        "stable banner image url",
		BannerImageLocalPath:  "stable image local path",
		PinnedTweetID:         TweetID(345),
	}
}

// Create a semi-stable Image based on the given ID
func create_image_from_id(id int) Image {
	filename := fmt.Sprintf("image%d.jpg", id)
	return Image{
		ID:            ImageID(id),
		TweetID:       -1,
		Width:         id * 10,
		Height:        id * 5,
		RemoteURL:     filename,
		LocalFilename: filename,
		IsDownloaded:  false,
	}
}

// Create a semi-stable Video based on the given ID
func create_video_from_id(id int) Video {
	filename := fmt.Sprintf("video%d.jpg", id)
	return Video{
		ID:                 VideoID(id),
		TweetID:            -1,
		Width:              id * 10,
		Height:             id * 5,
		RemoteURL:          filename,
		LocalFilename:      filename,
		ThumbnailRemoteUrl: filename,
		ThumbnailLocalPath: filename,
		Duration:           10000,
		ViewCount:          200,
		IsDownloaded:       false,
		IsGif:              false,
	}
}

// Create a semi-stable Url based on the given ID
func create_url_from_id(id int) Url {
	s := fmt.Sprint(id)
	return Url{
		TweetID:             -1,
		Domain:              s + "domain",
		Text:                s + "text",
		ShortText:           s + "shorttext",
		Title:               s + "title",
		Description:         s + "description",
		ThumbnailWidth:      id * 23,
		ThumbnailHeight:     id * 7,
		ThumbnailRemoteUrl:  s + "remote url",
		ThumbnailLocalPath:  s + "local path",
		CreatorID:           UserID(id),
		SiteID:              UserID(id),
		HasCard:             true,
		IsContentDownloaded: false,
	}
}

// Create a semi-stable Poll based on the given ID
func create_poll_from_id(id int) Poll {
	s := fmt.Sprint(id)
	return Poll{
		ID:             PollID(id),
		TweetID:        -1,
		NumChoices:     2,
		Choice1:        s,
		Choice1_Votes:  1000,
		Choice2:        "Not " + s,
		Choice2_Votes:  1500,
		VotingDuration: 10,
		VotingEndsAt:   TimestampFromUnix(10000000),
		LastUpdatedAt:  TimestampFromUnix(10000),
	}
}

// Create a stable tweet with a fixed ID and content
func create_stable_tweet() Tweet {
	tweet_id := TweetID(-1)
	return Tweet{
		ID:             tweet_id,
		UserID:         -1,
		Text:           "stable text",
		PostedAt:       TimestampFromUnix(10000000),
		NumLikes:       10,
		NumRetweets:    10,
		NumReplies:     10,
		NumQuoteTweets: 10,
		Videos: []Video{
			create_video_from_id(-1),
		},
		Urls: []Url{
			create_url_from_id(-1),
		},
		Images: []Image{
			create_image_from_id(-1),
		},
		Mentions: CommaSeparatedList{},
		Hashtags: CommaSeparatedList{},
		Polls: []Poll{
			create_poll_from_id(-1),
		},
		Spaces: []Space{
			create_space_from_id(-1),
		},
		SpaceID:               SpaceID("some_id_-1"),
		IsConversationScraped: true,
		LastScrapedAt:         TimestampFromUnix(100000000),
	}
}

// Create a stable retweet with a fixed ID and parameters
func create_stable_retweet() Retweet {
	retweet_id := TweetID(-1)
	return Retweet{
		RetweetID:     retweet_id,
		TweetID:       -1,
		RetweetedByID: -1,
		RetweetedAt:   TimestampFromUnix(20000000),
	}
}

// Create a new user with a random ID and handle
func create_dummy_user() User {
	userID := rand.Int()

	return User{
		ID:                    UserID(userID),
		DisplayName:           "display name",
		Handle:                UserHandle(fmt.Sprintf("handle%d", userID)),
		Bio:                   "bio",
		FollowersCount:        0,
		FollowingCount:        1000,
		Location:              "location",
		Website:               "website",
		JoinDate:              Timestamp{time.Now().Truncate(1e9)}, // Round to nearest second
		IsVerified:            false,
		IsPrivate:             true,
		ProfileImageUrl:       "profile image url",
		ProfileImageLocalPath: "profile image local path",
		BannerImageUrl:        "banner image url",
		BannerImageLocalPath:  "banner image local path",
		PinnedTweetID:         TweetID(234),
	}
}

// Create a new tweet from the stable User, with a random ID and content
func create_dummy_tweet() Tweet {
	tweet_id := TweetID(rand.Int())

	img1 := create_image_from_id(rand.Int())
	img1.TweetID = tweet_id
	img2 := create_image_from_id(rand.Int())
	img2.TweetID = tweet_id
	vid := create_video_from_id(rand.Int())
	vid.TweetID = tweet_id

	url1 := create_url_from_id(rand.Int())
	url1.TweetID = tweet_id
	url2 := create_url_from_id(rand.Int())
	url2.TweetID = tweet_id

	poll := create_poll_from_id(rand.Int())
	poll.TweetID = tweet_id

	space := create_space_from_id(rand.Int())
	space_id := space.ID

	return Tweet{
		ID:             tweet_id,
		UserID:         create_stable_user().ID,
		Text:           "text",
		PostedAt:       Timestamp{time.Now().Truncate(1e9)}, // Round to nearest second
		NumLikes:       1,
		NumRetweets:    2,
		NumReplies:     3,
		NumQuoteTweets: 4,
		Videos:         []Video{vid},
		Urls:           []Url{url1, url2},
		Images:         []Image{img1, img2},
		Mentions:       CommaSeparatedList{"mention1", "mention2"},
		ReplyMentions:  CommaSeparatedList{"replymention1", "replymention2"},
		Hashtags:       CommaSeparatedList{"hash1", "hash2"},
		Polls:          []Poll{poll},
		Spaces:         []Space{space},
		SpaceID:        space_id,
	}
}

// Create a random tombstone
func create_dummy_tombstone() Tweet {
	tweet_id := TweetID(rand.Int())

	return Tweet{
		ID:            tweet_id,
		UserID:        -1,
		TombstoneType: "deleted",
		TombstoneText: "This Tweet was deleted by the Tweet author",
		IsStub:        true,
		Mentions:      CommaSeparatedList{},
		ReplyMentions: CommaSeparatedList{},
		Hashtags:      CommaSeparatedList{},
		Spaces:        []Space{},
	}
}

// Create a new retweet with a random ID for a given TweetID
func create_dummy_retweet(tweet_id TweetID) Retweet {
	retweet_id := TweetID(rand.Int())

	return Retweet{
		RetweetID:     retweet_id,
		TweetID:       tweet_id,
		RetweetedByID: -1,
		RetweetedAt:   TimestampFromUnix(20000000),
	}
}

// Create a semi-stable Space given an ID
func create_space_from_id(id int) Space {
	return Space{
		ID:             SpaceID(fmt.Sprintf("some_id_%d", id)),
		ShortUrl:       fmt.Sprintf("short_url_%d", id),
		State:          "Running",
		Title:          "Some Title",
		CreatedAt:      TimestampFromUnix(1000),
		StartedAt:      TimestampFromUnix(2000),
		EndedAt:        TimestampFromUnix(3000),
		UpdatedAt:      TimestampFromUnix(4000),
		CreatedById:    -1,
		ParticipantIds: []UserID{-1},
	}
}

func create_dummy_like() Like {
	return Like{
		TweetID: create_stable_tweet().ID,
		UserID:  create_stable_user().ID,
		SortID:  LikeSortID(12345),
	}
}

func create_dummy_bookmark() Bookmark {
	return Bookmark{
		TweetID: create_stable_tweet().ID,
		UserID:  create_stable_user().ID,
		SortID:  BookmarkSortID(12345),
	}
}

func create_stable_chat_room() DMChatRoom {
	id := DMChatRoomID("some chat room ID")

	return DMChatRoom{
		ID:             id,
		Type:           "ONE_ON_ONE",
		LastMessagedAt: TimestampFromUnix(123),
		IsNSFW:         false,
		Participants: map[UserID]DMChatParticipant{
			UserID(-1): {
				DMChatRoomID:                   id,
				UserID:                         UserID(-1),
				LastReadEventID:                DMMessageID(0),
				IsChatSettingsValid:            true,
				IsNotificationsDisabled:        false,
				IsMentionNotificationsDisabled: false,
				IsReadOnly:                     false,
				IsTrusted:                      true,
				IsMuted:                        false,
				Status:                         "some status",
			},
		},
	}
}

func create_dummy_chat_room() DMChatRoom {
	id := DMChatRoomID(fmt.Sprintf("Chat Room #%d", rand.Int()))

	return DMChatRoom{
		ID:             id,
		Type:           "ONE_ON_ONE",
		LastMessagedAt: TimestampFromUnix(10000),
		IsNSFW:         false,
		Participants: map[UserID]DMChatParticipant{
			UserID(-1): {
				DMChatRoomID:                   id,
				UserID:                         UserID(-1),
				LastReadEventID:                DMMessageID(0),
				IsChatSettingsValid:            true,
				IsNotificationsDisabled:        false,
				IsMentionNotificationsDisabled: false,
				IsReadOnly:                     false,
				IsTrusted:                      true,
				IsMuted:                        false,
				Status:                         "some status",
			},
		},
	}
}

func create_dummy_chat_message() DMMessage {
	id := DMMessageID(rand.Int())
	vid := create_video_from_id(int(id))
	vid.TweetID = TweetID(0)
	vid.DMMessageID = id
	img := create_image_from_id(int(id))
	img.TweetID = TweetID(0)
	img.DMMessageID = id
	url := create_url_from_id(int(id))
	url.TweetID = TweetID(0)
	url.DMMessageID = id
	return DMMessage{
		ID:           id,
		DMChatRoomID: create_stable_chat_room().ID,
		SenderID:     create_stable_user().ID,
		SentAt:       TimestampFromUnix(50000),
		RequestID:    "fwjefkj",
		Text:         fmt.Sprintf("This is message #%d", id),
		Reactions: map[UserID]DMReaction{
			UserID(-1): {
				ID:          id + 1,
				DMMessageID: id,
				SenderID:    UserID(-1),
				SentAt:      TimestampFromUnix(51000),
				Emoji:       "🤔",
			},
		},
		Videos: []Video{vid},
		Images: []Image{img},
		Urls:   []Url{url},
	}
}

func create_dummy_notification() Notification {
	id := NotificationID(fmt.Sprintf("Notification #%d", rand.Int()))

	return Notification{
		ID:              id,
		Type:            NOTIFICATION_TYPE_REPLY,
		SentAt:          TimestampFromUnix(10000),
		SortIndex:       rand.Int63(),
		UserID:          create_stable_user().ID,
		ActionUserID:    create_stable_user().ID,
		ActionTweetID:   create_stable_tweet().ID,
		ActionRetweetID: create_stable_retweet().RetweetID,
		HasDetail:       true,
		LastScrapedAt:   TimestampFromUnix(57234728),
		TweetIDs:        []TweetID{create_stable_tweet().ID},
		UserIDs:         []UserID{create_stable_user().ID},
		RetweetIDs:      []TweetID{create_stable_retweet().RetweetID},
	}
}
