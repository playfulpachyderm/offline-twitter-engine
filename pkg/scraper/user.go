package scraper

import (
	"fmt"
	"path"
	"regexp"
	"strings"

	"gitlab.com/offline-twitter/twitter_offline_engine/pkg/terminal_utils"
)

const DEFAULT_PROFILE_IMAGE_URL = "https://abs.twimg.com/sticky/default_profile_images/default_profile.png"
const DEFAULT_PROFILE_IMAGE = "default_profile.png"

type UserID int64
type UserHandle string

func JoinArrayOfHandles(handles []UserHandle) string {
	ret := []string{}
	for _, h := range handles {
		ret = append(ret, string(h))
	}
	return strings.Join(ret, ",")
}

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

	IsFollowed          bool `db:"is_followed"`
	IsContentDownloaded bool `db:"is_content_downloaded"`
	IsNeedingFakeID     bool
	IsIdFake            bool `db:"is_id_fake"`
}

func (u User) String() string {
	var verified string
	if u.IsVerified {
		verified = "[\u2713]"
	}
	ret := fmt.Sprintf(
		`%s%s
@%s
    %s

Following: %d      Followers: %d

Joined %s
%s
%s
`,
		u.DisplayName,
		verified,
		u.Handle,
		terminal_utils.WrapText(u.Bio, 60),
		u.FollowingCount,
		u.FollowersCount,
		terminal_utils.FormatDate(u.JoinDate.Time),
		u.Location,
		u.Website,
	)
	if u.PinnedTweet != nil {
		ret += "\n" + terminal_utils.WrapText(u.PinnedTweet.Text, 60)
	} else {
		println("Pinned tweet id:", u.PinnedTweetID)
	}
	return ret
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

// Turn an APIUser, as returned from the scraper, into a properly structured User object
func ParseSingleUser(apiUser APIUser) (ret User, err error) {
	if apiUser.DoesntExist {
		// User may have been deleted, or there was a typo.  There's no data to parse
		if apiUser.ScreenName == "" {
			panic("ScreenName is empty!")
		}
		ret = GetUnknownUserWithHandle(UserHandle(apiUser.ScreenName))
		return
	}
	ret.ID = UserID(apiUser.ID)
	ret.Handle = UserHandle(apiUser.ScreenName)
	if apiUser.IsBanned {
		// Banned users won't have any further info, so just return here
		ret.IsBanned = true
		return
	}
	ret.DisplayName = apiUser.Name
	ret.Bio = apiUser.Description
	ret.FollowingCount = apiUser.FriendsCount
	ret.FollowersCount = apiUser.FollowersCount
	ret.Location = apiUser.Location
	if len(apiUser.Entities.URL.Urls) > 0 {
		ret.Website = apiUser.Entities.URL.Urls[0].ExpandedURL
	}
	ret.JoinDate, err = TimestampFromString(apiUser.CreatedAt)
	if err != nil {
		err = fmt.Errorf("Error parsing time on user ID %d: %w", ret.ID, err)
		return
	}
	ret.IsPrivate = apiUser.Protected
	ret.IsVerified = apiUser.Verified
	ret.ProfileImageUrl = apiUser.ProfileImageURLHTTPS

	if regexp.MustCompile(`_normal\.\w{2,4}`).MatchString(ret.ProfileImageUrl) {
		ret.ProfileImageUrl = strings.ReplaceAll(ret.ProfileImageUrl, "_normal.", ".")
	}
	ret.BannerImageUrl = apiUser.ProfileBannerURL

	ret.ProfileImageLocalPath = ret.compute_profile_image_local_path()
	ret.BannerImageLocalPath = ret.compute_banner_image_local_path()

	if len(apiUser.PinnedTweetIdsStr) > 0 {
		ret.PinnedTweetID = TweetID(idstr_to_int(apiUser.PinnedTweetIdsStr[0]))
	}
	return
}

// Calls API#GetUser and returns the parsed result
func GetUser(handle UserHandle) (User, error) {
	session, err := NewGuestSession() // This endpoint works better if you're not logged in
	if err != nil {
		return User{}, err
	}
	apiUser, err := session.GetUser(handle)
	if apiUser.ScreenName == "" {
		if apiUser.IsBanned || apiUser.DoesntExist {
			ret := GetUnknownUserWithHandle(handle)
			ret.IsBanned = apiUser.IsBanned
			ret.IsDeleted = apiUser.DoesntExist
			return ret, nil
		}
		apiUser.ScreenName = string(handle)
	}
	if err != nil {
		return User{}, fmt.Errorf("Error fetching user %q:\n  %w", handle, err)
	}
	return ParseSingleUser(apiUser)
}

/**
 * Make a filename for the profile image, that hopefully won't clobber other ones
 */
func (u User) compute_profile_image_local_path() string {
	return string(u.Handle) + "_profile_" + path.Base(u.ProfileImageUrl)
}

/**
 * Make a filename for the banner image, that hopefully won't clobber other ones.
 * Add a file extension if necessary (seems to be necessary).
 * If there is no banner image, just return nothing.
 */
func (u User) compute_banner_image_local_path() string {
	if u.BannerImageUrl == "" {
		return ""
	}
	base_name := path.Base(u.BannerImageUrl)

	// Check if it has an extension (e.g., ".png" or ".jpeg")
	if !regexp.MustCompile(`\.\w{2,4}$`).MatchString(base_name) {
		// If it doesn't have an extension, add one
		base_name += ".jpg"
	}
	return string(u.Handle) + "_banner_" + base_name
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
