package proxy

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/madkins23/go-type/reg"
)

type WrapperTestSuite struct {
	suite.Suite
}

func (suite *WrapperTestSuite) SetupSuite() {
	suite.Require().NoError(reg.Register(&MyGoober{}))
}

func TestWrapperSuite(t *testing.T) {
	suite.Run(t, new(WrapperTestSuite))
}

//////////////////////////////////////////////////////////////////////////

var _ = (Goober)(&MyGoober{})

type Goober interface {
	Wrappable

	Name() string
	Age() uint8
}

type MyGoober struct {
	Goober
	name string
	age  uint8
}

func (mg *MyGoober) Name() string {
	return mg.name
}

func (mg *MyGoober) Age() uint8 {
	return mg.age
}

//////////////////////////////////////////////////////////////////////////

const testDataType = "MyGoober"
const testName = "test"
const testType = "type"
const testAge = uint8(23)

func (suite *WrapperTestSuite) TestWrapper() {
	g := new(MyGoober)
	g.name = testName
	g.age = testAge
	w := &Wrapper[Goober]{
		TypeName: testType,
		Contents: g,
	}
	suite.Require().NotNil(w)
	suite.Assert().Equal(testType, w.TypeName)
	suite.Require().NotNil(w.Contents)
	wg, err := Wrap[Goober](g)
	suite.Require().NoError(err)
	suite.Require().NotNil(wg)
	gc := wg.Contents
	suite.Require().NotNil(gc)
	suite.Assert().Equal(gc.Name(), testName)
	suite.Assert().Equal(gc.Age(), testAge)
	contents, ok := gc.(*MyGoober) // TODO: This syntax is a little painful.
	suite.Require().True(ok)
	suite.Require().NotNil(contents)
	suite.Assert().Equal(contents.name, testName)
	suite.Assert().Equal(contents.age, testAge)
	suite.Assert().True(strings.Contains(wg.TypeName, testDataType))

	x := wg.Contents
	suite.Require().NotNil(x)
}
