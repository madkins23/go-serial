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
	myGoober := new(MyGoober)
	myGoober.name = testName
	myGoober.age = testAge

	wrappedGoober, err := Wrap[Goober](myGoober)
	suite.Require().NoError(err)
	suite.Require().NotNil(wrappedGoober)

	unwrappedGoober := wrappedGoober.Get()
	suite.Require().NotNil(unwrappedGoober)
	suite.Assert().Equal(unwrappedGoober.Name(), testName)
	suite.Assert().Equal(unwrappedGoober.Age(), testAge)

	unwrappedMyGoober, ok := unwrappedGoober.(*MyGoober) // TODO: This syntax is a little painful.
	suite.Require().True(ok)
	suite.Require().NotNil(unwrappedMyGoober)
	suite.Assert().Equal(unwrappedMyGoober.name, testName)
	suite.Assert().Equal(unwrappedMyGoober.age, testAge)
	suite.Assert().True(strings.Contains(wrappedGoober.TypeName, testDataType))
}
