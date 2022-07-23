package json

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/madkins23/go-type/reg"

	"github.com/madkins23/go-serial/proxy"
)

// Wrap a Wrappable item in a wrapper that can handle serialization.
// Creates a proxy.Wrapper object but doesn't Wrap() it for serialization.
func Wrap[W proxy.Wrappable](item W) *wrapper[W] {
	w := new(wrapper[W])
	w.Set(item)
	return w
}

var _ = (proxy.Wrapper[proxy.Wrappable])(&wrapper[proxy.Wrappable]{})

// wrapper is used to attach a type name to an item to be serialized.
// This supports re-creating the correct type for filling an interface field.
type wrapper[T proxy.Wrappable] struct {
	TypeName string          `json:"type"`
	RawForm  json.RawMessage `json:"data"`
	item     T
}

// Get the wrapped item.
func (w *wrapper[T]) Get() T {
	return w.item
}

// Set the wrapped item.
func (w *wrapper[T]) Set(t T) {
	w.item = t
}

// Wrap prepares the item for serialization if necessary.
func (w *wrapper[T]) Wrap() error {
	var err error
	if w.TypeName, err = reg.NameFor(w.item); err != nil {
		return fmt.Errorf("get type name for %#v: %w", w.item, err)
	}

	build := &strings.Builder{}
	encoder := json.NewEncoder(build)
	encoder.SetEscapeHTML(false)
	if err = encoder.Encode(w.item); err != nil {
		return fmt.Errorf("marshal wrapper item: %w", err)
	}
	w.RawForm = []byte(build.String())

	return nil
}

// Unwrap converts deserialized data back into the item if necessary.
// The type name contained in the wrapper is used to
// create an appropriate instance to which the JSON contents are decoded.
func (w *wrapper[T]) Unwrap() error {
	var ok bool
	if w.TypeName == "" {
		return fmt.Errorf("empty type field")
	} else if temp, err := reg.Make(w.TypeName); err != nil {
		return fmt.Errorf("make instance of type %s: %w", w.TypeName, err)
	} else if err = json.NewDecoder(strings.NewReader(string(w.RawForm))).Decode(&temp); err != nil {
		return fmt.Errorf("decode wrapper contents: %w", err)
	} else if w.item, ok = temp.(T); !ok {
		// TODO: How to get name of T?
		return fmt.Errorf("type %s not generic type", w.TypeName)
	} else {
		return nil
	}
}
