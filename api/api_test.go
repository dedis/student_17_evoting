package api

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMessages(t *testing.T) {
	assert.NotNil(t, Ping{})
	assert.NotNil(t, Link{})
	assert.NotNil(t, LinkReply{})
	assert.NotNil(t, Open{})
	assert.NotNil(t, OpenReply{})
	assert.NotNil(t, Cast{})
	assert.NotNil(t, CastReply{})
	assert.NotNil(t, Finalize{})
	assert.NotNil(t, FinalizeReply{})
	assert.NotNil(t, Aggregate{})
	assert.NotNil(t, AggregateReply{})
}
