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
	StartedAt            Timestamp `db:"started_at"`
	EndedAt              Timestamp `db:"ended_at"`
	UpdatedAt            Timestamp `db:"updated_at"`
	IsAvailableForReplay bool      `db:"is_available_for_replay"`
	ReplayWatchCount     int       `db:"replay_watch_count"`
	LiveListenersCount   int       `db:"live_listeners_count"`
	ParticipantIds       []UserID

	CreatedById UserID `db:"created_by_id"`
	TweetID     TweetID

	IsDetailsFetched bool `db:"is_details_fetched"`
}

// TODO: view-layer
// - view helpers should go in a view layer

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
