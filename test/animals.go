package test

import (
	"fmt"

	"github.com/madkins23/go-type/reg"
)

// Declarations used by test files.

// Register adds the 'test' alias and registers several structs.
// Uses the github.com/madkins23/go-type library to register structs by name.
func Register() error {
	if err := reg.AddAlias("test", &Bird{}); err != nil {
		return fmt.Errorf("adding 'test' alias: %w", err)
	}
	if err := reg.Register(&Bird{}); err != nil {
		return fmt.Errorf("registering Bird struct: %w", err)
	}
	if err := reg.Register(&Cat{}); err != nil {
		return fmt.Errorf("registering Cat struct: %w", err)
	}
	if err := reg.Register(&Dog{}); err != nil {
		return fmt.Errorf("registering Dog struct: %w", err)
	}
	return nil
}

type Animal interface {
	Name() string
	Moves() string
	Sound() string
}

type animal struct {
	Named string
	// TODO: Moves should be a nested interface object
}

func (a *animal) Name() string {
	return a.Named
}

func Animals() []Animal {
	return []Animal{
		Animal(NewBird(BirdName)),
		Animal(NewCat(CatName)),
		Animal(NewDog(DogName)),
	}
}

//------------------------------------------------------------------------

const (
	BirdName  = "Pretty"
	BirdMoves = "Flies"
	BirdSound = "Chirp"
)

var _ Animal = &Bird{}

type Bird struct {
	animal
}

func NewBird(named string) *Bird {
	return &Bird{animal{Named: named}}
}

func (c *Bird) Moves() string {
	return BirdMoves
}

func (c *Bird) Sound() string {
	return BirdSound
}

//------------------------------------------------------------------------

const (
	CatName  = "Kitty"
	CatMoves = "Walks"
	CatSound = "Meow"
)

var _ Animal = &Cat{}

type Cat struct {
	animal
}

func NewCat(named string) *Cat {
	return &Cat{animal{Named: named}}
}

func (c *Cat) Moves() string {
	return CatMoves
}

func (c *Cat) Sound() string {
	return CatSound
}

//------------------------------------------------------------------------

const (
	DogName  = "Rover"
	DogMoves = "Runs"
	DogSound = "Bark"
)

var _ Animal = &Dog{}

type Dog struct {
	animal
}

func NewDog(named string) *Dog {
	return &Dog{animal{Named: named}}
}

func (c *Dog) Moves() string {
	return DogMoves
}

func (d *Dog) Sound() string {
	return DogSound
}
