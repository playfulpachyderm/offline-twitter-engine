package persistence

import (
	"database/sql"
	"errors"
	"fmt"
	"gitlab.com/offline-twitter/twitter_offline_engine/pkg/scraper"
)

type SpaceParticipant struct {
	UserID  scraper.UserID  `db:"user_id"`
	SpaceID scraper.SpaceID `db:"space_id"`
}

// Save a Space
func (p Profile) SaveSpace(s scraper.Space) error {
	_, err := p.DB.NamedExec(`
		insert into spaces (id, created_by_id, short_url, state, title, created_at, started_at, ended_at, updated_at,
		                    is_available_for_replay, replay_watch_count, live_listeners_count, is_details_fetched)
		values (:id, nullif(:created_by_id, 0), :short_url, :state, :title, :created_at, :started_at, :ended_at, :updated_at,
			    :is_available_for_replay, :replay_watch_count, :live_listeners_count, :is_details_fetched)
		    on conflict do update
		   set id=:id,
		       created_by_id=case when created_by_id is not null then created_by_id else nullif(:created_by_id, 0) end,
		       short_url=case when short_url == "" then :short_url else short_url end,
		       state=case when :state != "" then :state else state end,
		       title=case when :is_details_fetched then :title else title end,
		       updated_at=max(:updated_at, updated_at),
		       ended_at=max(:ended_at, ended_at),
		       is_available_for_replay=:is_available_for_replay,
		       replay_watch_count=:replay_watch_count,
		       live_listeners_count=max(:live_listeners_count, live_listeners_count),
		       is_details_fetched=(is_details_fetched or :is_details_fetched)
	`, &s)
	if err != nil {
		return fmt.Errorf("Error saving space (space ID %q, value: %#v):\n  %w", s.ID, s, err)
	}

	space_participants := []SpaceParticipant{}
	for _, participant_id := range s.ParticipantIds {
		space_participants = append(space_participants, SpaceParticipant{UserID: participant_id, SpaceID: s.ID})
	}
	if len(space_participants) > 0 {
		_, err = p.DB.NamedExec(`
			insert or replace into space_participants (user_id, space_id) values (:user_id, :space_id)
		`, space_participants)
		if err != nil {
			return fmt.Errorf("Error saving participants (space ID %q, participants: %#v):\n  %w", s.ID, space_participants, err)
		}
	}
	return nil
}

// Get a Space by ID
func (p Profile) GetSpaceById(id scraper.SpaceID) (space scraper.Space, err error) {
	err = p.DB.Get(&space,
		`select id, created_by_id, short_url, state, title, created_at, started_at, ended_at, updated_at, is_available_for_replay,
	            replay_watch_count, live_listeners_count, is_details_fetched
	       from spaces
	      where id = ?`, id)
	if err != nil {
		return
	}
	space.ParticipantIds = []scraper.UserID{}
	rows, err := p.DB.Query(`select user_id from space_participants where space_id = ?`, id)
	if errors.Is(err, sql.ErrNoRows) {
		err = nil
	}
	defer rows.Close()
	if err != nil {
		panic(err)
	}
	var participant_id scraper.UserID
	for rows.Next() {
		err = rows.Scan(&participant_id)
		if err != nil {
			panic(err)
		}
		space.ParticipantIds = append(space.ParticipantIds, participant_id)
	}

	return
}
