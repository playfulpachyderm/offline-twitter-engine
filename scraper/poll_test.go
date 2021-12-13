package scraper_test

import (
    "testing"
    "io/ioutil"
    "encoding/json"

    "offline_twitter/scraper"
)

func TestParsePoll2Choices(t *testing.T) {
    data, err := ioutil.ReadFile("test_responses/tweet_content/poll_card_2_options.json")
    if err != nil {
        panic(err)
    }
    var apiCard scraper.APICard
    err = json.Unmarshal(data, &apiCard)
    if err != nil {
        t.Fatal(err.Error())
    }

    poll := scraper.ParseAPIPoll(apiCard)
    if poll.NumChoices != 2 {
        t.Errorf("Expected %d choices, got %d", 2, poll.NumChoices)
    }
    if poll.VotingDuration != 60 * 60 * 24 {
        t.Errorf("Expected duratino %d, got %d", 60 * 60 * 24, poll.VotingDuration)
    }
    expected_ending := int64(1636397201)
    if poll.VotingEndsAt.Unix() != expected_ending {
        t.Errorf("Expected closing time %d, got %d", expected_ending, poll.VotingEndsAt.Unix())
    }
    expected_last_updated := int64(1636318755)
    if poll.LastUpdatedAt.Unix() != expected_last_updated {
        t.Errorf("Expected last-updated time %d, got %d", expected_last_updated, poll.LastUpdatedAt.Unix())
    }
    if expected_last_updated > expected_ending {
        t.Errorf("Last updated should be before poll closes!")
    }

    if poll.Choice1 != "Yes" || poll.Choice2 != "No" {
        t.Errorf("Expected %q and %q, got %q and %q", "Yes", "No", poll.Choice1, poll.Choice2)
    }
    if poll.Choice1_Votes != 529 {
        t.Errorf("Expected %d votes for choice 1, got %d", 529, poll.Choice1_Votes)
    }
    if poll.Choice2_Votes != 2182 {
        t.Errorf("Expected %d votes for choice 2, got %d", 2182, poll.Choice2_Votes)
    }
}

func TestParsePoll4Choices(t *testing.T) {
    data, err := ioutil.ReadFile("test_responses/tweet_content/poll_card_4_options_ended.json")
    if err != nil {
        panic(err)
    }
    var apiCard scraper.APICard
    err = json.Unmarshal(data, &apiCard)
    if err != nil {
        t.Fatal(err.Error())
    }

    poll := scraper.ParseAPIPoll(apiCard)
    if poll.NumChoices != 4 {
        t.Errorf("Expected %d choices, got %d", 4, poll.NumChoices)
    }
    if poll.VotingDuration != 60 * 60 * 24 {
        t.Errorf("Expected duratino %d, got %d", 60 * 60 * 24, poll.VotingDuration)
    }
    expected_ending := int64(1635966221)
    if poll.VotingEndsAt.Unix() != expected_ending {
        t.Errorf("Expected closing time %d, got %d", expected_ending, poll.VotingEndsAt.Unix())
    }
    expected_last_updated := int64(1635966226)
    if poll.LastUpdatedAt.Unix() != expected_last_updated {
        t.Errorf("Expected last-updated time %d, got %d", expected_last_updated, poll.LastUpdatedAt.Unix())
    }
    if expected_last_updated < expected_ending {
        t.Errorf("Last updated should be after poll closes!")
    }

    if poll.Choice1 != "Alec Baldwin" || poll.Choice1_Votes != 1669 {
        t.Errorf("Expected %q with %d, got %q with %d", "Alec Baldwin", 1669, poll.Choice1, poll.Choice1_Votes)
    }
    if poll.Choice2 != "Andew Cuomo" || poll.Choice2_Votes != 272 {
        t.Errorf("Expected %q with %d, got %q with %d", "Andew Cuomo", 272, poll.Choice2, poll.Choice2_Votes)
    }
    if poll.Choice3 != "George Floyd" || poll.Choice3_Votes != 829 {
        t.Errorf("Expected %q with %d, got %q with %d", "George Floyd", 829, poll.Choice3, poll.Choice3_Votes)
    }
    if poll.Choice4 != "Derek Chauvin" || poll.Choice4_Votes != 2397 {
        t.Errorf("Expected %q with %d, got %q with %d", "Derek Chauvin", 2397, poll.Choice4, poll.Choice4_Votes)
    }
}
