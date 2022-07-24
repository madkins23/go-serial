package proxy

import (
	"fmt"

	"github.com/madkins23/go-type/reg"
	"github.com/madkins23/go-utils/check"
)

// Wrapper around an item to be serialized.
// The item will be represented by an interface.
type Wrapper[T Wrappable] interface {
	// Get the wrapped item.
	Get() T

	// Set the wrapped item.
	Set(T)
}

// Wrappable provides the interface for objects that can be wrapped.
//  TODO(mAdkins): is this necessary?
type Wrappable interface {
}

//////////////////////////////////////////////////////////////////////////

// Wrap a Wrappable item in a wrapper that can handle serialization.
func Wrap[W Wrappable](item W) *wrapper[W] {
	w := new(wrapper[W])
	w.Set(item)
	return w
}

// Default proxy.Wrapper object.
// TODO: is this only used for testing?
type wrapper[T Wrappable] struct {
	typeName string
	item     T
}

var _ = (Wrapper[Wrappable])(&wrapper[Wrappable]{})

// Get the wrapped item.
func (w *wrapper[T]) Get() T {
	return w.item
}

// Set the wrapped item.
func (w *wrapper[T]) Set(t T) {
	w.item = t
}

// Pack prepares the item for serialization if necessary.
func (w *wrapper[T]) Pack() error { // Nothing to do here.
	if check.IsZero(w.item) {
		return check.ErrIsZero
	}
	var err error
	if w.typeName, err = reg.NameFor(w.item); err != nil {
		return fmt.Errorf("get type name for %#v: %w", w.item, err)
	}
	return nil
}

// Unpack converts deserialized data back into the item if necessary.
func (w *wrapper[T]) Unpack() error {
	if check.IsZero(w.item) {
		return check.ErrIsZero
	}
	return nil
}
