package proxy

import (
	"fmt"

	"github.com/madkins23/go-type/reg"
)

// Wrappable provides the interface for objects that can be wrapped.
//  TODO(mAdkins): is this necessary?
type Wrappable interface{}

type Wrapper[T Wrappable] struct {
	TypeName string
	Contents T
	// Note: Serialization requires these field names to be public.
	//       Otherwise, we would prefer them to be private.
}

// Wrap a Wrappable item in a wrapper that can handle serialization.
func Wrap[V Wrappable](contents V) (*Wrapper[V], error) {
	if typeName, err := reg.NameFor(contents); err != nil {
		return nil, fmt.Errorf("get type name for %#v: %w", contents, err)
	} else {
		return &Wrapper[V]{
			TypeName: typeName,
			Contents: contents,
		}, nil
	}
}

// Get the contents of a wrapped item (the item itself).
func (w Wrapper[V]) Get() V {
	return w.Contents
}
