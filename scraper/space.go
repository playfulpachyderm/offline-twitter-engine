package scraper

import (
	"fmt"
)

type SpaceID string

type Space struct {
	ID                   SpaceID   `db:"id"`
	ShortUrl             string    `db:"short_url"`
	State                string    `db:"state"`
	Title                string    `db:"title"`
	CreatedAt            Timestamp `db:"created_at"`
	StartedAt            Timestamp
	EndedAt              Timestamp `db:"ended_at"`
	UpdatedAt            Timestamp
	IsAvailableForReplay bool
	ReplayWatchCount     int64
	LiveListenersCount   int64
	ParticipantIds       []UserID

	CreatedById UserID
	TweetID     TweetID

	IsDetailsFetched bool
}

func ParseAPISpace(apiCard APICard) Space {
	ret := Space{}
	ret.ID = SpaceID(apiCard.BindingValues.ID.StringValue)
	ret.ShortUrl = apiCard.ShortenedUrl

	// Indicate that this Space needs its details fetched still
	ret.IsDetailsFetched = false

	return ret
}

func FetchSpaceDetail(id SpaceID) (TweetTrove, error) {
	space_response, err := the_api.GetSpace(id)
	if err != nil {
		return TweetTrove{}, fmt.Errorf("Error in API call to fetch Space (id %q):\n  %w", id, err)
	}
	return space_response.ToTweetTrove(), nil
}
