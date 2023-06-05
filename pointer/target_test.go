package pointer

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/suite"
)

const (
	badGroup    = "badGroup"
	badKey      = "badKey"
	oldValue    = 17
	newValue    = 23
	finderGroup = "finderGroup"
	testGroup   = "testGroup"
	testKey     = "testKey"
	testNone    = "testNoSuchKey"
	testFinder  = "testFinder"
)

var _ Target = &testTarget{}

type testTarget struct {
	group, key string
	value      int
}

func newTestTarget(group, key string, value int) *testTarget {
	return &testTarget{group: group, key: key, value: value}
}

func (tt *testTarget) Group() string {
	return tt.group
}

func (tt *testTarget) Key() string {
	return tt.key
}

//////////////////////////////////////////////////////////////////////////

type TargetTestSuite struct {
	suite.Suite
}

func (suite *TargetTestSuite) SetupSuite() {
	ClearFinderCache()
	ClearTargetCache()
}

func TestTargetSuite(t *testing.T) {
	suite.Run(t, new(TargetTestSuite))
}

//////////////////////////////////////////////////////////////////////////

func (suite *TargetTestSuite) TestGetTarget_NoTarget() {
	suite.Assert().False(HasTarget(badGroup, badKey))
	target, err := GetTarget(badGroup, badKey, nil)
	suite.Assert().ErrorIs(err, ErrNoSuchTarget)
	suite.Assert().Nil(target)
	targetCache[badGroup] = make(map[string]Target)
	targetCache[badGroup][badKey] = nil
	suite.Assert().False(HasTarget(badGroup, badKey))
	target, err = GetTarget(badGroup, badKey, nil)
	suite.Assert().ErrorIs(err, ErrNoSuchTarget)
	suite.Assert().Nil(target)
}

func (suite *TargetTestSuite) TestGetTarget_Finder() {
	var (
		errFindError = errors.New("some Find() error")
		tgtForFinder = newTestTarget(testGroup, testFinder, oldValue)
	)

	suite.Assert().False(HasTarget(testGroup, testFinder))
	target, err := GetTarget(testGroup, testFinder,
		func(_ string) (Target, error) { return nil, errFindError })
	suite.Assert().ErrorIs(err, errFindError)
	suite.Assert().Nil(target)
	suite.Assert().False(HasTarget(testGroup, testFinder))
	target, err = GetTarget(testGroup, testFinder,
		func(_ string) (Target, error) { return nil, nil })
	suite.Assert().ErrorIs(err, ErrFinderTargetIsNil)
	suite.Assert().Nil(target)
	suite.Assert().False(HasTarget(testGroup, testFinder))
	target, err = GetTarget(testGroup, testFinder,
		func(_ string) (Target, error) {
			return newTestTarget(badGroup, testFinder, -1), nil
		})
	suite.Assert().ErrorIs(err, ErrBadTargetGroup)
	suite.Assert().Nil(target)
	suite.Assert().False(HasTarget(badGroup, testFinder))
	target, err = GetTarget(testGroup, testFinder,
		func(_ string) (Target, error) {
			return newTestTarget(testGroup, badKey, -1), nil
		})
	suite.Assert().ErrorIs(err, ErrBadTargetKey)
	suite.Assert().Nil(target)
	suite.Assert().False(HasTarget(testGroup, badKey))
	suite.Assert().False(HasTarget(testGroup, testFinder))
	target, err = GetTarget(testGroup, testFinder,
		func(_ string) (Target, error) { return tgtForFinder, nil })
	suite.Assert().NoError(err)
	suite.Assert().Equal(tgtForFinder, target)
	suite.Assert().True(HasTarget(testGroup, testFinder))
}

func (suite *TargetTestSuite) TestGetTarget_FinderCache() {
	var tgtForFinder = newTestTarget(finderGroup, testFinder, oldValue)
	suite.Require().False(HasTarget(finderGroup, testFinder))
	target, err := GetTarget(finderGroup, testFinder, nil)
	suite.Assert().ErrorIs(err, ErrNoSuchTarget)
	suite.Assert().Nil(target)
	suite.Require().NoError(SetFinder(finderGroup,
		func(_ string) (Target, error) { return tgtForFinder, nil },
		false))
	suite.Assert().False(HasTarget(finderGroup, testFinder))
	target, err = GetTarget(finderGroup, testFinder, nil)
	suite.Assert().NoError(err)
	suite.Assert().Equal(tgtForFinder, target)
	suite.Assert().True(HasTarget(finderGroup, testFinder))
}

func (suite *TargetTestSuite) TestSetTarget_BadTargets() {
	suite.Assert().False(HasTarget(testGroup, testNone))
	suite.Assert().ErrorIs(SetTarget(nil, false), ErrTargetIsNil)
	suite.Assert().ErrorIs(SetTarget(newTestTarget("", testKey, 0), false), ErrNoTargetGroup)
	suite.Assert().ErrorIs(SetTarget(newTestTarget(testGroup, "", 0), false), ErrNoTargetKey)
	suite.Assert().False(HasTarget(testGroup, testNone))
}

func (suite *TargetTestSuite) TestGetSetTarget() {
	// Target doesn't exist yet.
	suite.Assert().False(HasTarget(testGroup, testNone))
	target, err := GetTarget(testGroup, testNone, nil)
	suite.Assert().ErrorIs(err, ErrNoSuchTarget)
	suite.Assert().Nil(target)
	// Create target and check for it.
	suite.Assert().False(HasTarget(testGroup, testKey))
	oldTarget := newTestTarget(testGroup, testKey, oldValue)
	suite.Assert().NoError(SetTarget(oldTarget, false))
	suite.Assert().True(HasTarget(testGroup, testKey))
	target, err = GetTarget(testGroup, testKey, nil)
	suite.Assert().NoError(err)
	suite.Assert().Equal(oldTarget, target)
	suite.Assert().Equal(oldValue, oldTarget.value)
	// Try to set target again with a different value.
	suite.Assert().True(HasTarget(testGroup, testKey))
	newTarget := newTestTarget(testGroup, testKey, newValue)
	suite.Assert().ErrorIs(SetTarget(newTarget, false), ErrTargetAlreadyExists)
	target, err = GetTarget(testGroup, testKey, nil)
	suite.Assert().NoError(err)
	suite.Assert().Equal(oldTarget, target)
	// Set target again with replace.
	suite.Assert().True(HasTarget(testGroup, testKey))
	suite.Assert().NoError(SetTarget(newTarget, true))
	suite.Assert().True(HasTarget(testGroup, testKey))
	target, err = GetTarget(testGroup, testKey, nil)
	suite.Assert().NoError(err)
	suite.Assert().Equal(newTarget, target)
}
