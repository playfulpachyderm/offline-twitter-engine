package scraper_test

import (
    "testing"
    "io/ioutil"
    "encoding/json"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"

    . "offline_twitter/scraper"
)

func TestParseAPIMedia(t *testing.T) {
    assert := assert.New(t)
    data, err := ioutil.ReadFile("test_responses/tweet_content/image.json")
    if err != nil {
        panic(err)
    }
    var apimedia APIMedia
    err = json.Unmarshal(data, &apimedia)
    require.NoError(t, err)

    image := ParseAPIMedia(apimedia)
    assert.Equal(ImageID(1395882862289772553), image.ID)
    assert.Equal("https://pbs.twimg.com/media/E18sEUrWYAk8dBl.jpg", image.RemoteURL)
    assert.Equal(593, image.Width)
    assert.Equal(239, image.Height)
    assert.Equal("E18sEUrWYAk8dBl.jpg", image.LocalFilename)
    assert.False(image.IsDownloaded)
}
