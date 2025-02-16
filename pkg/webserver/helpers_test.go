package webserver

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetEntitiesNone(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	s := "This is just a simple string"
	entities := get_entities(s)

	require.Len(entities, 1)
	assert.Equal(ENTITY_TYPE_TEXT, entities[0].EntityType)
	assert.Equal(s, entities[0].Contents)
}

func TestGetEntitiesHashtagAndMention(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	s := "A string with a #hashtag and a @mention in it"
	entities := get_entities(s)

	require.Len(entities, 5)
	assert.Equal(entities[0].EntityType, ENTITY_TYPE_TEXT)
	assert.Equal(entities[0].Contents, "A string with a ")
	assert.Equal(entities[1].EntityType, ENTITY_TYPE_HASHTAG)
	assert.Equal(entities[1].Contents, "hashtag")
	assert.Equal(entities[2].EntityType, ENTITY_TYPE_TEXT)
	assert.Equal(entities[2].Contents, " and a ")
	assert.Equal(entities[3].EntityType, ENTITY_TYPE_MENTION)
	assert.Equal(entities[3].Contents, "mention")
	assert.Equal(entities[4].EntityType, ENTITY_TYPE_TEXT)
	assert.Equal(entities[4].Contents, " in it")
}

func TestGetEntitiesNoMatchEmail(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	s := "My email is somebody@somedomain.com"
	entities := get_entities(s)

	require.Len(entities, 1)
	assert.Equal(entities[0].EntityType, ENTITY_TYPE_TEXT)
	assert.Equal(entities[0].Contents, s)
}

func TestEntitiesWithParentheses(t *testing.T) {
	assert := assert.New(t)

	entities := get_entities("Companies are looking for ways to reduce costs (@BowTiedBull has said this), through process automation.)")
	assert.Len(entities, 3)
	assert.Equal(entities[0].EntityType, ENTITY_TYPE_TEXT)
	assert.Equal(entities[0].Contents, "Companies are looking for ways to reduce costs (")
	assert.Equal(entities[1].EntityType, ENTITY_TYPE_MENTION)
	assert.Equal(entities[1].Contents, "BowTiedBull")
	assert.Equal(entities[2].EntityType, ENTITY_TYPE_TEXT)
	assert.Equal(entities[2].Contents, " has said this), through process automation.)")
}
