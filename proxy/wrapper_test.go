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
}

type MyGoober struct {
	//Goober // TODO: (what?)
	text   string
	number int
}

func (mg *MyGoober) Text() string {
	return mg.text
}

//////////////////////////////////////////////////////////////////////////

const testDataType = "MyGoober"
const testText = "test"
const testType = "type"
const testNum = 23

func (suite *WrapperTestSuite) TestWrapper() {
	g := new(MyGoober)
	g.text = testText
	g.number = testNum
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
	contents, ok := wg.Contents.(*MyGoober) // TODO: This syntax is a little painful.
	suite.Require().True(ok)
	suite.Require().NotNil(contents)
	suite.Assert().Equal(contents.text, testText)
	suite.Assert().Equal(contents.number, testNum)
	suite.Assert().True(strings.Contains(wg.TypeName, testDataType))

	x := wg.GetContents()
	suite.Require().NotNil(x)
}
