package yaml

import (
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/madkins23/go-type/reg"
)

// Wrap a Wrappable item in a wrapper that can handle serialization.
// Creates a proxy.Wrapper object but doesn't Wrap() it for serialization.
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

type packed struct {
	TypeName string `yaml:"type"`
	RawForm  string `yaml:"data"`
}

func (w *Wrapper[T]) MarshalYAML() (interface{}, error) {
	var err error
	var pack packed
	if pack.TypeName, err = reg.NameFor(w.item); err != nil {
		return nil, fmt.Errorf("get type name for %#v: %w", w.item, err)
	}

	build := &strings.Builder{}
	encoder := yaml.NewEncoder(build)
	if err = encoder.Encode(w.item); err != nil {
		return nil, fmt.Errorf("marshal packed area: %w", err)
	}
	pack.RawForm = build.String()
	return &pack, nil
}

func (w *Wrapper[T]) UnmarshalYAML(node *yaml.Node) error {
	var pack packed
	if err := node.Decode(&pack); err != nil {
		return fmt.Errorf("unmarshal packed area: %w", err)
	}

	var ok bool
	if pack.TypeName == "" {
		return fmt.Errorf("empty type field")
	} else if temp, err := reg.Make(pack.TypeName); err != nil {
		return fmt.Errorf("make instance of type %s: %w", pack.TypeName, err)
	} else if err = yaml.NewDecoder(strings.NewReader(pack.RawForm)).Decode(temp); err != nil {
		return fmt.Errorf("decode wrapper contents: %w", err)
	} else if w.item, ok = temp.(T); !ok {
		// TODO(mAdkins): How to get name of T? Do we care?
		return fmt.Errorf("type %s not generic type", pack.TypeName)
	} else {
		return nil
	}
}
