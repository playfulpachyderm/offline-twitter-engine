package persistence

import (
	. "gitlab.com/offline-twitter/twitter_offline_engine/pkg/scraper"
)

func (p Profile) SaveNotification(n Notification) {
	tx, err := p.DB.Beginx()
	if err != nil {
		panic(err)
	}

	// Save the Notification
	_, err = tx.NamedExec(`
		insert into notifications(id, type, sent_at, sort_index, user_id, action_user_id, action_tweet_id, action_retweet_id)
		     values (:id, :type, :sent_at, :sort_index, :user_id, nullif(:action_user_id, 0), nullif(:action_tweet_id, 0),
		            nullif(:action_retweet_id, 0))
		         on conflict do update
		        set sent_at = max(sent_at, :sent_at),
		            sort_index = max(sort_index, :sort_index),
		            action_user_id = nullif(:action_user_id, 0),
		            action_tweet_id = nullif(:action_tweet_id, 0)
	`, n)
	if err != nil {
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
		        ifnull(action_tweet_id, 0) action_tweet_id, ifnull(action_retweet_id, 0) action_retweet_id
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
	return ret
}
