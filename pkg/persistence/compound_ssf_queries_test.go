package persistence_test

import (
	"testing"

	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	. "gitlab.com/offline-twitter/twitter_offline_engine/pkg/persistence"
	. "gitlab.com/offline-twitter/twitter_offline_engine/pkg/scraper"
)

// Use a cursor, sort by newest
func TestCursorSearchByNewest(t *testing.T) {
	require := require.New(t)
	assert := assert.New(t)

	profile, err := LoadProfile("../../sample_data/profile")
	require.NoError(err)

	c := NewCursor()
	c.PageSize = 3
	c.Keywords = []string{"think"}
	c.SortOrder = SORT_ORDER_NEWEST

	feed, err := profile.NextPage(c, UserID(0))
	require.NoError(err)

	assert.Len(feed.Items, 3)
	assert.Len(feed.Retweets, 0)
	assert.Equal(feed.Items[0].TweetID, TweetID(1439067163508150272))
	assert.Equal(feed.Items[1].TweetID, TweetID(1439027915404939265))
	assert.Equal(feed.Items[2].TweetID, TweetID(1428939163961790466))

	next_cursor := feed.CursorBottom
	assert.Equal(next_cursor.CursorPosition, CURSOR_MIDDLE)
	assert.Equal(next_cursor.SortOrder, c.SortOrder)
	assert.Equal(next_cursor.Keywords, c.Keywords)
	assert.Equal(next_cursor.PageSize, c.PageSize)
	assert.Equal(next_cursor.CursorValue, 1629520619000)

	feed, err = profile.NextPage(next_cursor, UserID(0))
	require.NoError(err)

	assert.Len(feed.Items, 2)
	assert.Len(feed.Retweets, 0)
	assert.Equal(feed.Items[0].TweetID, TweetID(1413772782358433792))
	assert.Equal(feed.Items[1].TweetID, TweetID(1343633011364016128))

	next_cursor = feed.CursorBottom
	assert.Equal(next_cursor.CursorPosition, CURSOR_END)
}

// Search retweets, sorted by oldest
func TestCursorSearchWithRetweets(t *testing.T) {
	require := require.New(t)
	assert := assert.New(t)

	profile, err := LoadProfile("../../sample_data/profile")
	require.NoError(err)

	c := NewCursor()
	c.PageSize = 3
	c.RetweetedByUserHandle = "cernovich"
	c.FilterRetweets = REQUIRE
	c.SortOrder = SORT_ORDER_OLDEST

	feed, err := profile.NextPage(c, UserID(0))
	require.NoError(err)

	assert.Len(feed.Items, 3)
	assert.Len(feed.Retweets, 3)
	assert.Equal(feed.Items[0].RetweetID, TweetID(1490100255987171332))
	assert.Equal(feed.Items[1].RetweetID, TweetID(1490119308692766723))
	assert.Equal(feed.Items[2].RetweetID, TweetID(1490135787144237058))

	next_cursor := feed.CursorBottom
	assert.Equal(next_cursor.CursorPosition, CURSOR_MIDDLE)
	assert.Equal(next_cursor.SortOrder, c.SortOrder)
	assert.Equal(next_cursor.Keywords, c.Keywords)
	assert.Equal(next_cursor.PageSize, c.PageSize)
	assert.Equal(next_cursor.CursorValue, 1644111031000)

	feed, err = profile.NextPage(next_cursor, UserID(0))
	require.NoError(err)

	assert.Len(feed.Items, 0)
	next_cursor = feed.CursorBottom
	assert.Equal(next_cursor.CursorPosition, CURSOR_END)
}

// Offline Following Timeline
func TestTimeline(t *testing.T) {
	require := require.New(t)
	assert := assert.New(t)

	profile, err := LoadProfile("../../sample_data/profile")
	require.NoError(err)

	c := NewTimelineCursor()
	c.PageSize = 6

	feed, err := profile.NextPage(c, UserID(0))
	require.NoError(err)

	assert.Len(feed.Items, 6)
	assert.Len(feed.Retweets, 4)
	assert.Equal(feed.Items[0].TweetID, TweetID(1826778617705115868))
	assert.Equal(feed.Items[1].RetweetID, TweetID(1490135787144237058))
	assert.Equal(feed.Items[2].RetweetID, TweetID(1490135787124232223))
	assert.Equal(feed.Items[3].RetweetID, TweetID(1490119308692766723))
	assert.Equal(feed.Items[4].RetweetID, TweetID(1490100255987171332))
	assert.Equal(feed.Items[5].TweetID, TweetID(1453461248142495744))

	next_cursor := feed.CursorBottom
	assert.Equal(next_cursor.CursorPosition, CURSOR_MIDDLE)
	assert.Equal(next_cursor.SortOrder, c.SortOrder)
	assert.Equal(next_cursor.Keywords, c.Keywords)
	assert.Equal(next_cursor.PageSize, c.PageSize)
	assert.Equal(next_cursor.CursorValue, 1635367140000)

	next_cursor.CursorValue = 1631935323000 // Scroll down a bit, kind of randomly
	next_cursor.PageSize = 5
	feed, err = profile.NextPage(next_cursor, UserID(0))
	require.NoError(err)

	assert.Len(feed.Items, 5)
	assert.Len(feed.Retweets, 1)
	assert.Equal(feed.Items[0].TweetID, TweetID(1439027915404939265))
	assert.Equal(feed.Items[1].TweetID, TweetID(1413773185296650241))
	assert.Equal(feed.Items[2].TweetID, TweetID(1413664406995566593))
	assert.Equal(feed.Items[3].RetweetID, TweetID(144919526660333333))
	assert.Equal(feed.Items[4].TweetID, TweetID(1413658466795737091))
}

func TestKeywordSearch(t *testing.T) {
	require := require.New(t)
	assert := assert.New(t)

	profile, err := LoadProfile("../../sample_data/profile")
	require.NoError(err)
	c := NewCursor()

	// Multiple words without quotes
	c.Keywords = []string{"who", "are"}
	feed, err := profile.NextPage(c, UserID(0))
	require.NoError(err)
	assert.True(len(feed.Items) > 1)

	// Add quotes
	c.Keywords = []string{"who are"}
	feed, err = profile.NextPage(c, UserID(0))
	require.NoError(err)
	assert.Len(feed.Items, 1)
	assert.Equal(feed.Items[0].TweetID, TweetID(1261483383483293700))

	// With gibberish (no matches)
	c.Keywords = []string{"fasdfjkafsldfjsff"}
	feed, err = profile.NextPage(c, UserID(0))
	require.NoError(err)
	assert.Len(feed.Items, 0)
}

func TestSearchReplyingToUser(t *testing.T) {
	require := require.New(t)
	assert := assert.New(t)

	profile, err := LoadProfile("../../sample_data/profile")
	require.NoError(err)
	c := NewCursor()

	// Replying to a user
	c.ToUserHandles = []UserHandle{"spacex"}
	feed, err := profile.NextPage(c, UserID(0))
	require.NoError(err)
	assert.Len(feed.Items, 2)
	assert.Equal(feed.Items[0].TweetID, TweetID(1428951883058753537))
	assert.Equal(feed.Items[1].TweetID, TweetID(1428939163961790466))

	// Replying to two users
	c.ToUserHandles = []UserHandle{"spacex", "covfefeanon"}
	feed, err = profile.NextPage(c, UserID(0))
	require.NoError(err)
	assert.Len(feed.Items, 1)
	assert.Equal(feed.Items[0].TweetID, TweetID(1428939163961790466))
}

func TestSearchDateFilters(t *testing.T) {
	require := require.New(t)
	assert := assert.New(t)

	profile, err := LoadProfile("../../sample_data/profile")
	require.NoError(err)
	c := NewCursor()
	c.SortOrder = SORT_ORDER_MOST_LIKES

	// Since timestamp
	c.SinceTimestamp.Time = time.Date(2021, 10, 1, 0, 0, 0, 0, time.UTC)
	c.FromUserHandle = UserHandle("cernovich")
	feed, err := profile.NextPage(c, UserID(0))
	require.NoError(err)
	assert.Len(feed.Items, 1)
	assert.Equal(feed.Items[0].TweetID, TweetID(1453461248142495744))

	// Until timestamp
	c.SinceTimestamp = TimestampFromUnix(0)
	c.UntilTimestamp.Time = time.Date(2021, 10, 1, 0, 0, 0, 0, time.UTC)
	feed, err = profile.NextPage(c, UserID(0))
	require.NoError(err)
	assert.Len(feed.Items, 4)
	assert.Equal(feed.Items[0].TweetID, TweetID(1439747634277740546))
	assert.Equal(feed.Items[1].TweetID, TweetID(1439027915404939265))
	assert.Equal(feed.Items[2].TweetID, TweetID(1439068749336748043))
	assert.Equal(feed.Items[3].TweetID, TweetID(1439067163508150272))
}

func TestSearchMediaFilters(t *testing.T) {
	require := require.New(t)
	assert := assert.New(t)

	profile, err := LoadProfile("../../sample_data/profile")
	require.NoError(err)

	// Links
	c := NewCursor()
	c.SortOrder = SORT_ORDER_MOST_LIKES
	c.FilterLinks = REQUIRE
	feed, err := profile.NextPage(c, UserID(0))
	require.NoError(err)
	assert.Len(feed.Items, 2)
	assert.Equal(feed.Items[0].TweetID, TweetID(1438642143170646017))
	assert.Equal(feed.Items[1].TweetID, TweetID(1413665734866186243))

	// Images
	c = NewCursor()
	c.SortOrder = SORT_ORDER_MOST_LIKES
	c.FilterImages = REQUIRE
	feed, err = profile.NextPage(c, UserID(0))
	require.NoError(err)
	assert.Len(feed.Items, 2)
	assert.Equal(feed.Items[0].TweetID, TweetID(1261483383483293700))
	assert.Equal(feed.Items[1].TweetID, TweetID(1426669666928414720))

	// Videos
	c = NewCursor()
	c.SortOrder = SORT_ORDER_MOST_LIKES
	c.FilterVideos = REQUIRE
	feed, err = profile.NextPage(c, UserID(0))
	require.NoError(err)
	assert.Len(feed.Items, 2)
	assert.Equal(feed.Items[0].TweetID, TweetID(1426619468327882761))
	assert.Equal(feed.Items[1].TweetID, TweetID(1453461248142495744))

	// Media (generic)
	c = NewCursor()
	c.SortOrder = SORT_ORDER_MOST_LIKES
	c.FilterMedia = REQUIRE
	feed, err = profile.NextPage(c, UserID(0))
	require.NoError(err)
	assert.Len(feed.Items, 4)
	assert.Equal(feed.Items[0].TweetID, TweetID(1426619468327882761))
	assert.Equal(feed.Items[1].TweetID, TweetID(1261483383483293700))
	assert.Equal(feed.Items[2].TweetID, TweetID(1426669666928414720))
	assert.Equal(feed.Items[3].TweetID, TweetID(1453461248142495744))

	// Polls
	c = NewCursor()
	c.FilterPolls = REQUIRE
	feed, err = profile.NextPage(c, UserID(0))
	require.NoError(err)
	assert.Len(feed.Items, 1)
	assert.Equal(feed.Items[0].TweetID, TweetID(1465534109573390348))

	// Spaces
	c = NewCursor()
	c.FilterSpaces = REQUIRE
	feed, err = profile.NextPage(c, UserID(0))
	require.NoError(err)
	assert.Len(feed.Items, 1)
	assert.Equal(feed.Items[0].TweetID, TweetID(1624833173514293249))

	// Negative filter (images)
	c = NewCursor()
	c.FilterImages = EXCLUDE
	c.FromUserHandle = UserHandle("covfefeanon")
	feed, err = profile.NextPage(c, UserID(0))
	require.NoError(err)
	assert.Len(feed.Items, 1)
	assert.Equal(feed.Items[0].TweetID, TweetID(1428951883058753537))
}
