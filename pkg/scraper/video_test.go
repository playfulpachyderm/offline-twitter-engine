package scraper_test

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	. "gitlab.com/offline-twitter/twitter_offline_engine/pkg/scraper"
)

func TestParseAPIVideo(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)
	data, err := os.ReadFile("test_responses/tweet_content/video.json")
	require.NoError(err)

	var apivideo APIExtendedMedia
	err = json.Unmarshal(data, &apivideo)
	require.NoError(err)

	video := ParseAPIVideo(apivideo)
	assert.Equal(VideoID(1418951950020845568), video.ID)
	assert.Equal(1280, video.Height)
	assert.Equal(720, video.Width)
	assert.Equal("https://video.twimg.com/ext_tw_video/1418951950020845568/pu/vid/720x1280/sm4iL9_f8Lclh0aa.mp4?tag=12", video.RemoteURL)
	assert.Equal("sm/sm4iL9_f8Lclh0aa.mp4", video.LocalFilename)
	assert.Equal("https://pbs.twimg.com/ext_tw_video_thumb/1418951950020845568/pu/img/eUTaYYfuAJ8FyjUi.jpg", video.ThumbnailRemoteUrl)
	assert.Equal("eU/eUTaYYfuAJ8FyjUi.jpg", video.ThumbnailLocalPath)
	assert.Equal(275952, video.ViewCount)
	assert.Equal(88300, video.Duration)
	assert.False(video.IsDownloaded)
}

func TestParseGeoblockedVideo(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)
	data, err := os.ReadFile("test_responses/tweet_content/video_geoblocked.json")
	require.NoError(err)

	var apivideo APIExtendedMedia
	err = json.Unmarshal(data, &apivideo)
	require.NoError(err)

	video := ParseAPIVideo(apivideo)
	assert.True(video.IsGeoblocked)
}
