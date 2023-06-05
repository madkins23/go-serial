package pointer

import "errors"

// Finder returns a new Target item with the specified key to fill in the targetCache.
// This method can be defined to pull items out of a DB or other source.
type Finder func(key string) (Target, error)

// -----------------------------------------------------------------------

var (
	// ErrFinderIsNil is returned from SetFinder if the specified Finder is nil.
	ErrFinderIsNil = errors.New("finder is nil")

	// ErrNoFinderGroup is returned from SetFinder when the specified group is empty ("").
	ErrNoFinderGroup = errors.New("empty group for finder")

	// ErrFinderAlreadyExists is returned from SetFinder if the targetCache already
	// has a Target for the new Target's group and key and the replace flag is false.
	ErrFinderAlreadyExists = errors.New("finder already exists")
)

// -----------------------------------------------------------------------

var finderCache = make(map[string]Finder)

// ClearFinderCache clears the finderCache of Finder functions by group.
func ClearFinderCache() {
	finderCache = make(map[string]Finder)
}

// HasFinder returns true if the specified group have a Finder in the Finder cache.
func HasFinder(group string) bool {
	finder, found := finderCache[group]
	return found && finder != nil
}

// GetFinder acquires a pointer.Finder by group.
func GetFinder(group string) Finder {
	if finder, found := finderCache[group]; found {
		return finder
	} else {
		return nil
	}
}

// SetFinder configures a pointer.Finder for the specified group.
func SetFinder(group string, finder Finder, replace bool) error {
	if group == "" {
		return ErrNoFinderGroup
	} else if finder == nil {
		return ErrFinderIsNil
	} else if HasFinder(group) && !replace {
		return ErrFinderAlreadyExists
	}

	finderCache[group] = finder
	return nil
}
