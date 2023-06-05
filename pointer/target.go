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

// Finder returns a new Target item with the specified key to fill in the cache.
// This method can be defined to pull items out of a DB or other source.
type Finder func(key string) (Target, error)

var (
	// ErrNoSuchTarget is returned from GetTarget when no target is found.
	ErrNoSuchTarget = errors.New("no such target")

	// ErrNoTargetGroup is returned from SetTarget when the target has an empty group ("").
	ErrNoTargetGroup = errors.New("no group for target")

	// ErrNoTargetKey is returned from SetTarget when the target has an empty key ("").
	ErrNoTargetKey = errors.New("no key for target")

	// ErrTargetIsNil is returned from SetTarget if the specified Target is nil.
	ErrTargetIsNil = errors.New("target is nil")

	// ErrFinderTargetIsNil is returned from SetTarget if the specified Target is nil.
	ErrFinderTargetIsNil = errors.New("target returned by finder is nil")

	// ErrTargetAlreadyExists is returned from SetTarget if the cache already
	// has a Target for the new Target's group and key and the replace flag is false.
	ErrTargetAlreadyExists = errors.New("target already exists")
)

// Internal cache for Target items.
var cache = make(map[string]map[string]Target)

// ClearCache removes all cache entries.
// For test purposes, all other usage suspect.
func ClearCache() {
	cache = make(map[string]map[string]Target)
}

// HasTarget returns true if the specified group and key have a Target in the cache.
func HasTarget(group, key string) bool {
	target, found := cache[group][key]
	return found && target != nil
}

// GetTarget returns a Target object from the cache for use in Pointer implementations.
// If there is no such Target and no Finder the ErrNoSuchTarget error is returned.
// If the Finder is used to acquire the Target it is added to the cache and returned.
func GetTarget(group, key string, finder Finder) (Target, error) {
	if target, found := cache[group][key]; found && target != nil {
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

// SetTarget adds the specified Target to the cache.
// Use this function for Pointer implementations and preloading the cache.
func SetTarget(target Target, replace bool) error {
	if target == nil {
		return ErrTargetIsNil
	} else if group := target.Group(); group == "" {
		return ErrNoTargetGroup
	} else if key := target.Key(); key == "" {
		return ErrNoTargetKey
	} else {
		if _, err := GetTarget(group, key, nil); err == nil {
			if !replace {
				return ErrTargetAlreadyExists
			}
		} else if !errors.Is(err, ErrNoSuchTarget) {
			return fmt.Errorf("get target before set: %w", err)
		}
		cacheGroup, found := cache[group]
		if !found {
			cache[group] = make(map[string]Target)
			cacheGroup = cache[group]
		}
		cacheGroup[key] = target
		return nil
	}
}
