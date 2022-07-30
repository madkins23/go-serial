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
	suite.Require().NoError(reg.AddAlias("json", Bond{}), "creating json test alias")
	suite.Require().NoError(test.Register())
	suite.Require().NoError(reg.Register(Bond{}))
	suite.Require().NoError(reg.Register(WrappedBond{}))
}

func TestJsonSuite(t *testing.T) {
	suite.Run(t, new(JsonTestSuite))
}

//////////////////////////////////////////////////////////////////////////

func (suite *JsonTestSuite) TestWrapper() {
	stock := test.MakeCostco()
	suite.Require().NotNil(stock)
	suite.Assert().Equal(test.StockCostcoName, stock.Named)
	suite.Assert().Equal(test.StockCostcoSymbol, stock.Symbol)
	suite.Assert().Equal(test.StockCostcoShares, stock.Shares)
	suite.Assert().Equal(test.StockCostcoPrice, stock.Price)
	wrapped := Wrap(stock)
	suite.Require().NotNil(wrapped)
	suite.Assert().Equal(test.StockCostcoName, wrapped.Get().Named)
	suite.Assert().Equal(test.StockCostcoSymbol, wrapped.Get().Symbol)
	suite.Assert().Equal(test.StockCostcoShares, wrapped.Get().Shares)
	suite.Assert().Equal(test.StockCostcoPrice, wrapped.Get().Price)
	clearPacked := ClearPackedAfterMarshal
	ClearPackedAfterMarshal = false
	defer func() { ClearPackedAfterMarshal = clearPacked }()
	marshaledBytes, err := wrapped.MarshalJSON()
	suite.Require().NoError(err)
	marshaled := string(marshaledBytes)
	suite.Assert().Contains(marshaled, "type\":")
	suite.Assert().Contains(marshaled, "data\":")
	suite.Assert().Contains(marshaled, "[test]Stock")
	suite.Assert().Equal("[test]Stock", wrapped.Packed.TypeName)
	suite.Assert().Contains(string(wrapped.Packed.RawForm), test.StockCostcoName)
	suite.Assert().Contains(string(wrapped.Packed.RawForm), test.StockCostcoSymbol)
}

//------------------------------------------------------------------------

// TestNormal tests the "normal" case which requires custom un/marshaling.
// In this case the Portfolio fields do not need to be dereferenced.
// See the Portfolio MarshalJSON() and UnmarshalJSON() below.
func (suite *JsonTestSuite) TestNormal() {
	MarshalCycle[Portfolio](suite, MakePortfolio(),
		func(suite *JsonTestSuite, marshaled string) {
			suite.Assert().Contains(marshaled, "type\":")
			suite.Assert().Contains(marshaled, "data\":")
			suite.Assert().Contains(marshaled, "[test]Stock")
			suite.Assert().Contains(marshaled, "[test]Federal")
			suite.Assert().Contains(marshaled, "[test]State")
		},
		func(suite *JsonTestSuite, portfolio *Portfolio) {
			// In the "normal" case the portfolio fields are referenced directly.
			suite.Assert().Equal(test.StockCostcoName, portfolio.Favorite.Name())
			suite.Assert().Equal(test.StockCostcoShares*test.StockCostcoPrice, portfolio.Favorite.Value())
			suite.Assert().Equal(test.StockWalmartName, portfolio.Lookup[test.StockWalmartSymbol].Name())
			suite.Assert().Equal(test.StockWalmartShares*test.StockWalmartPrice, portfolio.Lookup[test.StockWalmartSymbol].Value())
		})
}

//------------------------------------------------------------------------

// TestWrapped tests the expected usage of json.Wrap() and json.Wrapper.
// In this case all references to interface values are wrapped.
func (suite *JsonTestSuite) TestWrapped() {
	MarshalCycle[WrappedPortfolio](suite, MakeWrappedPortfolio(),
		func(suite *JsonTestSuite, marshaled string) {
			suite.Assert().Contains(marshaled, "type\":")
			suite.Assert().Contains(marshaled, "data\":")
			suite.Assert().Contains(marshaled, "[test]Stock")
			suite.Assert().Contains(marshaled, "[test]Federal")
			suite.Assert().Contains(marshaled, "[test]State")
		},
		func(suite *JsonTestSuite, portfolio *WrappedPortfolio) {
			// In the "wrapped" case the portfolio fields must be dereferenced from their wrappers.
			suite.Assert().Equal(test.StockCostcoName, portfolio.Favorite.Get().Name())
			suite.Assert().Equal(test.StockCostcoShares*test.StockCostcoPrice, portfolio.Favorite.Get().Value())
			suite.Assert().Equal(test.StockWalmartName, portfolio.Lookup[test.StockWalmartSymbol].Get().Name())
			suite.Assert().Equal(test.StockWalmartShares*test.StockWalmartPrice, portfolio.Lookup[test.StockWalmartSymbol].Get().Value())
		})
}

//////////////////////////////////////////////////////////////////////////

// MarshalCycle has common code for testing a marshal/unmarshal cycle.
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

type Portfolio struct {
	Favorite  test.Investment
	Positions []test.Investment
	Lookup    map[string]test.Investment
}

//------------------------------------------------------------------------

func MakePortfolio() *Portfolio {
	return MakePortfolioWith(
		test.MakeCostco(), test.MakeWalmart(),
		MakeStateBond(), MakeTBill())
}

func MakePortfolioWith(investments ...test.Investment) *Portfolio {
	portfolio := &Portfolio{
		Positions: make([]test.Investment, len(investments)),
		Lookup:    make(map[string]test.Investment),
	}
	for i, investment := range investments {
		portfolio.Positions[i] = investment
		switch it := investment.(type) {
		case *test.Stock:
			portfolio.Lookup[it.Symbol] = investment
		}
		if i == 0 {
			portfolio.Favorite = investment
		}
	}
	return portfolio
}

//------------------------------------------------------------------------

// MarshalJSON is required in the "normal" case to generate a WrappedPortfolio which is then marshaled.
func (p *Portfolio) MarshalJSON() ([]byte, error) {
	w := &WrappedPortfolio{
		Positions: make([]*Wrapper[test.Investment], len(p.Positions)),
		Lookup:    make(map[string]*Wrapper[test.Investment], len(p.Positions)),
	}
	for i, position := range p.Positions {
		w.Positions[i] = Wrap[test.Investment](position)
		if key := position.Key(); key != "" {
			w.Lookup[key] = w.Positions[i]
		}
		if i == 0 {
			w.Favorite = w.Positions[i]
		}
	}
	return json.Marshal(w)
}

// UnmarshalJSON is required in the "normal" case to convert the WrappedPortfolio into a Portfolio.
func (p *Portfolio) UnmarshalJSON(marshaled []byte) error {
	w := new(WrappedPortfolio)
	if err := json.Unmarshal(marshaled, w); err != nil {
		return err
	}
	p.Lookup = make(map[string]test.Investment, len(w.Lookup))
	for k, position := range w.Lookup {
		p.Lookup[k] = position.Get()
	}
	p.Positions = make([]test.Investment, len(w.Positions))
	for i, position := range w.Positions {
		key := position.Get().Key()
		if key != "" {
			if pos, found := p.Lookup[key]; found {
				p.Positions[i] = pos
				continue
			}
		}
		p.Positions[i] = position.Get()
	}
	p.Favorite = p.Positions[0]
	return nil
}

//========================================================================

type WrappedPortfolio struct {
	Favorite  *Wrapper[test.Investment]
	Positions []*Wrapper[test.Investment]
	Lookup    map[string]*Wrapper[test.Investment]
}

func MakeWrappedPortfolio() *WrappedPortfolio {
	return MakeWrappedPortfolioWith(
		test.MakeCostco(), test.MakeWalmart(),
		MakeWrappedStateBond(), MakeWrappedTBill())
}

func MakeWrappedPortfolioWith(investments ...test.Investment) *WrappedPortfolio {
	p := &WrappedPortfolio{
		Positions: make([]*Wrapper[test.Investment], len(investments)),
		Lookup:    make(map[string]*Wrapper[test.Investment]),
	}
	for i, investment := range investments {
		wrapped := Wrap[test.Investment](investment)
		p.Positions[i] = wrapped
		if stock, ok := wrapped.Get().(*test.Stock); ok {
			p.Lookup[stock.Symbol] = wrapped
		}
		if i == 0 {
			p.Favorite = wrapped
		}
	}
	return p
}

//////////////////////////////////////////////////////////////////////////
// Bonds contain an interface type Borrower which tests nested interface objects.

var _ test.Investment = &Bond{}

type Bond struct {
	test.BondData
	Source test.Borrower
}

func MakeStateBond() *Bond {
	return &Bond{
		BondData: test.StateBondData(),
		Source:   test.StateBondSource(),
	}
}

func MakeTBill() *Bond {
	return &Bond{
		BondData: test.TBillData(),
		Source:   test.TBillSource(),
	}
}

//------------------------------------------------------------------------

// MarshalJSON is required in the "normal" case to generate a WrappedBond which is then marshaled.
func (b *Bond) MarshalJSON() ([]byte, error) {
	w := &WrappedBond{
		BondData: b.BondData,
		Source:   Wrap[test.Borrower](b.Source),
	}
	return json.Marshal(w)
}

// UnmarshalJSON is required in the "normal" case to convert the WrappedBond into a Bond.
func (b *Bond) UnmarshalJSON(marshaled []byte) error {
	w := new(WrappedBond)
	if err := json.Unmarshal(marshaled, w); err != nil {
		return err
	}
	b.BondData = w.BondData
	b.Source = w.Source.Get()
	return nil
}

//========================================================================

var _ test.Investment = &WrappedBond{}

type WrappedBond struct {
	test.BondData
	Source *Wrapper[test.Borrower]
}

func (b *WrappedBond) Value() float32 {
	return float32(b.BondData.Units) * b.BondData.Price
}

func MakeWrappedStateBond() *WrappedBond {
	return &WrappedBond{
		BondData: test.StateBondData(),
		Source:   Wrap[test.Borrower](test.StateBondSource()),
	}
}

func MakeWrappedTBill() *WrappedBond {
	return &WrappedBond{
		BondData: test.TBillData(),
		Source:   Wrap[test.Borrower](test.TBillSource()),
	}
}
