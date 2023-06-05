package yaml

import (
	"fmt"
	"os"
	"strconv"
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/stretchr/testify/suite"
	"gopkg.in/yaml.v3"

	"github.com/madkins23/go-serial/pointer"
	"github.com/madkins23/go-serial/test"
)

type YamlPointerTestSuite struct {
	suite.Suite
	showSerialized bool
}

func (suite *YamlPointerTestSuite) SetupSuite() {
	if showSerialized, found := os.LookupEnv("GO-TYPE-SHOW-SERIALIZED"); found {
		var err error
		suite.showSerialized, err = strconv.ParseBool(showSerialized)
		suite.Require().NoError(err)
	}
	pointer.ClearTargetCache()
	suite.Require().NoError(test.CachePets())
}

func TestYamlPointerSuite(t *testing.T) {
	suite.Run(t, new(YamlPointerTestSuite))
}

//////////////////////////////////////////////////////////////////////////

func (suite *YamlPointerTestSuite) TestPointer() {
	ptr := Point[*test.Pet](test.Lacey)
	suite.Assert().Equal(test.Lacey, ptr.Get())
	ptr.Set(test.Noah)
	suite.Assert().Equal(test.Noah, ptr.Get())
}

type animals struct {
	Cats []*Pointer[*test.Pet]
	Dog  *Pointer[*test.Pet]
}

func makeAnimals() *animals {
	return &animals{
		Cats: []*Pointer[*test.Pet]{
			Point[*test.Pet](test.Noah),
			Point[*test.Pet](test.Lacey),
			Point[*test.Pet](test.Orca),
		},
		Dog: Point[*test.Pet](test.Knight),
	}
}

func (suite *YamlPointerTestSuite) TestMarshalCycle() {
	start := makeAnimals()
	marshaled, err := yaml.Marshal(start)
	suite.Require().NoError(err)
	suite.Require().NotNil(marshaled)
	if suite.showSerialized {
		fmt.Println(string(marshaled))
	}

	finish := new(animals)
	suite.Require().NotNil(finish)
	suite.Require().NoError(yaml.Unmarshal(marshaled, finish))
	if suite.showSerialized {
		fmt.Println("---------------------------")
		spew.Dump(finish)
	}

	suite.Require().Equal(start, finish)
}
