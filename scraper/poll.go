package scraper

import (
	"net/url"
	"strconv"
	"strings"
)

type PollID int64

type Poll struct {
	ID         PollID  `db:"id"`
	TweetID    TweetID `db:"tweet_id"`
	NumChoices int     `db:"num_choices"`

	Choice1       string `db:"choice1"`
	Choice1_Votes int    `db:"choice1_votes"`
	Choice2       string `db:"choice2"`
	Choice2_Votes int    `db:"choice2_votes"`
	Choice3       string `db:"choice3"`
	Choice3_Votes int    `db:"choice3_votes"`
	Choice4       string `db:"choice4"`
	Choice4_Votes int    `db:"choice4_votes"`

	VotingDuration int       `db:"voting_duration"` // In seconds
	VotingEndsAt   Timestamp `db:"voting_ends_at"`

	LastUpdatedAt Timestamp `db:"last_scraped_at"`
}

func ParseAPIPoll(apiCard APICard) Poll {
	card_url, err := url.Parse(apiCard.ShortenedUrl)
	if err != nil {
		panic(err)
	}
	id := int_or_panic(card_url.Hostname())

	ret := Poll{}
	ret.ID = PollID(id)
	ret.NumChoices = parse_num_choices(apiCard.Name)
	ret.VotingDuration = int_or_panic(apiCard.BindingValues.DurationMinutes.StringValue) * 60
	ret.VotingEndsAt, err = TimestampFromString(apiCard.BindingValues.EndDatetimeUTC.StringValue)
	if err != nil {
		panic(err)
	}
	ret.LastUpdatedAt, err = TimestampFromString(apiCard.BindingValues.LastUpdatedAt.StringValue)
	if err != nil {
		panic(err)
	}

	ret.Choice1 = apiCard.BindingValues.Choice1.StringValue
	ret.Choice1_Votes = int_or_panic(apiCard.BindingValues.Choice1_Count.StringValue)
	ret.Choice2 = apiCard.BindingValues.Choice2.StringValue
	ret.Choice2_Votes = int_or_panic(apiCard.BindingValues.Choice2_Count.StringValue)

	if ret.NumChoices > 2 {
		ret.Choice3 = apiCard.BindingValues.Choice3.StringValue
		ret.Choice3_Votes = int_or_panic(apiCard.BindingValues.Choice3_Count.StringValue)
	}
	if ret.NumChoices > 3 {
		ret.Choice4 = apiCard.BindingValues.Choice4.StringValue
		ret.Choice4_Votes = int_or_panic(apiCard.BindingValues.Choice4_Count.StringValue)
	}

	return ret
}

func parse_num_choices(card_name string) int {
	if strings.Index(card_name, "poll") != 0 || strings.Index(card_name, "choice") != 5 {
		panic("Not valid card name: " + card_name)
	}

	return int_or_panic(card_name[4:5])
}

func int_or_panic(s string) int {
	result, err := strconv.Atoi(s)
	if err != nil {
		panic(err)
	}
	return result
}
