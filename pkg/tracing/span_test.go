package tracing_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	. "gitlab.com/offline-twitter/twitter_offline_engine/pkg/tracing"
)

func TestSpanInContext(t *testing.T) {
	assert := assert.New(t)

	base_ctx := context.Background()

	// Initially, only root span is active
	new_ctx, root_span := InitTrace(base_ctx, "root")
	assert.Equal(GetActiveSpan(new_ctx), root_span)

	// Add a child span
	assert.Len(root_span.Children, 0)
	child1 := root_span.AddChild("child1")
	assert.Len(root_span.Children, 1)
	require.True(t, child1 == root_span.Children[0])
	assert.Equal(GetActiveSpan(new_ctx), child1) // Now a nested span should be active

	// Close it
	time.Sleep(3 * time.Millisecond)
	assert.True(child1.IsActive)
	child1.End()
	assert.False(child1.IsActive)
	assert.Equal(GetActiveSpan(new_ctx), root_span) // First child is closed; back to root span
	assert.Greater(child1.Duration(), 3*time.Millisecond)

	// Add a second child
	assert.Len(root_span.Children, 1)
	child2 := root_span.AddChild("child2")
	assert.Len(root_span.Children, 2)
	require.True(t, child2 == root_span.Children[1])
	assert.Equal(GetActiveSpan(new_ctx), child2) // should be active

	// Close it
	time.Sleep(3 * time.Millisecond)
	assert.True(child2.IsActive)
	child2.End()
	assert.False(child2.IsActive)
	assert.Equal(GetActiveSpan(new_ctx), root_span) // child is closed; back to root span again
	assert.Greater(child2.Duration(), 3*time.Millisecond)

	// Save it
	assert.Equal(SpanID(0), root_span.ID)
	assert.Equal(SpanID(0), child1.ID)
	assert.Equal(SpanID(0), child2.ID)
	test_db.SaveSpan(root_span)
	assert.NotEqual(SpanID(0), root_span.ID)
	assert.NotEqual(SpanID(0), child1.ID)
	assert.NotEqual(SpanID(0), child2.ID)
}
