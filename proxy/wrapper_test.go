package proxy

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/madkins23/go-utils/check"

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
	name string
	age  uint8
}

func (mg *MyGoober) Name() string {
	return mg.name
}

func (mg *MyGoober) Age() uint8 {
	return mg.age
}

func (mg *MyGoober) Wrap() error {
	return nil
}

func (mg *MyGoober) Unwrap() error {
	return nil
}

//////////////////////////////////////////////////////////////////////////

const testDataType = "MyGoober"
const testName = "test"
const testAge = uint8(23)

func (suite *WrapperTestSuite) TestNoItem() {
	noItemWrapper := Wrap[Goober](nil)
	suite.Assert().ErrorIs(noItemWrapper.Wrap(), check.ErrIsZero)
	suite.Assert().ErrorIs(noItemWrapper.Unwrap(), check.ErrIsZero)
}

func (suite *WrapperTestSuite) TestWrapper() {
	myGoober := &MyGoober{
		name: testName,
		age:  testAge,
	}

	// Wrap the specific object.
	wrappedGoober := Wrap[Goober](myGoober)
	suite.Require().NotNil(wrappedGoober)

	suite.Assert().NoError(wrappedGoober.Wrap())
	// Serialization and deserialization would occur here.
	suite.Assert().NoError(wrappedGoober.Unwrap())

	// Get deserialized object.
	unwrappedGoober := wrappedGoober.Get()
	suite.Require().NotNil(unwrappedGoober)
	suite.Assert().Equal(unwrappedGoober.Name(), testName)
	suite.Assert().Equal(unwrappedGoober.Age(), testAge)

	// Convert interface object to struct object pointer.
	unwrappedMyGoober, ok := unwrappedGoober.(*MyGoober)
	suite.Require().True(ok)
	suite.Require().NotNil(unwrappedMyGoober)
	suite.Assert().Equal(unwrappedMyGoober.name, testName)
	suite.Assert().Equal(unwrappedMyGoober.age, testAge)
	suite.Assert().True(strings.Contains(wrappedGoober.typeName, testDataType))
}
