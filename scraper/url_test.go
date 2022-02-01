package scraper_test

import (
    "testing"
    "io/ioutil"
    "encoding/json"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"

    . "offline_twitter/scraper"
)

func TestParseAPIUrlCard(t *testing.T) {
    assert := assert.New(t)
    data, err := ioutil.ReadFile("test_responses/tweet_content/url_card.json")
    if err != nil {
        panic(err)
    }
    var apiCard APICard
    err = json.Unmarshal(data, &apiCard)
    require.NoError(t, err)

    url := ParseAPIUrlCard(apiCard)
    assert.Equal("reason.com", url.Domain)
    assert.Equal("L.A. Teachers Union Leader: 'There's No Such Thing As Learning Loss'", url.Title)
    assert.Equal("\"It’s OK that our babies may not have learned all their times tables,\" says Cecily Myart-Cruz. \"They learned resilience.\"", url.Description)
    assert.Equal(600, url.ThumbnailWidth)
    assert.Equal(315, url.ThumbnailHeight)
    assert.Equal("https://pbs.twimg.com/card_img/1434998862305968129/odDi9EqO?format=jpg&name=600x600", url.ThumbnailRemoteUrl)
    assert.Equal("odDi9EqO_600x600.jpg", url.ThumbnailLocalPath)
    assert.Equal(UserID(155581583), url.CreatorID)
    assert.Equal(UserID(16467567), url.SiteID)
    assert.True(url.HasThumbnail)
    assert.False(url.IsContentDownloaded)
}

func TestParseAPIUrlCardWithPlayer(t *testing.T) {
    assert := assert.New(t)
    data, err := ioutil.ReadFile("test_responses/tweet_content/url_card_with_player.json")
    if err != nil {
        panic(err)
    }
    var apiCard APICard
    err = json.Unmarshal(data, &apiCard)
    require.NoError(t, err)

    url := ParseAPIUrlCard(apiCard)
    assert.Equal("www.youtube.com", url.Domain)
    assert.Equal("The Politically Incorrect Guide to the Constitution (Starring Tom...", url.Title)
    assert.Equal("Watch this episode on LBRY/Odysee: https://odysee.com/@capitalresearch:5/the-politically-incorrect-guide-to-the:8Watch this episode on Rumble: https://rumble...", url.Description)
    assert.Equal("https://pbs.twimg.com/card_img/1437849456423194639/_1t0btyt?format=jpg&name=800x320_1", url.ThumbnailRemoteUrl)
    assert.Equal("_1t0btyt_800x320_1.jpg", url.ThumbnailLocalPath)
    assert.Equal(UserID(10228272), url.SiteID)
    assert.True(url.HasThumbnail)
    assert.False(url.IsContentDownloaded)
}

func TestParseAPIUrlCardWithPlayerAndPlaceholderThumbnail(t *testing.T) {
    assert := assert.New(t)
    data, err := ioutil.ReadFile("test_responses/tweet_content/url_card_with_player_placeholder_image.json")
    if err != nil {
        panic(err)
    }
    var apiCard APICard
    err = json.Unmarshal(data, &apiCard)
    require.NoError(t, err)

    url := ParseAPIUrlCard(apiCard)
    assert.Equal("www.youtube.com", url.Domain)
    assert.Equal("Did Michael Malice Turn Me into an Anarchist? | Ep 181", url.Title)
    assert.Equal("SUBSCRIBE TO THE NEW SHOW W/ ELIJAH & SYDNEY: \"YOU ARE HERE\"YT: https://www.youtube.com/youareheredaily______________________________________________________...", url.Description)
    assert.Equal("https://pbs.twimg.com/cards/player-placeholder.png", url.ThumbnailRemoteUrl)
    assert.Equal("player-placeholder.png", url.ThumbnailLocalPath)
    assert.Equal(UserID(10228272), url.SiteID)
    assert.True(url.HasThumbnail)
    assert.False(url.IsContentDownloaded)
}

func TestParseAPIUrlCardWithoutThumbnail(t *testing.T) {
    assert := assert.New(t)
    data, err := ioutil.ReadFile("test_responses/tweet_content/url_card_without_thumbnail.json")
    if err != nil {
        panic(err)
    }
    var apiCard APICard
    err = json.Unmarshal(data, &apiCard)
    require.NoError(t, err)

    url := ParseAPIUrlCard(apiCard)
    assert.Equal("en.m.wikipedia.org", url.Domain)
    assert.Equal("Entryism - Wikipedia", url.Title)
    assert.Equal("", url.Description)
    assert.True(url.HasCard)
    assert.False(url.HasThumbnail)
}
