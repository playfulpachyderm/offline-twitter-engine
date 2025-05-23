package scraper

import (
	"errors"
	"fmt"
	"net/url"
	"path/filepath"
	"regexp"

	. "gitlab.com/offline-twitter/twitter_offline_engine/pkg/persistence"
)

type UserResponse struct {
	Data struct {
		User _UserResults `json:"user"`
	} `json:"data"`
	Errors []struct {
		Message string `json:"message"`
		Name    string `json:"name"`
		Code    int    `json:"code"`
	} `json:"errors"`
}

func (u UserResponse) ConvertToAPIUser() (APIUser, error) {
	if u.Data.User.Result.MetaTypename == "" {
		// Completely empty response (user not found)
		return APIUser{}, ErrDoesntExist
	}

	ret := u.Data.User.Result.Legacy
	ret.ID = u.Data.User.Result.ID
	ret.Verified = u.Data.User.Result.IsBlueVerified

	// Banned users
	for _, api_error := range u.Errors {
		if api_error.Message == "Authorization: User has been suspended. (63)" {
			ret.IsBanned = true
		} else if api_error.Name == "NotFoundError" {
			// TODO: not sure what kind of request returns this
			ret.DoesntExist = true
		} else {
			panic(fmt.Errorf("Unknown api error %q:\n  %w", api_error.Message, ErrExternalApiError))
		}
	}

	// Banned users, new version
	if u.Data.User.Result.Reason == "Suspended" {
		ret.IsBanned = true
	}

	// Deleted users
	if ret.ID == 0 && ret.ScreenName == "" && u.Data.User.Result.Reason != "Suspended" {
		ret.DoesntExist = true
	}

	return ret, nil
}

func (api API) GetUser(handle UserHandle) (User, error) {
	url, err := url.Parse(GraphqlURL{
		BaseUrl: "https://api.twitter.com/graphql/SAMkL5y_N9pmahSw8yy6gw/UserByScreenName",
		Variables: GraphqlVariables{
			ScreenName:                  handle,
			Count:                       20,
			IncludePromotedContent:      false,
			WithSuperFollowsUserFields:  true,
			WithDownvotePerspective:     false,
			WithReactionsMetadata:       false,
			WithReactionsPerspective:    false,
			WithSuperFollowsTweetFields: true,
			WithBirdwatchNotes:          false,
			WithVoice:                   true,
			WithV2Timeline:              false,
		},
		Features: GraphqlFeatures{
			ResponsiveWebTwitterBlueVerifiedBadgeIsEnabled:                 true,
			VerifiedPhoneLabelEnabled:                                      false,
			ResponsiveWebGraphqlTimelineNavigationEnabled:                  true,
			UnifiedCardsAdMetadataContainerDynamicCardContentQueryEnabled:  true,
			TweetypieUnmentionOptimizationEnabled:                          true,
			ResponsiveWebUcGqlEnabled:                                      true,
			VibeApiEnabled:                                                 true,
			ResponsiveWebEditTweetApiEnabled:                               true,
			GraphqlIsTranslatableRWebTweetIsTranslatableEnabled:            true,
			StandardizedNudgesMisinfo:                                      true,
			TweetWithVisibilityResultsPreferGqlLimitedActionsPolicyEnabled: false,
			InteractiveTextEnabled:                                         true,
			ResponsiveWebTextConversationsEnabled:                          false,
			ResponsiveWebEnhanceCardsEnabled:                               true,
		},
	}.String())
	if err != nil {
		panic(err)
	}

	var response UserResponse
	err = api.do_http(url.String(), "", &response)
	if err != nil {
		return User{}, err
	}
	apiUser, err := response.ConvertToAPIUser()
	if errors.Is(err, ErrDoesntExist) {
		return User{}, err
	}
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

// Calls API#GetUserByID and returns the parsed result
func GetUserByID(u_id UserID) (User, error) {
	session, err := NewGuestSession() // This endpoint works better if you're not logged in
	if err != nil {
		return User{}, err
	}
	return session.GetUserByID(u_id)
}

func (api API) GetUserByID(u_id UserID) (User, error) {
	if u_id == UserID(0) {
		panic("No Users with ID 0")
	}
	url, err := url.Parse(GraphqlURL{
		BaseUrl: "https://x.com/i/api/graphql/Qw77dDjp9xCpUY-AXwt-yQ/UserByRestId",
		Variables: GraphqlVariables{
			UserID: u_id,
		},
		Features: GraphqlFeatures{
			RWebTipjarConsumptionEnabled:                              true,
			ResponsiveWebGraphqlExcludeDirectiveEnabled:               true,
			VerifiedPhoneLabelEnabled:                                 false,
			ResponsiveWebGraphqlSkipUserProfileImageExtensionsEnabled: false,
			ResponsiveWebGraphqlTimelineNavigationEnabled:             true,
			SubscriptionsFeatureCanGiftPremium:                        true,
			ResponsiveWebTwitterArticleNotesTabEnabled:                true,
		},
	}.String())
	if err != nil {
		panic(err)
	}

	var response UserResponse
	err = api.do_http(url.String(), "", &response)
	if err != nil {
		return User{}, err
	}
	apiUser, err := response.ConvertToAPIUser()
	if errors.Is(err, ErrDoesntExist) {
		return User{}, err
	}
	if apiUser.ScreenName == "" {
		if apiUser.IsBanned {
			return User{}, ErrUserIsBanned
		} else {
			return User{}, ErrDoesntExist
		}
	}
	if err != nil {
		return User{}, fmt.Errorf("Error fetching user ID %d:\n  %w", u_id, err)
	}
	return ParseSingleUser(apiUser)
}

// Make a filename for the profile image, that hopefully won't clobber other ones
func compute_profile_image_local_path(u User) string {
	return string(u.Handle) + "_profile_" + filepath.Base(u.ProfileImageUrl)
}

// Make a filename for the banner image, that hopefully won't clobber other ones.
// Add a file extension if necessary (seems to be necessary).
// If there is no banner image, just return nothing.
func compute_banner_image_local_path(u User) string {
	if u.BannerImageUrl == "" {
		return ""
	}
	base_name := filepath.Base(u.BannerImageUrl)

	// Check if it has an extension (e.g., ".png" or ".jpeg")
	if !regexp.MustCompile(`\.\w{2,4}$`).MatchString(base_name) {
		// If it doesn't have an extension, add one
		base_name += ".jpg"
	}
	return string(u.Handle) + "_banner_" + base_name
}
