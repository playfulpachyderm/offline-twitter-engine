package scraper

import (
	"errors"
	"net/url"
	"regexp"
	"sort"
	"strings"

	log "github.com/sirupsen/logrus"
)

// TODO: pagination
func (api *API) GetNotificationsPage(cursor string) (TweetResponse, error) {
	url, err := url.Parse("https://api.twitter.com/2/notifications/all.json")
	if err != nil {
		panic(err)
	}

	query := url.Query()
	add_tweet_query_params(&query)
	url.RawQuery = query.Encode()

	var result TweetResponse
	err = api.do_http(url.String(), cursor, &result)

	return result, err
}

func (api *API) GetNotifications(how_many int) (TweetTrove, error) {
	resp, err := api.GetNotificationsPage("")
	if err != nil {
		return TweetTrove{}, err
	}
	trove, err := resp.ToTweetTroveAsNotifications(api.UserID)
	if err != nil {
		panic(err)
	}

	for len(trove.Notifications) < how_many {
		resp, err = api.GetNotificationsPage(resp.GetCursor())
		if errors.Is(err, ErrRateLimited) {
			log.Warnf("Rate limited!")
			break
		} else if err != nil {
			return TweetTrove{}, err
		}
		if resp.IsEndOfFeed() {
			log.Infof("End of feed!")
			break
		}

		new_trove, err := resp.ToTweetTroveAsNotifications(api.UserID)
		if err != nil {
			panic(err)
		}
		trove.MergeWith(new_trove)
	}
	return trove, nil
}

func (t *TweetResponse) ToTweetTroveAsNotifications(current_user_id UserID) (TweetTrove, error) {
	ret, err := t.ToTweetTrove()
	if err != nil {
		return TweetTrove{}, err
	}

	// Find the "addEntries" instruction
	for _, instr := range t.Timeline.Instructions {
		sort.Sort(instr.AddEntries.Entries)
		for _, entry := range instr.AddEntries.Entries {
			id_re := regexp.MustCompile(`notification-([\w-]+)`)
			matches := id_re.FindStringSubmatch(entry.EntryID)
			if matches == nil || len(matches) == 1 {
				// Not a notification entry
				continue
			}
			notification_id := matches[1]
			notification, is_ok := ret.Notifications[NotificationID(notification_id)]
			if !is_ok {
				// Tweet entry (e.g., someone replied to you)
				notification = Notification{ID: NotificationID(notification_id)}
			}
			notification.UserID = current_user_id
			notification.SortIndex = entry.SortIndex
			if strings.Contains(entry.Content.Item.ClientEventInfo.Element, "replied") {
				notification.Type = NOTIFICATION_TYPE_REPLY
			} else if strings.Contains(entry.Content.Item.ClientEventInfo.Element, "recommended") {
				notification.Type = NOTIFICATION_TYPE_RECOMMENDED_POST
			} else if strings.Contains(entry.Content.Item.ClientEventInfo.Element, "quoted") {
				notification.Type = NOTIFICATION_TYPE_QUOTE_TWEET
			} else if strings.Contains(entry.Content.Item.ClientEventInfo.Element, "mentioned") {
				notification.Type = NOTIFICATION_TYPE_MENTION
			}
			if entry.Content.Item.Content.Tweet.ID != 0 {
				notification.ActionTweetID = TweetID(entry.Content.Item.Content.Tweet.ID)
				notification.ActionUserID = UserID(ret.Tweets[notification.ActionTweetID].UserID)
			}

			if entry.Content.Item.Content.Notification.ID != "" {
				notification.UserIDs = []UserID{}
				for _, u_id := range entry.Content.Item.Content.Notification.FromUsers {
					notification.UserIDs = append(notification.UserIDs, UserID(u_id))
					notification.ActionUserID = UserID(u_id)
				}

				notification.TweetIDs = []TweetID{}
				for _, t_id := range entry.Content.Item.Content.Notification.TargetTweets {
					notification.TweetIDs = append(notification.TweetIDs, TweetID(t_id))
					notification.ActionTweetID = TweetID(t_id)
				}
			}
			ret.Notifications[notification.ID] = notification
		}
	}
	return ret, err
}

func ParseSingleNotification(n APINotification) Notification {
	ret := Notification{}
	ret.ID = NotificationID(n.ID)

	for i := len(n.Message.Entities) - 1; i >= 0; i -= 1 {
		from := n.Message.Entities[i].FromIndex
		to := n.Message.Entities[i].ToIndex

		runetext := []rune(n.Message.Text)

		n.Message.Text = string(runetext[0:from]) + string(runetext[to:])
	}
	// t.Entities.ReplyMentions = strings.TrimSpace(string([]rune(t.FullText)[0:t.DisplayTextRange[0]]))

	if strings.HasSuffix(n.Message.Text, "followed you") {
		ret.Type = NOTIFICATION_TYPE_FOLLOW
	} else if strings.Contains(n.Message.Text, "liked") {
		ret.Type = NOTIFICATION_TYPE_LIKE
	} else if strings.Contains(n.Message.Text, "reposted") {
		ret.Type = NOTIFICATION_TYPE_RETWEET
	} else if strings.Contains(n.Message.Text, "There was a login to your account") {
		ret.Type = NOTIFICATION_TYPE_LOGIN
	}
	// TODO: more types?

	ret.SentAt = TimestampFromUnixMilli(n.TimestampMs)
	ret.UserIDs = []UserID{}
	for _, u := range n.Template.AggregateUserActionsV1.FromUsers {
		ret.UserIDs = append(ret.UserIDs, UserID(u.User.ID))
	}

	target_objs := n.Template.AggregateUserActionsV1.TargetObjects
	if len(target_objs) > 0 {
		if strings.HasSuffix(n.Message.Text, "liked your repost") {
			// Retweet
			ret.ActionRetweetID = TweetID(target_objs[0].Tweet.ID)
		} else {
			// Normal tweet
			ret.ActionTweetID = TweetID(target_objs[0].Tweet.ID)
			ret.TweetIDs = []TweetID{TweetID(target_objs[0].Tweet.ID)}
		}
	}

	return ret
}
