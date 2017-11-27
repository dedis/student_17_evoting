package master

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/qantik/nevv/election"
)

func TestIsAdmin(t *testing.T) {
	master := &Master{nil, []election.User{123456}}
	assert.True(t, master.IsAdmin(123456))
	assert.False(t, master.IsAdmin(654321))
}
