package scraper_test

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	. "gitlab.com/offline-twitter/twitter_offline_engine/pkg/scraper"
)

func TestParsePoll2Choices(t *testing.T) {
	assert := assert.New(t)
	data, err := os.ReadFile("test_responses/tweet_content/poll_card_2_options.json")
	if err != nil {
		panic(err)
	}
	var apiCard APICard
	err = json.Unmarshal(data, &apiCard)
	require.NoError(t, err)

	poll := ParseAPIPoll(apiCard)
	assert.Equal(PollID(1457419248461131776), poll.ID)
	assert.Equal(2, poll.NumChoices)
	assert.Equal(60*60*24, poll.VotingDuration)
	assert.Equal(int64(1636397201), poll.VotingEndsAt.Unix())
	assert.Equal(int64(1636318755), poll.LastUpdatedAt.Unix())

	assert.Less(poll.LastUpdatedAt.Unix(), poll.VotingEndsAt.Unix())
	assert.Equal("Yes", poll.Choice1)
	assert.Equal("No", poll.Choice2)
	assert.Equal(529, poll.Choice1_Votes)
	assert.Equal(2182, poll.Choice2_Votes)
}

func TestParsePoll4Choices(t *testing.T) {
	assert := assert.New(t)
	data, err := os.ReadFile("test_responses/tweet_content/poll_card_4_options_ended.json")
	if err != nil {
		panic(err)
	}
	var apiCard APICard
	err = json.Unmarshal(data, &apiCard)
	require.NoError(t, err)

	poll := ParseAPIPoll(apiCard)
	assert.Equal(PollID(1455611588854140929), poll.ID)
	assert.Equal(4, poll.NumChoices)
	assert.Equal(60*60*24, poll.VotingDuration)
	assert.Equal(int64(1635966221), poll.VotingEndsAt.Unix())
	assert.Equal(int64(1635966226), poll.LastUpdatedAt.Unix())
	assert.Greater(poll.LastUpdatedAt.Unix(), poll.VotingEndsAt.Unix())

	assert.Equal("Alec Baldwin", poll.Choice1)
	assert.Equal(1669, poll.Choice1_Votes)

	assert.Equal("Andew Cuomo", poll.Choice2)
	assert.Equal(272, poll.Choice2_Votes)

	assert.Equal("George Floyd", poll.Choice3)
	assert.Equal(829, poll.Choice3_Votes)

	assert.Equal("Derek Chauvin", poll.Choice4)
	assert.Equal(2397, poll.Choice4_Votes)
}
