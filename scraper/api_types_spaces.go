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
				TotalReplayWatched          int64  `json:"total_replay_watched"`
				TotalLiveListeners          int64  `json:"total_live_listeners"`
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
	// TODO: break up this URL into params so it's readable
	url, err := url.Parse("https://twitter.com/i/api/graphql/Ha9BKBF0uAz9d4-lz0jnYA/AudioSpaceById?variables=%7B%22id%22%3A%22" + string(id) + "%22%2C%22isMetatagsQuery%22%3Afalse%2C%22withSuperFollowsUserFields%22%3Atrue%2C%22withDownvotePerspective%22%3Afalse%2C%22withReactionsMetadata%22%3Afalse%2C%22withReactionsPerspective%22%3Afalse%2C%22withSuperFollowsTweetFields%22%3Atrue%2C%22withReplays%22%3Atrue%7D&features=%7B%22spaces_2022_h2_clipping%22%3Atrue%2C%22spaces_2022_h2_spaces_communities%22%3Atrue%2C%22responsive_web_twitter_blue_verified_badge_is_enabled%22%3Atrue%2C%22verified_phone_label_enabled%22%3Afalse%2C%22tweetypie_unmention_optimization_enabled%22%3Atrue%2C%22responsive_web_uc_gql_enabled%22%3Atrue%2C%22vibe_api_enabled%22%3Atrue%2C%22responsive_web_edit_tweet_api_enabled%22%3Atrue%2C%22graphql_is_translatable_rweb_tweet_is_translatable_enabled%22%3Atrue%2C%22standardized_nudges_misinfo%22%3Atrue%2C%22tweet_with_visibility_results_prefer_gql_limited_actions_policy_enabled%22%3Afalse%2C%22responsive_web_graphql_timeline_navigation_enabled%22%3Atrue%2C%22interactive_text_enabled%22%3Atrue%2C%22responsive_web_text_conversations_enabled%22%3Afalse%2C%22responsive_web_enhance_cards_enabled%22%3Atrue%7D") //nolint:lll  // It's a URL, come on
	if err != nil {
		panic(err)
	}

	var result SpaceResponse
	err = api.do_http(url.String(), "", &result)
	return result, err
}
