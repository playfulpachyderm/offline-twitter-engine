package persistence

import (
	"database/sql"
	"errors"
	"fmt"
)

func (p Profile) SaveNotification(n Notification) {
	tx, err := p.DB.Beginx()
	defer func() {
		if r := recover(); r != nil {
			err := tx.Rollback()
			if err != nil {
				panic(err)
			}
			panic(r) // Re-raise the panic
		}
	}()

	if err != nil {
		panic(err)
	}

	// Save the Notification
	_, err = tx.NamedExec(`
		insert into notifications(id, type, sent_at, sort_index, user_id, action_user_id, action_tweet_id, action_retweet_id,
			                      has_detail, last_scraped_at)
		     values (:id, :type, :sent_at, :sort_index, :user_id, nullif(:action_user_id, 0), nullif(:action_tweet_id, 0),
		            nullif(:action_retweet_id, 0), :has_detail, :last_scraped_at)
		         on conflict do update
		        set sent_at = max(sent_at, :sent_at),
		            sort_index = max(sort_index, :sort_index),
		            action_user_id = nullif(:action_user_id, 0),
		            action_tweet_id = nullif(:action_tweet_id, 0),
		            has_detail = has_detail or :has_detail,
		            last_scraped_at = max(last_scraped_at, :last_scraped_at)
	`, n)
	if err != nil {
		fmt.Printf("failed to save notification %#v\n", n)
		panic(err)
	}

	// Save relevant users and tweets
	for _, u_id := range n.UserIDs {
		_, err = tx.Exec(`
			insert into notification_users(notification_id, user_id) values (?, ?) on conflict do nothing
		`, n.ID, u_id)
		if err != nil {
			panic(err)
		}
	}
	for _, t_id := range n.TweetIDs {
		_, err = tx.Exec(`
			insert into notification_tweets(notification_id, tweet_id) values (?, ?) on conflict do nothing
		`, n.ID, t_id)
		if err != nil {
			fmt.Printf("failed to save notification %#v\n", n)
			panic(err)
		}
	}
	for _, r_id := range n.RetweetIDs {
		_, err = tx.Exec(`
			insert into notification_retweets(notification_id, retweet_id) values (?, ?) on conflict do nothing
		`, n.ID, r_id)
		if err != nil {
			fmt.Printf("failed to save notification %#v\n", n)
			panic(err)
		}
	}
	err = tx.Commit()
	if err != nil {
		panic(err)
	}
}

func (p Profile) GetNotification(id NotificationID) Notification {
	var ret Notification
	err := p.DB.Get(&ret,
		`select id, type, sent_at, sort_index, user_id, ifnull(action_user_id, 0) action_user_id,
		        ifnull(action_tweet_id, 0) action_tweet_id, ifnull(action_retweet_id, 0) action_retweet_id, has_detail, last_scraped_at
		   from notifications where id = ?`,
		id)
	if err != nil {
		panic(err)
	}
	err = p.DB.Select(&ret.UserIDs, `select user_id from notification_users where notification_id = ?`, id)
	if err != nil {
		panic(err)
	}
	err = p.DB.Select(&ret.TweetIDs, `select tweet_id from notification_tweets where notification_id = ?`, id)
	if err != nil {
		panic(err)
	}
	err = p.DB.Select(&ret.RetweetIDs, `select retweet_id from notification_retweets where notification_id = ?`, id)
	if err != nil {
		panic(err)
	}
	return ret
}

func (p Profile) CheckNotificationScrapesNeeded(trove TweetTrove) []NotificationID {
	ret := []NotificationID{}
	for n_id, notification := range trove.Notifications {
		// If there's no detail page, skip
		if !notification.HasDetail {
			continue
		}

		// Check its last-scraped
		var last_scraped_at Timestamp
		err := p.DB.Get(&last_scraped_at, `select last_scraped_at from notifications where id = ?`, n_id)
		if errors.Is(err, sql.ErrNoRows) {
			// It's not scraped at all yet
			ret = append(ret, n_id)
			continue
		} else if err != nil {
			panic(err)
		}
		// If the latest scrape is not fresh (older than the notification sent-at time), add it
		if last_scraped_at.Time.Before(notification.SentAt.Time) {
			ret = append(ret, n_id)
		}
	}
	return ret
}

func (p Profile) GetUnreadNotificationsCount(u_id UserID, since_sort_index int64) int {
	var ret int
	err := p.DB.Get(&ret, `select count(*) from notifications where sort_index > ? and user_id = ?`, since_sort_index, u_id)
	if err != nil {
		panic(err)
	}
	return ret
}
