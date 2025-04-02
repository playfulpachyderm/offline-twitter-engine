package persistence

import (
	"fmt"
	"path"
	"regexp"
)

const DEFAULT_PROFILE_IMAGE_URL = "https://abs.twimg.com/sticky/default_profile_images/default_profile.png"
const DEFAULT_PROFILE_IMAGE = "default_profile.png"

type UserID int64
type UserHandle string

type User struct {
	ID                    UserID     `db:"id"`
	DisplayName           string     `db:"display_name"`
	Handle                UserHandle `db:"handle"`
	Bio                   string     `db:"bio"`
	FollowingCount        int        `db:"following_count"`
	FollowersCount        int        `db:"followers_count"`
	Location              string     `db:"location"`
	Website               string     `db:"website"`
	JoinDate              Timestamp  `db:"join_date"`
	IsPrivate             bool       `db:"is_private"`
	IsVerified            bool       `db:"is_verified"`
	IsBanned              bool       `db:"is_banned"`
	IsDeleted             bool       `db:"is_deleted"`
	ProfileImageUrl       string     `db:"profile_image_url"`
	ProfileImageLocalPath string     `db:"profile_image_local_path"`
	BannerImageUrl        string     `db:"banner_image_url"`
	BannerImageLocalPath  string     `db:"banner_image_local_path"`

	PinnedTweetID TweetID `db:"pinned_tweet_id"`
	PinnedTweet   *Tweet

	IsContentDownloaded bool `db:"is_content_downloaded"`
	IsNeedingFakeID     bool
	IsIdFake            bool `db:"is_id_fake"`

	IsFollowed       bool `db:"is_followed"`
	IsFollowingYou   bool
	Lists            []List
	FollowersYouKnow []User
}

func GetUnknownUser() User {
	return User{
		ID:              UserID(0x4000000000000000), // 2^62
		DisplayName:     "<Unknown User>",
		Handle:          UserHandle("<UNKNOWN USER>"),
		Bio:             "<blank>",
		FollowersCount:  0,
		FollowingCount:  0,
		Location:        "<blank>",
		Website:         "<blank>",
		JoinDate:        TimestampFromUnix(0),
		IsVerified:      false,
		IsPrivate:       false,
		IsNeedingFakeID: false,
		IsIdFake:        true,
	}
}

/**
 * Unknown Users with handles are only created by direct GetUser calls (either `twitter fetch_user`
 * subcommand or as part of tombstone user fetching.)
 */
func GetUnknownUserWithHandle(handle UserHandle) User {
	return User{
		ID:              UserID(0), // 2^62 + 1...
		DisplayName:     string(handle),
		Handle:          handle,
		Bio:             "<blank>",
		FollowersCount:  0,
		FollowingCount:  0,
		Location:        "<blank>",
		Website:         "<blank>",
		JoinDate:        TimestampFromUnix(0),
		IsVerified:      false,
		IsPrivate:       false,
		IsNeedingFakeID: true,
		IsIdFake:        true,
	}
}

/**
 * Get the URL where we would expect to find a User's tiny profile image
 */
func (u User) GetTinyProfileImageUrl() string {
	// If profile image is empty, then just use the default profile image
	if u.ProfileImageUrl == "" {
		return DEFAULT_PROFILE_IMAGE_URL
	}

	// Check that the format is as expected
	r := regexp.MustCompile(`(\.\w{2,4})$`)
	if !r.MatchString(u.ProfileImageUrl) {
		return u.ProfileImageUrl
	}

	return r.ReplaceAllString(u.ProfileImageUrl, "_normal$1")
}

/**
 * If user has a profile image, return the local path for its tiny version.
 * If user has a blank or default profile image, return a non-personalized default path.
 */
func (u User) GetTinyProfileImageLocalPath() string {
	if u.ProfileImageUrl == "" {
		return path.Base(u.GetTinyProfileImageUrl())
	}

	r := regexp.MustCompile(`(\.\w{2,4})$`)
	if !r.MatchString(u.GetTinyProfileImageUrl()) {
		return string(u.Handle) + "_profile_" + path.Base(u.GetTinyProfileImageUrl()+".jpg")
	}

	return string(u.Handle) + "_profile_" + path.Base(u.GetTinyProfileImageUrl())
}

// Compute a path that will actually contain an image on disk (relative to the Profile)
// TODO: why there are so many functions that appear to do roughly the same thing?
func (u User) GetProfileImageLocalPath() string {
	if u.IsContentDownloaded || u.ProfileImageLocalPath == DEFAULT_PROFILE_IMAGE {
		return fmt.Sprintf("/profile_images/%s", u.ProfileImageLocalPath)
	}

	r := regexp.MustCompile(`(\.\w{2,4})$`)
	return fmt.Sprintf("/profile_images/%s", r.ReplaceAllString(u.ProfileImageLocalPath, "_normal$1"))
}
