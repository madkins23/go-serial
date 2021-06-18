package json

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/madkins23/go-type/reg"
)

type Wrapper struct {
	TypeName string
	Contents json.RawMessage
}

func WrapItem(item interface{}) (*Wrapper, error) {
	w := &Wrapper{}
	return w, w.Wrap(item)
}

func (w *Wrapper) Wrap(item interface{}) error {
	var err error
	if w.TypeName, err = reg.NameFor(item); err != nil {
		return fmt.Errorf("get type name for %#v: %w", item, err)
	}

	build := &strings.Builder{}
	encoder := json.NewEncoder(build)
	encoder.SetEscapeHTML(false)
	if err = encoder.Encode(item); err != nil {
		return fmt.Errorf("marshal wrapper contents: %w", err)
	}
	w.Contents = []byte(build.String())

	return nil
}

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
