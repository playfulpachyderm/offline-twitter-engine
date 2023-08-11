package persistence_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/offline-twitter/twitter_offline_engine/pkg/persistence"
	. "gitlab.com/offline-twitter/twitter_offline_engine/pkg/scraper"
)

// Use a cursor, sort by newest
func TestCursorSearchByNewest(t *testing.T) {
	require := require.New(t)
	assert := assert.New(t)

	profile, err := persistence.LoadProfile("../../sample_data/profile")
	require.NoError(err)

	c := persistence.NewCursor()
	c.PageSize = 3
	c.Keywords = []string{"think"}
	c.SortOrder = persistence.SORT_ORDER_NEWEST

	feed, err := profile.NextPage(c)
	require.NoError(err)

	assert.Len(feed.Items, 3)
	assert.Len(feed.Retweets, 0)
	assert.Equal(feed.Items[0].TweetID, TweetID(1439067163508150272))
	assert.Equal(feed.Items[1].TweetID, TweetID(1439027915404939265))
	assert.Equal(feed.Items[2].TweetID, TweetID(1428939163961790466))

	next_cursor := feed.CursorBottom
	assert.Equal(next_cursor.CursorPosition, persistence.CURSOR_MIDDLE)
	assert.Equal(next_cursor.SortOrder, c.SortOrder)
	assert.Equal(next_cursor.Keywords, c.Keywords)
	assert.Equal(next_cursor.PageSize, c.PageSize)
	assert.Equal(next_cursor.CursorValue, 1629520619)

	feed, err = profile.NextPage(next_cursor)
	require.NoError(err)

	assert.Len(feed.Items, 2)
	assert.Len(feed.Retweets, 0)
	assert.Equal(feed.Items[0].TweetID, TweetID(1413772782358433792))
	assert.Equal(feed.Items[1].TweetID, TweetID(1343633011364016128))

	next_cursor = feed.CursorBottom
	assert.Equal(next_cursor.CursorPosition, persistence.CURSOR_END)
}

// Search retweets, sorted by oldest
func TestCursorSearchWithRetweets(t *testing.T) {
	require := require.New(t)
	assert := assert.New(t)

	profile, err := persistence.LoadProfile("../../sample_data/profile")
	require.NoError(err)

	c := persistence.NewCursor()
	c.PageSize = 3
	c.RetweetedByUserHandle = "cernovich"
	c.SortOrder = persistence.SORT_ORDER_OLDEST

	feed, err := profile.NextPage(c)
	require.NoError(err)

	assert.Len(feed.Items, 3)
	assert.Len(feed.Retweets, 3)
	assert.Equal(feed.Items[0].RetweetID, TweetID(1490100255987171332))
	assert.Equal(feed.Items[1].RetweetID, TweetID(1490119308692766723))
	assert.Equal(feed.Items[2].RetweetID, TweetID(1490135787144237058))

	next_cursor := feed.CursorBottom
	assert.Equal(next_cursor.CursorPosition, persistence.CURSOR_MIDDLE)
	assert.Equal(next_cursor.SortOrder, c.SortOrder)
	assert.Equal(next_cursor.Keywords, c.Keywords)
	assert.Equal(next_cursor.PageSize, c.PageSize)
	assert.Equal(next_cursor.CursorValue, 1644111031)

	feed, err = profile.NextPage(next_cursor)
	require.NoError(err)

	assert.Len(feed.Items, 0)
	next_cursor = feed.CursorBottom
	assert.Equal(next_cursor.CursorPosition, persistence.CURSOR_END)
}
