package scraper

import (
	"net/url"
)

type SpaceResponse struct {
	Data struct {
		AudioSpace struct {
			Metadata struct {
				RestId                      string `json:"rest_id"`
				State                       string
				Title                       string
				MediaKey                    string `json:"media_key"`
				CreatedAt                   int64  `json:"created_at"`
				StartedAt                   int64  `json:"started_at"`
				EndedAt                     int64  `json:"ended_at,string"`
				UpdatedAt                   int64  `json:"updated_at"`
				DisallowJoin                bool   `json:"disallow_join"`
				NarrowCastSpaceType         int64  `json:"narrow_cast_space_type"`
				IsEmployeeOnly              bool   `json:"is_employee_only"`
				IsLocked                    bool   `json:"is_locked"`
				IsSpaceAvailableForReplay   bool   `json:"is_space_available_for_replay"`
				IsSpaceAvailableForClipping bool   `json:"is_space_available_for_clipping"`
				ConversationControls        int64  `json:"conversation_controls"`
				TotalReplayWatched          int    `json:"total_replay_watched"`
				TotalLiveListeners          int    `json:"total_live_listeners"`
				CreatorResults              struct {
					Result struct {
						ID     int64   `json:"rest_id,string"`
						Legacy APIUser `json:"legacy"`
					} `json:"result"`
				} `json:"creator_results"`
			}
			Participants struct {
				Total  int
				Admins []struct {
					Start int
					User  struct {
						RestId int64 `json:"rest_id,string"`
					}
				}
				Speakers []struct {
					User struct {
						RestId int64 `json:"rest_id,string"`
					}
				}
			}
		}
	}
}

func (r SpaceResponse) ToTweetTrove() TweetTrove {
	data := r.Data.AudioSpace

	ret := NewTweetTrove()
	space := Space{}
	space.ID = SpaceID(data.Metadata.RestId)
	if space.ID == "" {
		// The response is empty.  Abort processing
		return ret
	}

	space.Title = data.Metadata.Title
	space.State = data.Metadata.State
	space.CreatedById = UserID(data.Metadata.CreatorResults.Result.ID)
	space.CreatedAt = TimestampFromUnix(data.Metadata.CreatedAt)
	space.StartedAt = TimestampFromUnix(data.Metadata.StartedAt)
	space.EndedAt = TimestampFromUnix(data.Metadata.EndedAt)
	space.UpdatedAt = TimestampFromUnix(data.Metadata.UpdatedAt)
	space.IsAvailableForReplay = data.Metadata.IsSpaceAvailableForReplay
	space.ReplayWatchCount = data.Metadata.TotalReplayWatched
	space.LiveListenersCount = data.Metadata.TotalLiveListeners
	space.IsDetailsFetched = true

	for _, admin := range data.Participants.Admins {
		space.ParticipantIds = append(space.ParticipantIds, UserID(admin.User.RestId))
	}
	for _, speaker := range data.Participants.Speakers {
		space.ParticipantIds = append(space.ParticipantIds, UserID(speaker.User.RestId))
	}

	ret.Spaces[space.ID] = space

	creator, err := ParseSingleUser(data.Metadata.CreatorResults.Result.Legacy)
	if err != nil {
		panic(err)
	}
	creator.ID = space.CreatedById
	ret.Users[creator.ID] = creator

	return ret
}

func (api API) GetSpace(id SpaceID) (SpaceResponse, error) {
	url, err := url.Parse(GraphqlURL{
		BaseUrl: "https://twitter.com/i/api/graphql/Ha9BKBF0uAz9d4-lz0jnYA/AudioSpaceById",
		Variables: GraphqlVariables{
			ID:                          id,
			IsMetatagsQuery:             false,
			WithSuperFollowsUserFields:  true,
			WithDownvotePerspective:     false,
			WithReactionsMetadata:       false,
			WithReactionsPerspective:    false,
			WithSuperFollowsTweetFields: true,
			WithReplays:                 true,
		},
		Features: GraphqlFeatures{
			Spaces2022H2Clipping:                                           true,
			Spaces2022H2SpacesCommunities:                                  true,
			ResponsiveWebTwitterBlueVerifiedBadgeIsEnabled:                 true,
			VerifiedPhoneLabelEnabled:                                      false,
			TweetypieUnmentionOptimizationEnabled:                          true,
			ResponsiveWebUcGqlEnabled:                                      true,
			VibeApiEnabled:                                                 true,
			ResponsiveWebEditTweetApiEnabled:                               true,
			GraphqlIsTranslatableRWebTweetIsTranslatableEnabled:            true,
			StandardizedNudgesMisinfo:                                      true,
			TweetWithVisibilityResultsPreferGqlLimitedActionsPolicyEnabled: false,
			ResponsiveWebGraphqlTimelineNavigationEnabled:                  true,
			InteractiveTextEnabled:                                         true,
			ResponsiveWebTextConversationsEnabled:                          false,
			ResponsiveWebEnhanceCardsEnabled:                               true,
		},
	}.String())
	if err != nil {
		panic(err)
	}

	var result SpaceResponse
	err = api.do_http(url.String(), "", &result)
	return result, err
}
