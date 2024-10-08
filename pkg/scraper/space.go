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
	ReplayWatchCount     int
	LiveListenersCount   int
	ParticipantIds       []UserID

	CreatedById UserID
	TweetID     TweetID

	IsDetailsFetched bool
}

func (space Space) FormatDuration() string {
	duration := space.EndedAt.Time.Sub(space.StartedAt.Time)
	h := int(duration.Hours())
	m := int(duration.Minutes()) % 60
	s := int(duration.Seconds()) % 60

	if h != 0 {
		return fmt.Sprintf("%dh%02dm", h, m)
	}
	return fmt.Sprintf("%dm%02ds", m, s)
}

func ParseAPISpace(apiCard APICard) Space {
	ret := Space{}
	ret.ID = SpaceID(apiCard.BindingValues.ID.StringValue)
	ret.ShortUrl = apiCard.ShortenedUrl

	// Indicate that this Space needs its details fetched still
	ret.IsDetailsFetched = false

	return ret
}
