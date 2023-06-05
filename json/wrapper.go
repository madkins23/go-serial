package json

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/madkins23/go-type/reg"
)

// Wrap an item in a JSON wrapper that can handle serialization.
func Wrap[W any](item W) *Wrapper[W] {
	w := new(Wrapper[W])
	w.Set(item)
	return w
}

// Wrapper is used to attach a type name to an item to be serialized.
// This supports re-creating the correct type for filling an interface field.
type Wrapper[T any] struct {
	item T
}

// Get the wrapped item.
func (w *Wrapper[T]) Get() T {
	return w.item
}

// Set the wrapped item.
func (w *Wrapper[T]) Set(t T) {
	w.item = t
}

// -----------------------------------------------------------------------

type packed struct {
	TypeName string          `json:"type"`
	RawForm  json.RawMessage `json:"data"`
}

func (w *Wrapper[T]) MarshalJSON() ([]byte, error) {
	var err error
	var pack packed
	if pack.TypeName, err = reg.NameFor(w.item); err != nil {
		return nil, fmt.Errorf("get type name for %#v: %w", w.item, err)
	}

	build := &strings.Builder{}
	encoder := json.NewEncoder(build)
	encoder.SetEscapeHTML(false)
	if err = encoder.Encode(w.item); err != nil {
		return nil, fmt.Errorf("marshal packed area: %w", err)
	}
	// Must get rid of extraneous ending newline that is not unmarshaled.
	pack.RawForm = []byte(strings.TrimSuffix(build.String(), "\n"))

	var marshaled []byte
	marshaled, err = json.Marshal(pack)
	if err != nil {
		return []byte(""), fmt.Errorf("marshal packed form: %w", err)
	}
	return marshaled, nil
}

var errEmptyTypeField = errors.New("empty type field")

func (w *Wrapper[T]) UnmarshalJSON(marshaled []byte) error {
	var pack packed
	if err := json.Unmarshal(marshaled, &pack); err != nil {
		return fmt.Errorf("unmarshal packed area: %w", err)
	}

	var ok bool
	if pack.TypeName == "" {
		return errEmptyTypeField
	} else if temp, err := reg.Make(pack.TypeName); err != nil {
		return fmt.Errorf("make instance of type %s: %w", pack.TypeName, err)
	} else if err = json.NewDecoder(strings.NewReader(string(pack.RawForm))).Decode(&temp); err != nil {
		return fmt.Errorf("decode wrapper contents: %w", err)
	} else if w.item, ok = temp.(T); !ok {
		// TODO: How to get name of T?
		return fmt.Errorf("type %s not generic type", pack.TypeName)
	} else {
		return nil
	}
}
