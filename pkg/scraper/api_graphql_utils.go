package scraper

import (
	"encoding/json"
	"net/url"

	. "gitlab.com/offline-twitter/twitter_offline_engine/pkg/persistence"
)

type GraphqlVariables struct {
	UserID                                 UserID     `json:"userId,string,omitempty"`
	ScreenName                             UserHandle `json:"screen_name,omitempty"`
	RawQuery                               string     `json:"rawQuery,omitempty"`
	Product                                string     `json:"product,omitempty"`
	FocalTweetID                           TweetID    `json:"focalTweetId,string,omitempty"`
	Cursor                                 string     `json:"cursor,omitempty"`
	WithRuxInjections                      bool       `json:"with_rux_injections"`
	IncludePromotedContent                 bool       `json:"includePromotedContent"`
	Count                                  int        `json:"count,omitempty"`
	WithCommunity                          bool       `json:"withCommunity"`
	WithQuickPromoteEligibilityTweetFields bool       `json:"withQuickPromoteEligibilityTweetFields"`
	WithSuperFollowsUserFields             bool       `json:"withSuperFollowsUserFields,omitempty"`
	WithBirdwatchPivots                    bool       `json:"withBirdwatchPivots"`
	WithBirdwatchNotes                     bool       `json:"withBirdwatchNotes,omitempty"`
	WithDownvotePerspective                bool       `json:"withDownvotePerspective"`
	WithReactionsMetadata                  bool       `json:"withReactionsMetadata"`
	WithReactionsPerspective               bool       `json:"withReactionsPerspective"`
	WithSuperFollowsTweetFields            bool       `json:"withSuperFollowsTweetFields,omitempty"`
	WithVoice                              bool       `json:"withVoice"`
	WithV2Timeline                         bool       `json:"withV2Timeline"`
	FSInteractiveText                      bool       `json:"__fs_interactive_text,omitempty"`
	FSResponsiveWebUCGqlEnabled            bool       `json:"__fs_responsive_web_uc_gql_enabled,omitempty"`
	FSDontMentionMeViewApiEnabled          bool       `json:"__fs_dont_mention_me_view_api_enabled,omitempty"`

	// Spaces
	ID              SpaceID `json:"id"`
	IsMetatagsQuery bool    `json:"isMetatagsQuery"`
	WithReplays     bool    `json:"withReplays"`
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
	ResponsiveWebTwitterArticleTweetConsumptionEnabled             bool `json:"responsive_web_twitter_article_tweet_consumption_enabled"`
	ResponsiveWebMediaDownloadVideoEnabled                         bool `json:"responsive_web_media_download_video_enabled"`
	ResponsiveWebTwitterArticleNotesTabEnabled                     bool `json:"responsive_web_twitter_article_notes_tab_enabled"`
	SubscriptionsVerificationInfoVerifiedSinceEnabled              bool `json:"subscriptions_verification_info_verified_since_enabled"`
	HiddenProfileLikesEnabled                                      bool `json:"hidden_profile_likes_enabled"`
	HiddenProfileSubscriptionsEnabled                              bool `json:"hidden_profile_subscriptions_enabled"`
	HighlightsTweetsTabUIEnabled                                   bool `json:"highlights_tweets_tab_ui_enabled"`
	SubscriptionsVerificationInfoIsIdentityVerifiedEnabled         bool `json:"subscriptions_verification_info_is_identity_verified_enabled"` //nolint:lll // I didn't choose this field name
	C9sTweetAnatomyModeratorBadgeEnabled                           bool `json:"c9s_tweet_anatomy_moderator_badge_enabled"`
	RwebVideoTimestampsEnabled                                     bool `json:"rweb_video_timestamps_enabled"`
	PremiumContentApiReadEnabled                                   bool `json:"premium_content_api_read_enabled"`
	ResponsiveWebGrokAnalysisButtonFromBackend                     bool `json:"responsive_web_grok_analysis_button_from_backend"`
	ProfileLabelImprovementsPcfLabelInPostEnabled                  bool `json:"profile_label_improvements_pcf_label_in_post_enabled"`
	ResponsiveWebJetfuelFrame                                      bool `json:"responsive_web_jetfuel_frame"`
	RwebVideoScreenEnabled                                         bool `json:"rweb_video_screen_enabled"`

	// Spaces
	Spaces2022H2Clipping          bool `json:"spaces_2022_h2_clipping,omitempty"`
	Spaces2022H2SpacesCommunities bool `json:"spaces_2022_h2_spaces_communities,omitempty"`

	// Bookmarks
	CommunitiesWebEnableTweetCommunityResultsFetch bool `json:"communities_web_enable_tweet_community_results_fetch,omitempty"`
	RWebTipjarConsumptionEnabled                   bool `json:"rweb_tipjar_consumption_enabled"`
	ArticlesPreviewEnabled                         bool `json:"articles_preview_enabled"`
	GraphqlTimelineV2BookmarkTimeline              bool `json:"graphql_timeline_v2_bookmark_timeline,omitempty"`
	CreatorSubscriptionsQuoteTweetPreviewEnabled   bool `json:"creator_subscriptions_quote_tweet_preview_enabled"`
	SubscriptionsFeatureCanGiftPremium             bool `json:"subscriptions_feature_can_gift_premium,omitempty"`

	// Grok stuff
	ResponsiveWebGrokShareAttachmentEnabled          bool `json:"responsive_web_grok_share_attachment_enabled"`
	ResponsiveWebGrokImageAnnotationEnabled          bool `json:"responsive_web_grok_image_annotation_enabled"`
	ResponsiveWebGrokAnalyzeButtonFetchTrendsEnabled bool `json:"responsive_web_grok_analyze_button_fetch_trends_enabled"`
	ResponsiveWebGrokAnalyzePostFollowupsEnabled     bool `json:"responsive_web_grok_analyze_post_followups_enabled"`
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
