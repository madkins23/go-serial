package json

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/madkins23/go-type/reg"
)

// Wrapper is used to attach a type name to an item to be serialized.
// This supports re-creating the correct type for filling an interface field.
type Wrapper struct {
	TypeName string
	Contents json.RawMessage
}

// WrapItem returns the specified item wrapped for serialization.
func WrapItem(item interface{}) (*Wrapper, error) {
	w := &Wrapper{}
	var err error
	if w.TypeName, err = reg.NameFor(item); err != nil {
		return nil, fmt.Errorf("get type name for %#v: %w", item, err)
	}

	build := &strings.Builder{}
	encoder := json.NewEncoder(build)
	encoder.SetEscapeHTML(false)
	if err = encoder.Encode(item); err != nil {
		return nil, fmt.Errorf("marshal wrapper contents: %w", err)
	}
	w.Contents = []byte(build.String())

	return w, nil
}

// Unwrap returns the object contained in the wrapper.
// The type name contained in the wrapper is used to
// create an appropriate instance to which the JSON contents are decoded.
func (w *Wrapper) Unwrap() (interface{}, error) {
	if w.TypeName == "" {
		return nil, fmt.Errorf("empty type field")
	} else if item, err := reg.Make(w.TypeName); err != nil {
		return nil, fmt.Errorf("make instance of type %s: %w", w.TypeName, err)
	} else if err = json.NewDecoder(strings.NewReader(string(w.Contents))).Decode(item); err != nil {
		return nil, fmt.Errorf("decode wrapper contents: %w", err)
	} else {
		return item, nil
	}
}
