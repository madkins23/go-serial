package pointer

import (
	"errors"
	"fmt"
)

// Target defines the interface for items that can be referenced by Pointer objects.
type Target interface {
	// Group returns an arbitrary group name for the target.
	// This is not the same as a type, as various subtypes may be grouped together.
	// Caching items by group allows each group to have its own key space.
	Group() string

	// Key returns a unique key name for the target.
	Key() string
}

//------------------------------------------------------------------------

var (
	// ErrNoSuchTarget is returned from GetTarget when no Target is found.
	ErrNoSuchTarget = errors.New("no such target")

	// ErrNoTargetGroup is returned from SetTarget when the Target has an empty group ("").
	ErrNoTargetGroup = errors.New("no group for target")

	// ErrNoTargetKey is returned from SetTarget when the Target has an empty key ("").
	ErrNoTargetKey = errors.New("no key for target")

	// ErrTargetIsNil is returned from SetTarget if the specified Target is nil.
	ErrTargetIsNil = errors.New("target is nil")

	// ErrFinderTargetIsNil is returned from SetTarget if the Target returned by the Finder is nil.
	ErrFinderTargetIsNil = errors.New("target returned by finder is nil")

	// ErrTargetAlreadyExists is returned from SetTarget if the targetCache already
	// has a Target for the new Target's group and key and the replace flag is false.
	ErrTargetAlreadyExists = errors.New("target already exists")
)

//------------------------------------------------------------------------

// Internal targetCache for Target items.
var targetCache = make(map[string]map[string]Target)

// ClearCache removes all target finderCache entries.
// For test purposes, all other usage suspect.
//
// Deprecated: use ClearTargetCache instead.
func ClearCache() {
	ClearTargetCache()
}

// ClearTargetCache removes all target finderCache entries.
// For test purposes, all other usage suspect.
func ClearTargetCache() {
	targetCache = make(map[string]map[string]Target)
}

//------------------------------------------------------------------------

// HasTarget returns true if the specified group and key have a Target in the Target cache.
func HasTarget(group, key string) bool {
	target, found := targetCache[group][key]
	return found && target != nil
}

// GetTarget returns a Target object from the targetCache for use in Pointer implementations.
// If there is no such Target and no Finder the ErrNoSuchTarget error is returned.
// If the Finder is used to acquire the Target it is added to the targetCache and returned.
func GetTarget(group, key string, finder Finder) (Target, error) {
	if target, found := targetCache[group][key]; found && target != nil {
		return target, nil
	} else if finder == nil {
		return nil, ErrNoSuchTarget
	} else if target, err := finder(key); err != nil {
		return nil, fmt.Errorf("find item: %w", err)
	} else if target == nil {
		return nil, ErrFinderTargetIsNil
	} else if err := SetTarget(target, false); err != nil {
		return nil, fmt.Errorf("set target: %w", err)
	} else {
		return target, nil
	}
}

// SetTarget adds the specified Target to the targetCache.
// Use this function for Pointer implementations and preloading the targetCache.
func SetTarget(target Target, replace bool) error {
	if target == nil {
		return ErrTargetIsNil
	} else if group := target.Group(); group == "" {
		return ErrNoTargetGroup
	} else if key := target.Key(); key == "" {
		return ErrNoTargetKey
	} else {
		if HasTarget(group, key) && !replace {
			return ErrTargetAlreadyExists
		}
		cacheGroup, found := targetCache[group]
		if !found {
			targetCache[group] = make(map[string]Target)
			cacheGroup = targetCache[group]
		}
		cacheGroup[key] = target
		return nil
	}
}
