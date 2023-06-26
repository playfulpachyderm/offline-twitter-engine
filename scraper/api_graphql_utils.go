package scraper

import (
	"encoding/json"
	"net/url"
)

type GraphqlVariables struct {
	UserID                                 UserID  `json:"userId,string,omitempty"`
	FocalTweetID                           TweetID `json:"focalTweetId,string,omitempty"`
	Cursor                                 string  `json:"cursor,omitempty"`
	WithRuxInjections                      bool    `json:"with_rux_injections"`
	IncludePromotedContent                 bool    `json:"includePromotedContent"`
	Count                                  int     `json:"count,omitempty"`
	WithCommunity                          bool    `json:"withCommunity"`
	WithQuickPromoteEligibilityTweetFields bool    `json:"withQuickPromoteEligibilityTweetFields"`
	WithSuperFollowsUserFields             bool    `json:"withSuperFollowsUserFields,omitempty"`
	WithBirdwatchPivots                    bool    `json:"withBirdwatchPivots"`
	WithBirdwatchNotes                     bool    `json:"withBirdwatchNotes,omitempty"`
	WithDownvotePerspective                bool    `json:"withDownvotePerspective"`
	WithReactionsMetadata                  bool    `json:"withReactionsMetadata"`
	WithReactionsPerspective               bool    `json:"withReactionsPerspective"`
	WithSuperFollowsTweetFields            bool    `json:"withSuperFollowsTweetFields,omitempty"`
	WithVoice                              bool    `json:"withVoice"`
	WithV2Timeline                         bool    `json:"withV2Timeline"`
	FSInteractiveText                      bool    `json:"__fs_interactive_text,omitempty"`
	FSResponsiveWebUCGqlEnabled            bool    `json:"__fs_responsive_web_uc_gql_enabled,omitempty"`
	FSDontMentionMeViewApiEnabled          bool    `json:"__fs_dont_mention_me_view_api_enabled,omitempty"`
}

type GraphqlFeatures struct {
	ResponsiveWebTwitterBlueVerifiedBadgeIsEnabled                 bool `json:"responsive_web_twitter_blue_verified_badge_is_enabled,omitempty"` //nolint:lll // I didn't choose this field name
	RWebListsTimelineRedesignEnabled                               bool `json:"rweb_lists_timeline_redesign_enabled"`
	ResponsiveWebGraphqlExcludeDirectiveEnabled                    bool `json:"responsive_web_graphql_exclude_directive_enabled"`
	VerifiedPhoneLabelEnabled                                      bool `json:"verified_phone_label_enabled"`
	CreatorSubscriptionsTweetPreviewApiEnabled                     bool `json:"creator_subscriptions_tweet_preview_api_enabled"`
	ResponsiveWebGraphqlTimelineNavigationEnabled                  bool `json:"responsive_web_graphql_timeline_navigation_enabled"`
	ResponsiveWebGraphqlSkipUserProfileImageExtensionsEnabled      bool `json:"responsive_web_graphql_skip_user_profile_image_extensions_enabled"` //nolint:lll // I didn't choose this field name
	TweetypieUnmentionOptimizationEnabled                          bool `json:"tweetypie_unmention_optimization_enabled"`
	ResponsiveWebEditTweetApiEnabled                               bool `json:"responsive_web_edit_tweet_api_enabled"`
	GraphqlIsTranslatableRWebTweetIsTranslatableEnabled            bool `json:"graphql_is_translatable_rweb_tweet_is_translatable_enabled"`
	ViewCountsEverywhereApiEnabled                                 bool `json:"view_counts_everywhere_api_enabled"`
	LongformNotetweetsConsumptionEnabled                           bool `json:"longform_notetweets_consumption_enabled"`
	TweetAwardsWebTippingEnabled                                   bool `json:"tweet_awards_web_tipping_enabled"`
	FreedomOfSpeechNotReachFetchEnabled                            bool `json:"freedom_of_speech_not_reach_fetch_enabled"`
	StandardizedNudgesMisinfo                                      bool `json:"standardized_nudges_misinfo"`
	TweetWithVisibilityResultsPreferGqlLimitedActionsPolicyEnabled bool `json:"tweet_with_visibility_results_prefer_gql_limited_actions_policy_enabled"` //nolint:lll // I didn't choose this field name
	LongformNotetweetsRichTextReadEnabled                          bool `json:"longform_notetweets_rich_text_read_enabled"`
	LongformNotetweetsInlineMediaEnabled                           bool `json:"longform_notetweets_inline_media_enabled"`
	ResponsiveWebEnhanceCardsEnabled                               bool `json:"responsive_web_enhance_cards_enabled"`
	UnifiedCardsAdMetadataContainerDynamicCardContentQueryEnabled  bool `json:"unified_cards_ad_metadata_container_dynamic_card_content_query_enabled,omitempty"` //nolint:lll // I didn't choose this field name
	ResponsiveWebUcGqlEnabled                                      bool `json:"responsive_web_uc_gql_enabled,omitempty"`
	VibeApiEnabled                                                 bool `json:"vibe_api_enabled,omitempty"`
	InteractiveTextEnabled                                         bool `json:"interactive_text_enabled,omitempty"`
	ResponsiveWebTextConversationsEnabled                          bool `json:"responsive_web_text_conversations_enabled"`
}

type GraphqlURL struct {
	BaseUrl   string
	Variables GraphqlVariables
	Features  GraphqlFeatures
}

func (u GraphqlURL) String() string {
	features_bytes, err := json.Marshal(u.Features)
	if err != nil {
		panic(err)
	}
	vars_bytes, err := json.Marshal(u.Variables)
	if err != nil {
		panic(err)
	}

	ret, err := url.Parse(u.BaseUrl)
	if err != nil {
		panic(err)
	}
	q := ret.Query()
	if u.Features != (GraphqlFeatures{}) {
		q.Add("features", string(features_bytes))
	}
	q.Add("variables", string(vars_bytes))
	ret.RawQuery = q.Encode()
	return ret.String()
}
