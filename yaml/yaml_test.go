package yaml

import (
	"fmt"
	"os"
	"strconv"
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/stretchr/testify/suite"
	"gopkg.in/yaml.v3"

	"github.com/madkins23/go-type/reg"

	"github.com/madkins23/go-serial/test"
)

type YamlTestSuite struct {
	suite.Suite
	showSerialized bool
}

func (suite *YamlTestSuite) SetupSuite() {
	if showSerialized, found := os.LookupEnv("GO-TYPE-SHOW-SERIALIZED"); found {
		var err error
		suite.showSerialized, err = strconv.ParseBool(showSerialized)
		suite.Require().NoError(err)
	}
	reg.Highlander().Clear()
	suite.Require().NoError(reg.AddAlias("yaml", Bond{}), "creating yaml test alias")
	suite.Require().NoError(test.Register())
	suite.Require().NoError(reg.Register(Bond{}))
	suite.Require().NoError(reg.Register(WrappedBond{}))
}

func TestYamlSuite(t *testing.T) {
	suite.Run(t, new(YamlTestSuite))
}

//////////////////////////////////////////////////////////////////////////

func (suite *YamlTestSuite) TestWrapper() {
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
	packedVersion, err := wrapped.MarshalYAML()
	suite.Require().NoError(err)
	packed, ok := packedVersion.(*Packed)
	suite.Require().True(ok)
	suite.Assert().Equal(packed, &wrapped.Packed)
	suite.Assert().Equal("[test]Stock", packed.TypeName)
	suite.Assert().Contains(packed.RawForm, "market: "+test.MarketNASDAQ)
	suite.Assert().Contains(packed.RawForm, "named: "+test.StockCostcoName)
	suite.Assert().Contains(packed.RawForm, "symbol: "+test.StockCostcoSymbol)
}

//------------------------------------------------------------------------

// TestNormal tests the "normal" case which requires custom un/marshaling.
// In this case the Zoo fields do not need to be dereferenced.
// See the Zoo MarshalYaml() and UnmarshalYaml() below.
func (suite *YamlTestSuite) TestNormal() {
	MarshalCycle[Portfolio](suite, MakePortfolio(),
		func(suite *YamlTestSuite, marshaled string) {
			suite.Assert().Contains(marshaled, "type:")
			suite.Assert().Contains(marshaled, "data:")
			suite.Assert().Contains(marshaled, "[test]Stock")
			suite.Assert().Contains(marshaled, "[test]Federal")
			suite.Assert().Contains(marshaled, "[test]State")
		},
		func(suite *YamlTestSuite, portfolio *Portfolio) {
			// In the "normal" case the portfolio fields are referenced directly.
			suite.Assert().Equal(test.StockCostcoName, portfolio.Favorite.Name())
			suite.Assert().Equal(test.StockCostcoShares*test.StockCostcoPrice, portfolio.Favorite.Value())
			suite.Assert().Equal(test.StockWalmartName, portfolio.Lookup[test.StockWalmartSymbol].Name())
			suite.Assert().Equal(test.StockWalmartShares*test.StockWalmartPrice, portfolio.Lookup[test.StockWalmartSymbol].Value())
		})
}

//------------------------------------------------------------------------

// TestWrapped tests the expected usage of Yaml.Wrap() and Yaml.Wrapper.
// In this case all references to interface values are wrapped.
func (suite *YamlTestSuite) TestWrapped() {
	MarshalCycle[WrappedPortfolio](suite, MakeWrappedPortfolio(),
		func(suite *YamlTestSuite, marshaled string) {
			suite.Assert().Contains(marshaled, "type:")
			suite.Assert().Contains(marshaled, "data:")
			suite.Assert().Contains(marshaled, "[test]Stock")
			suite.Assert().Contains(marshaled, "[test]Federal")
			suite.Assert().Contains(marshaled, "[test]State")
		},
		func(suite *YamlTestSuite, portfolio *WrappedPortfolio) {
			// In the "wrapped" case the zoo fields must be dereferenced from their wrappers.
			suite.Assert().Equal(test.StockCostcoName, portfolio.Favorite.Get().Name())
			suite.Assert().Equal(test.StockCostcoShares*test.StockCostcoPrice, portfolio.Favorite.Get().Value())
			suite.Assert().Equal(test.StockWalmartName, portfolio.Lookup[test.StockWalmartSymbol].Get().Name())
			suite.Assert().Equal(test.StockWalmartShares*test.StockWalmartPrice, portfolio.Lookup[test.StockWalmartSymbol].Get().Value())
		})
}

//////////////////////////////////////////////////////////////////////////

// MarshalCycle has common code for testing a marshal/unmarshal cycle.
func MarshalCycle[T any](suite *YamlTestSuite, data *T,
	marshaledTests func(suite *YamlTestSuite, marshaled string),
	unmarshaledTests func(suite *YamlTestSuite, unmarshaled *T)) {
	marshaled, err := yaml.Marshal(data)
	suite.Require().NoError(err)
	suite.Require().NotNil(marshaled)
	if suite.showSerialized {
		fmt.Println(string(marshaled))
	}
	if marshaledTests != nil {
		marshaledTests(suite, string(marshaled))
	}

	newData := new(T)
	suite.Require().NotNil(newData)
	clearPacked := ClearPackedAfterUnmarshal
	ClearPackedAfterUnmarshal = false
	defer func() { ClearPackedAfterUnmarshal = clearPacked }()
	suite.Require().NoError(yaml.Unmarshal(marshaled, newData))
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

// MarshalYaml is required in the "normal" case to generate a WrappedPortfolio which is then marshaled.
func (p *Portfolio) MarshalYAML() (interface{}, error) {
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
	return w, nil
}

// UnmarshalYaml is required in the "normal" case to convert the WrappedPortfolio into a Portfolio.
func (p *Portfolio) UnmarshalYAML(node *yaml.Node) error {
	w := new(WrappedPortfolio)
	if err := node.Decode(w); err != nil {
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

// MarshalYaml is required in the "normal" case to generate a WrappedBond which is then marshaled.
func (b *Bond) MarshalYAML() (interface{}, error) {
	w := &WrappedBond{
		BondData: b.BondData,
		Source:   Wrap[test.Borrower](b.Source),
	}
	return w, nil
}

// UnmarshalYaml is required in the "normal" case to convert the WrappedBond into a Bond.
func (b *Bond) UnmarshalYAML(node *yaml.Node) error {
	w := new(WrappedBond)
	if err := node.Decode(w); err != nil {
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
