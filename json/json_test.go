package json

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/stretchr/testify/suite"

	"github.com/madkins23/go-type/reg"

	"github.com/madkins23/go-serial/test"
)

type JsonTestSuite struct {
	suite.Suite
	showSerialized bool
}

func (suite *JsonTestSuite) SetupSuite() {
	if showSerialized, found := os.LookupEnv("GO-TYPE-SHOW-SERIALIZED"); found {
		var err error
		suite.showSerialized, err = strconv.ParseBool(showSerialized)
		suite.Require().NoError(err)
	}
	reg.Highlander().Clear()
	suite.Require().NoError(test.Register())
	//suite.Require().NoError(reg.AddAlias("jsonManualTest", ManualAccount{}),
	//	"creating json manual test alias")
	//suite.Require().NoError(reg.Register(&ManualAccount{}))
	//suite.Require().NoError(reg.Register(&ManualBond{}))
}

func TestJsonSuite(t *testing.T) {
	suite.Run(t, new(JsonTestSuite))
}

//////////////////////////////////////////////////////////////////////////

func (suite *JsonTestSuite) TestWrapper() {
	cat := test.NewCat(test.CatName)
	suite.Require().NotNil(cat)
	suite.Assert().Equal(test.CatName, cat.Name())
	suite.Assert().Equal(test.CatMoves, cat.Moves())
	suite.Assert().Equal(test.CatSound, cat.Sound())
	wrapCat := Wrap(cat)
	suite.Require().NotNil(wrapCat)
	suite.Assert().Equal(test.CatName, wrapCat.Get().Name())
	suite.Assert().Equal(test.CatMoves, wrapCat.Get().Moves())
	suite.Assert().Equal(test.CatSound, wrapCat.Get().Sound())
	ClearPackedAfterMarshal = false
	defer func() { ClearPackedAfterMarshal = true }()
	marshaledBytes, err := wrapCat.MarshalJSON()
	suite.Require().NoError(err)
	marshaled := string(marshaledBytes)
	suite.Assert().Contains(marshaled, "type\":")
	suite.Assert().Contains(marshaled, "data\":")
	suite.Assert().Contains(marshaled, "[test]Cat")
	suite.Assert().Equal("[test]Cat", wrapCat.Packed.TypeName)
	suite.Assert().Contains(string(wrapCat.Packed.RawForm), test.CatName)
}

//------------------------------------------------------------------------

// TestWrapped tests the expected usage of json.Wrap() and json.Wrapper.
// In this case all references to interface values are wrapped.
func (suite *JsonTestSuite) TestWrapped() {
	MarshalCycle(suite, MakeWrappedZoo(),
		func(suite *JsonTestSuite, marshaled string) {
			suite.Assert().Contains(marshaled, "type\":")
			suite.Assert().Contains(marshaled, "data\":")
			suite.Assert().Contains(marshaled, "[test]Bird")
			suite.Assert().Contains(marshaled, "[test]Cat")
			suite.Assert().Contains(marshaled, "[test]Dog")
		},
		func(suite *JsonTestSuite, zoo *WrappedZoo) {
			// In the "wrapped" case the zoo fields must be dereferenced from their wrappers.
			suite.Assert().Equal(test.BirdName, zoo.Favorite.Get().Name())
			suite.Assert().Equal(test.BirdMoves, zoo.Favorite.Get().Moves())
			suite.Assert().Equal(test.BirdSound, zoo.Favorite.Get().Sound())
			suite.Assert().Equal(test.BirdSound, zoo.Named[test.BirdName].Get().Sound())
			suite.Assert().Equal(test.CatSound, zoo.Named[test.CatName].Get().Sound())
			suite.Assert().Equal(test.DogSound, zoo.Named[test.DogName].Get().Sound())
		})
}

//------------------------------------------------------------------------

// TestNormal tests the "normal" case which requires custom un/marshaling.
// In this case the Zoo fields do not need to be dereferenced.
// See the Zoo MarshalJSON() and UnmarshalJSON() below.
func (suite *JsonTestSuite) TestNormal() {
	MarshalCycle(suite, MakeZoo(),
		func(suite *JsonTestSuite, marshaled string) {
			suite.Assert().Contains(marshaled, "type\":")
			suite.Assert().Contains(marshaled, "data\":")
			suite.Assert().Contains(marshaled, "[test]Bird")
			suite.Assert().Contains(marshaled, "[test]Cat")
			suite.Assert().Contains(marshaled, "[test]Dog")
		},
		func(suite *JsonTestSuite, zoo *Zoo) {
			// In the "normal" case the Zoo fields are referenced directly.
			suite.Assert().Equal(test.BirdName, zoo.Favorite.Name())
			suite.Assert().Equal(test.BirdMoves, zoo.Favorite.Moves())
			suite.Assert().Equal(test.BirdSound, zoo.Favorite.Sound())
			suite.Assert().Equal(test.BirdSound, zoo.Named[test.BirdName].Sound())
			suite.Assert().Equal(test.CatSound, zoo.Named[test.CatName].Sound())
			suite.Assert().Equal(test.DogSound, zoo.Named[test.DogName].Sound())
		})
}

//////////////////////////////////////////////////////////////////////////

func MarshalCycle[T any](suite *JsonTestSuite, data *T,
	marshaledTests func(suite *JsonTestSuite, marshaled string),
	unmarshaledTests func(suite *JsonTestSuite, unmarshaled *T)) {
	marshaled, err := json.Marshal(data)
	suite.Require().NoError(err)
	suite.Require().NotNil(marshaled)
	if suite.showSerialized {
		var buf bytes.Buffer
		suite.Require().NoError(json.Indent(&buf, marshaled, "", "  "))
		fmt.Println(buf.String())
	}
	if marshaledTests != nil {
		marshaledTests(suite, string(marshaled))
	}

	newData := new(T)
	suite.Require().NotNil(newData)
	suite.Require().NoError(json.Unmarshal(marshaled, newData))
	if suite.showSerialized {
		fmt.Println("---------------------------")
		spew.Dump(newData)
	}
	suite.Assert().Equal(data, newData)
	if unmarshaledTests != nil {
		unmarshaledTests(suite, newData)
	}
}

//////////////////////////////////////////////////////////////////////////

type Zoo struct {
	Favorite test.Animal
	Animals  []test.Animal
	Named    map[string]test.Animal
}

func MakeZoo() *Zoo {
	animals := test.Animals()
	zoo := &Zoo{
		Favorite: animals[0],
		Animals:  animals,
		Named:    make(map[string]test.Animal, len(animals)),
	}
	for _, animal := range animals {
		zoo.Named[animal.Name()] = animal
	}
	return zoo
}

// MarshalJSON is required in the "normal" case to generate a WrappedZoo which is then marshaled.
func (z *Zoo) MarshalJSON() ([]byte, error) {
	w := &WrappedZoo{
		Animals: make([]*Wrapper[test.Animal], len(z.Animals)),
		Named:   make(map[string]*Wrapper[test.Animal], len(z.Animals)),
	}
	for i, animal := range z.Animals {
		w.Animals[i] = Wrap[test.Animal](animal)
		w.Named[animal.Name()] = w.Animals[i]
		if i == 0 {
			w.Favorite = w.Animals[i]
		}
	}
	return json.Marshal(w)
}

// UnmarshalJSON is required in the "normal" case to convert the WrappedZoo into a Zoo.
func (z *Zoo) UnmarshalJSON(marshaled []byte) error {
	w := new(WrappedZoo)
	if err := json.Unmarshal(marshaled, w); err != nil {
		return fmt.Errorf("unmarshal packed area: %w", err)
	}
	z.Named = make(map[string]test.Animal, len(w.Named))
	for k, animal := range w.Named {
		z.Named[k] = animal.Get()
	}
	z.Animals = make([]test.Animal, len(w.Animals))
	for i, animal := range w.Animals {
		if a, found := z.Named[animal.Get().Name()]; found {
			z.Animals[i] = a
		} else {
			z.Animals[i] = animal.Get()
		}
	}
	z.Favorite = z.Animals[0]
	return nil
}

//------------------------------------------------------------------------

type WrappedZoo struct {
	Favorite *Wrapper[test.Animal]
	Animals  []*Wrapper[test.Animal]
	Named    map[string]*Wrapper[test.Animal]
}

func MakeWrappedZoo() *WrappedZoo {
	testAnimals := test.Animals()
	zoo := &WrappedZoo{
		Animals: make([]*Wrapper[test.Animal], len(testAnimals)),
		Named:   make(map[string]*Wrapper[test.Animal]),
	}
	for i, animal := range testAnimals {
		zoo.Animals[i] = Wrap[test.Animal](animal)
		zoo.Named[zoo.Animals[i].Get().Name()] = zoo.Animals[i]
		if i == 0 {
			zoo.Favorite = zoo.Animals[i]
		}
	}
	return zoo
}
