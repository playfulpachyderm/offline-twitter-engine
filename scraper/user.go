package scraper

import (
    "time"
    "fmt"
    "strings"
    "regexp"
    "path"

    "offline_twitter/terminal_utils"
)

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
    ProfileImageUrl       string
    ProfileImageLocalPath string
    BannerImageUrl        string
    BannerImageLocalPath  string

    PinnedTweetID   TweetID
    PinnedTweet     *Tweet

    IsContentDownloaded bool
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

// Turn an APIUser, as returned from the scraper, into a properly structured User object
func ParseSingleUser(apiUser APIUser) (ret User, err error) {
    ret.ID = UserID(apiUser.ID)
    ret.DisplayName = apiUser.Name
    ret.Handle = UserHandle(apiUser.ScreenName)
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
    match, err := regexp.MatchString("\\.\\w{2,4}$", base_name)
    if err != nil {
        panic(err)
    }
    // If it doesn't have an extension, add one
    if !match {
        base_name += ".jpg"
    }
    return string(u.Handle) + "_banner_" + base_name
}
