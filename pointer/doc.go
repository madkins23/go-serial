// Package pointer supports serialization and deserialization of pointer references.
//
// A pointer reference is stored as a combination of group and key strings.
// When referenced later the Target item is pulled from an internal targetCache.
// The targetCache may be preloaded or a Finder function may be used to load Target
// items into the targetCache dynamically as they are referenced.
//
// This package defines the Target interface.
// Pointer implementations are defined in the json and yaml packages.
//
// Note: Target items used in Pointer references must be unique and permanent.
// There is no 'listener' mechanism to cause Target items to be updated.
// Do not use Pointer references for large domain or mutable DB objects.

package pointer
