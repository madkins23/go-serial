package proxy

import (
	"fmt"

	"github.com/madkins23/go-type/reg"
)

type Wrappable interface{}

type Wrapper[T Wrappable] struct {
	// Serialization requires these field names to be public.
	// Otherwise, we would prefer them to be private.

	TypeName string
	Contents T
}

// Wrap a Wrappable item in a wrapper that can handle serialization.
// When changing the contents of a wrapper prefer replacing the entire thing.
// Because otherwise multiple references point to the same object?
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

func (w Wrapper[V]) GetContents() V {
	return w.Contents
}
