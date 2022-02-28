package scraper

import (
    "time"
    "fmt"
    "strings"
    "regexp"
    "path"

    "offline_twitter/terminal_utils"
)

const DEFAULT_PROFILE_IMAGE_URL = "https://abs.twimg.com/sticky/default_profile_images/default_profile.png"

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
    ID                    UserID
    DisplayName           string
    Handle                UserHandle
    Bio                   string
    FollowingCount        int
    FollowersCount        int
    Location              string
    Website               string
    JoinDate              time.Time
    IsPrivate             bool
    IsVerified            bool
    IsBanned              bool
    ProfileImageUrl       string
    ProfileImageLocalPath string
    BannerImageUrl        string
    BannerImageLocalPath  string

    PinnedTweetID   TweetID
    PinnedTweet     *Tweet

    IsFollowed          bool
    IsContentDownloaded bool
    IsNeedingFakeID     bool
    IsIdFake            bool
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
        terminal_utils.FormatDate(u.JoinDate),
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

/**
 * Given a tweet URL, return the corresponding user handle.
 * If tweet url is not valid, return an error.
 */
func ParseHandleFromTweetUrl(tweet_url string) (UserHandle, error) {
    short_url_regex := regexp.MustCompile(`^https://t.co/\w{5,20}$`)
    if short_url_regex.MatchString(tweet_url) {
        tweet_url = ExpandShortUrl(tweet_url)
    }

    r := regexp.MustCompile(`^https://twitter.com/(\w+)/status/\d+(?:\?.*)?$`)
    matches := r.FindStringSubmatch(tweet_url)
    if len(matches) != 2 {  // matches[0] is the full string
        return "", fmt.Errorf("Invalid tweet url: %s", tweet_url)
    }
    return UserHandle(matches[1]), nil
}

func GetUnknownUserWithHandle(handle UserHandle) User {
    return User{
        ID: UserID(0),  // 2^62 + 1...
        DisplayName: string(handle),
        Handle: handle,
        Bio: "<blank>",
        FollowersCount: 0,
        FollowingCount: 0,
        Location: "<blank>",
        Website:"<blank>",
        JoinDate: time.Unix(0, 0),
        IsVerified: false,
        IsPrivate: true,
        IsNeedingFakeID: true,
        IsIdFake: true,
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
    ret.JoinDate, err = time.Parse(time.RubyDate, apiUser.CreatedAt)
    if err != nil {
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
    api := API{}
    apiUser, err := api.GetUser(handle)
    if apiUser.ScreenName == "" {
        apiUser.ScreenName = string(handle)
    }
    if err != nil {
        return User{}, err
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
        panic(fmt.Sprintf("Weird profile image url: %s", u.ProfileImageUrl))
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
    return string(u.Handle) + "_profile_" + path.Base(u.GetTinyProfileImageUrl())
}
