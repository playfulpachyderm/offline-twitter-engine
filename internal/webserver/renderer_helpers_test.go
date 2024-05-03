package webserver

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

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
