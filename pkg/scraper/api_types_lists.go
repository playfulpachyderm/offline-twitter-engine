package scraper

import (
	"net/url"
)

func (api *API) GetFollowees(user_id UserID, cursor string) (APIV2Response, error) {
	url, err := url.Parse(GraphqlURL{
		BaseUrl: "https://twitter.com/i/api/graphql/0yD6Eiv23DKXRDU9VxlG2A/Following",
		Variables: GraphqlVariables{
			UserID:                 user_id,
			Cursor:                 cursor,
			Count:                  20,
			IncludePromotedContent: false,
		},
		Features: GraphqlFeatures{
			ResponsiveWebGraphqlExcludeDirectiveEnabled:                    true,
			VerifiedPhoneLabelEnabled:                                      false,
			CreatorSubscriptionsTweetPreviewApiEnabled:                     true,
			ResponsiveWebGraphqlTimelineNavigationEnabled:                  true,
			ResponsiveWebGraphqlSkipUserProfileImageExtensionsEnabled:      false,
			C9sTweetAnatomyModeratorBadgeEnabled:                           true,
			TweetypieUnmentionOptimizationEnabled:                          true,
			ResponsiveWebEditTweetApiEnabled:                               true,
			GraphqlIsTranslatableRWebTweetIsTranslatableEnabled:            true,
			ViewCountsEverywhereApiEnabled:                                 true,
			LongformNotetweetsConsumptionEnabled:                           true,
			ResponsiveWebTwitterArticleTweetConsumptionEnabled:             false,
			TweetAwardsWebTippingEnabled:                                   false,
			FreedomOfSpeechNotReachFetchEnabled:                            true,
			StandardizedNudgesMisinfo:                                      true,
			TweetWithVisibilityResultsPreferGqlLimitedActionsPolicyEnabled: true,
			RwebVideoTimestampsEnabled:                                     true,
			LongformNotetweetsRichTextReadEnabled:                          true,
			LongformNotetweetsInlineMediaEnabled:                           true,
			ResponsiveWebMediaDownloadVideoEnabled:                         false,
			ResponsiveWebEnhanceCardsEnabled:                               false,
		},
	}.String())
	if err != nil {
		panic(err)
	}

	var result APIV2Response
	err = api.do_http(url.String(), "", &result)
	return result, err
}

type PaginatedFollowees struct {
	user_id UserID
}

func (p PaginatedFollowees) NextPage(api *API, cursor string) (APIV2Response, error) {
	return api.GetFollowees(p.user_id, cursor)
}
func (p PaginatedFollowees) ToTweetTrove(r APIV2Response) (TweetTrove, error) {
	return r.ToTweetTrove()
}

func GetFollowees(user_id UserID, how_many int) (TweetTrove, error) {
	return the_api.GetPaginatedQuery(PaginatedFollowees{user_id}, how_many)
}

func (api *API) GetFollowers(user_id UserID, cursor string) (APIV2Response, error) {
	url, err := url.Parse(GraphqlURL{
		BaseUrl: "https://twitter.com/i/api/graphql/3_7xfjmh897x8h_n6QBqTA/Followers",
		Variables: GraphqlVariables{
			UserID:                 user_id,
			Cursor:                 cursor,
			Count:                  20,
			IncludePromotedContent: false,
		},
		Features: GraphqlFeatures{
			ResponsiveWebGraphqlExcludeDirectiveEnabled:                    true,
			VerifiedPhoneLabelEnabled:                                      false,
			CreatorSubscriptionsTweetPreviewApiEnabled:                     true,
			ResponsiveWebGraphqlTimelineNavigationEnabled:                  true,
			ResponsiveWebGraphqlSkipUserProfileImageExtensionsEnabled:      false,
			C9sTweetAnatomyModeratorBadgeEnabled:                           true,
			TweetypieUnmentionOptimizationEnabled:                          true,
			ResponsiveWebEditTweetApiEnabled:                               true,
			GraphqlIsTranslatableRWebTweetIsTranslatableEnabled:            true,
			ViewCountsEverywhereApiEnabled:                                 true,
			LongformNotetweetsConsumptionEnabled:                           true,
			ResponsiveWebTwitterArticleTweetConsumptionEnabled:             false,
			TweetAwardsWebTippingEnabled:                                   false,
			FreedomOfSpeechNotReachFetchEnabled:                            true,
			StandardizedNudgesMisinfo:                                      true,
			TweetWithVisibilityResultsPreferGqlLimitedActionsPolicyEnabled: true,
			RwebVideoTimestampsEnabled:                                     true,
			LongformNotetweetsRichTextReadEnabled:                          true,
			LongformNotetweetsInlineMediaEnabled:                           true,
			ResponsiveWebMediaDownloadVideoEnabled:                         false,
			ResponsiveWebEnhanceCardsEnabled:                               false,
		},
	}.String())
	if err != nil {
		panic(err)
	}

	var result APIV2Response
	err = api.do_http(url.String(), "", &result)
	return result, err
}

type PaginatedFollowers struct {
	user_id UserID
}

func (p PaginatedFollowers) NextPage(api *API, cursor string) (APIV2Response, error) {
	return api.GetFollowers(p.user_id, cursor)
}
func (p PaginatedFollowers) ToTweetTrove(r APIV2Response) (TweetTrove, error) {
	return r.ToTweetTrove()
}

func GetFollowers(user_id UserID, how_many int) (TweetTrove, error) {
	return the_api.GetPaginatedQuery(PaginatedFollowers{user_id}, how_many)
}
