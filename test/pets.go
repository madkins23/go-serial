package test

import (
	"fmt"

	"github.com/madkins23/go-serial/pointer"
)

var (
	Lacey  = &Pet{Name: "Lacey", Type: "cat"}
	Noah   = &Pet{Name: "Noah", Type: "cat"}
	Orca   = &Pet{Name: "Orca", Type: "cat"}
	Knight = &Pet{Name: "Knight", Type: "dog"}
	pets   = []*Pet{Lacey, Noah, Orca, Knight}
)

// -----------------------------------------------------------------------

type Pet struct {
	Name string
	Type string
}

var _ pointer.Target = &Pet{}

func (p *Pet) Group() string {
	return p.Type
}

func (p *Pet) Key() string {
	return p.Name
}

// -----------------------------------------------------------------------

func CachePets() error {
	for _, pet := range pets {
		if err := pointer.SetTarget(pet, false); err != nil {
			return fmt.Errorf("set target %v: %w", pet, err)
		}
	}
	return nil
}
