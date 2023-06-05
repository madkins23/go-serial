package pointer

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFinder(t *testing.T) {
	ClearFinderCache()
	ClearTargetCache()
	assert.Nil(t, GetFinder(testGroup))
	assert.False(t, HasFinder(testGroup))
	assert.NoError(t,
		SetFinder(testGroup, func(_ string) (Target, error) { return nil, nil },
			false))
	assert.NotNil(t, GetFinder(testGroup))
	assert.True(t, HasFinder(testGroup))
	assert.ErrorIs(t,
		SetFinder(testGroup, func(_ string) (Target, error) { return nil, nil }, false),
		ErrFinderAlreadyExists)
	assert.NoError(t,
		SetFinder(testGroup, func(_ string) (Target, error) { return nil, nil }, true))
	assert.ErrorIs(t,
		SetFinder("", func(_ string) (Target, error) { return nil, nil }, false),
		ErrNoFinderGroup)
	assert.ErrorIs(t, SetFinder(testGroup, nil, false), ErrFinderIsNil)
}
