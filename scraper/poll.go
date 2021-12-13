package scraper

import (
    "time"
    "strings"
    "strconv"
)

type Poll struct {
    TweetID TweetID
    NumChoices int

    Choice1 string
    Choice1_Votes int
    Choice2 string
    Choice2_Votes int
    Choice3 string
    Choice3_Votes int
    Choice4 string
    Choice4_Votes int

    VotingDuration int  // In seconds
    VotingEndsAt time.Time

    LastUpdatedAt time.Time
}

func ParseAPIPoll(apiCard APICard) Poll {
    voting_ends_at, err := time.Parse(time.RFC3339, apiCard.BindingValues.EndDatetimeUTC.StringValue)
    if err != nil {
        panic(err)
    }
    last_updated_at, err := time.Parse(time.RFC3339, apiCard.BindingValues.LastUpdatedAt.StringValue)
    if err != nil {
        panic(err)
    }

    ret := Poll{}
    ret.NumChoices = parse_num_choices(apiCard.Name)
    ret.VotingDuration = int_or_panic(apiCard.BindingValues.DurationMinutes.StringValue) * 60
    ret.VotingEndsAt = voting_ends_at
    ret.LastUpdatedAt = last_updated_at

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
