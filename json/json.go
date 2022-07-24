package json

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/madkins23/go-type/reg"

	"github.com/madkins23/go-serial/proxy"
)

// Wrap a Wrappable item in a JSON wrapper that can handle serialization.
// Creates a json.Wrapper object but doesn't Pack() it for serialization.
func Wrap[W proxy.Wrappable](item W) *Wrapper[W] {
	w := new(Wrapper[W])
	w.Set(item)
	return w
}

var _ = (proxy.Wrapper[proxy.Wrappable])(&Wrapper[proxy.Wrappable]{})

// Wrapper is used to attach a type name to an item to be serialized.
// This supports re-creating the correct type for filling an interface field.
type Wrapper[T proxy.Wrappable] struct {
	item   T
	Packed struct {
		TypeName string          `json:"type"`
		RawForm  json.RawMessage `json:"data"`
	}
}

// Get the wrapped item.
func (w *Wrapper[T]) Get() T {
	return w.item
}

// Set the wrapped item.
func (w *Wrapper[T]) Set(t T) {
	w.item = t
}

func (w *Wrapper[T]) MarshalJSON() ([]byte, error) {
	var err error
	if w.Packed.TypeName, err = reg.NameFor(w.item); err != nil {
		return nil, fmt.Errorf("get type name for %#v: %w", w.item, err)
	}

	build := &strings.Builder{}
	encoder := json.NewEncoder(build)
	encoder.SetEscapeHTML(false)
	if err = encoder.Encode(w.item); err != nil {
		return nil, fmt.Errorf("marshal packed area: %w", err)
	}
	w.Packed.RawForm = []byte(build.String())

	return json.Marshal(w.Packed)
}

func (w *Wrapper[T]) UnmarshalJSON(marshaled []byte) error {
	if err := json.Unmarshal(marshaled, &w.Packed); err != nil {
		return fmt.Errorf("unmarshal packed area: %w", err)
	}

	var ok bool
	if w.Packed.TypeName == "" {
		return fmt.Errorf("empty type field")
	} else if temp, err := reg.Make(w.Packed.TypeName); err != nil {
		return fmt.Errorf("make instance of type %s: %w", w.Packed.TypeName, err)
	} else if err = json.NewDecoder(strings.NewReader(string(w.Packed.RawForm))).Decode(&temp); err != nil {
		return fmt.Errorf("decode wrapper contents: %w", err)
	} else if w.item, ok = temp.(T); !ok {
		// TODO: How to get name of T?
		return fmt.Errorf("type %s not generic type", w.Packed.TypeName)
	} else {
		return nil
	}
}
