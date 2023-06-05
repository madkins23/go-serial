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

	"github.com/madkins23/go-serial/pointer"
	"github.com/madkins23/go-serial/test"
)

type JsonPointerTestSuite struct {
	suite.Suite
	showSerialized bool
}

func (suite *JsonPointerTestSuite) SetupSuite() {
	if showSerialized, found := os.LookupEnv("GO-TYPE-SHOW-SERIALIZED"); found {
		var err error
		suite.showSerialized, err = strconv.ParseBool(showSerialized)
		suite.Require().NoError(err)
	}
	pointer.ClearTargetCache()
	suite.Require().NoError(test.CachePets())
}

func TestJsonPointerSuite(t *testing.T) {
	suite.Run(t, new(JsonPointerTestSuite))
}

//////////////////////////////////////////////////////////////////////////

func (suite *JsonPointerTestSuite) TestPointer() {
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

func (suite *JsonPointerTestSuite) TestMarshalCycle() {
	start := makeAnimals()
	marshaled, err := json.Marshal(start)
	suite.Require().NoError(err)
	suite.Require().NotNil(marshaled)
	if suite.showSerialized {
		var buf bytes.Buffer
		suite.Require().NoError(json.Indent(&buf, marshaled, "", "  "))
		fmt.Println(buf.String())
	}

	finish := new(animals)
	suite.Require().NotNil(finish)
	suite.Require().NoError(json.Unmarshal(marshaled, finish))
	if suite.showSerialized {
		fmt.Println("---------------------------")
		spew.Dump(finish)
	}

	suite.Require().Equal(start, finish)
}
