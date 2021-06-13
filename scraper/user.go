package scraper

import (
    "time"
    "fmt"
)

type UserID string

type User struct {
    ID              UserID
    DisplayName     string
    Handle          string
    Bio             string
    FollowingCount  int
    FollowersCount  int
    Location        string
    Website         string
    JoinDate        time.Time
    IsPrivate       bool
    IsVerified      bool
    ProfileImageUrl string
    BannerImageUrl  string
    PinnedTweet     TweetID
}

func (u User) String() string {
    return fmt.Sprintf("%s (@%s)[%s]: %q", u.DisplayName, u.Handle, u.ID, u.Bio)
}

// Turn an APIUser, as returned from the scraper, into a properly structured User object
func ParseSingleUser(apiUser APIUser) (ret User, err error) {
    ret.ID = UserID(apiUser.IDStr)
    ret.DisplayName = apiUser.Name
    ret.Handle = apiUser.ScreenName
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
    if len(apiUser.PinnedTweetIdsStr) > 0 {
        ret.PinnedTweet = TweetID(apiUser.PinnedTweetIdsStr[0])
    }
    return
}
