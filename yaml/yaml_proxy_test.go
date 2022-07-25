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

// The (newer) "proxy" mechanism for YAML.
// The proxy way requires all code references to interface objects to dereference the proxy.
// On the other hand there is no need to write custom un/marshaling code for every
// struct that contains interface fields.
type YamlProxyTestSuite struct {
	suite.Suite
	showAccount bool
}

func (suite *YamlProxyTestSuite) SetupSuite() {
	if showAccount, found := os.LookupEnv("GO-TYPE-SHOW-ACCOUNT"); found {
		var err error
		suite.showAccount, err = strconv.ParseBool(showAccount)
		suite.Require().NoError(err)
	}
	reg.Highlander().Clear()
	suite.Require().NoError(test.Registration())
	suite.Require().NoError(reg.AddAlias("yamlProxyTest", ProxyAccount{}), "creating test alias")
	suite.Require().NoError(reg.Register(&ProxyAccount{}))
	suite.Require().NoError(reg.Register(&ProxyBond{}))
}

func TestProxyYamlSuite(t *testing.T) {
	suite.Run(t, new(YamlProxyTestSuite))
}

//////////////////////////////////////////////////////////////////////////

// TestMarshalCycle verifies the JSON Marshal/Unmarshal works as expected.
func (suite *YamlProxyTestSuite) TestMarshalCycle() {
	account := MakeProxyAccount()

	marshaled, err := yaml.Marshal(account)
	suite.Require().NoError(err)
	if suite.showAccount {
		fmt.Println(string(marshaled))
	}
	suite.Assert().Contains(string(marshaled), "type:")
	suite.Assert().Contains(string(marshaled), "data:")
	suite.Assert().Contains(string(marshaled), "[test]Stock")
	suite.Assert().Contains(string(marshaled), "[test]Federal")
	suite.Assert().Contains(string(marshaled), "[test]State")
	suite.Assert().Contains(string(marshaled), "[yamlProxyTest]ProxyBond")

	var newAccount ProxyAccount
	suite.Require().NoError(yaml.Unmarshal(marshaled, &newAccount))
	if suite.showAccount {
		fmt.Println("---------------------------")
		spew.Dump(newAccount)
	}

	suite.Assert().NotEqual(account, newAccount)
	account.Favorite.Get().ClearPrivateFields()
	for _, position := range account.Positions {
		position.Get().ClearPrivateFields()
	}
	for _, position := range account.Lookup {
		position.Get().ClearPrivateFields()
	}
	// Succeeds now that unexported (private) fields are gone.
	suite.Assert().Equal(account, &newAccount)
}

//////////////////////////////////////////////////////////////////////////

type ProxyAccount struct {
	// Can't embed test.Account since we're changing its fields.

	test.AccountData
	Favorite  *Wrapper[test.Investment]
	Positions []*Wrapper[test.Investment]
	Lookup    map[string]*Wrapper[test.Investment]
}

func MakeProxyAccount() *ProxyAccount {
	account := &ProxyAccount{}
	tBill := &ProxyBond{}
	tBill.ConfigureTBill()
	state := &ProxyBond{}
	state.ConfigureStateBond()
	account.MakeFake(test.MakeCostco(), test.MakeWalmart(), tBill, state)
	return account
}

func (pa *ProxyAccount) MakeFake(investments ...test.Investment) {
	acct := &test.Account{}
	acct.MakeFake(investments...)
	pa.AccountData = acct.AccountData
	pa.Favorite = Wrap[test.Investment](acct.Favorite)
	pa.Positions = make([]*Wrapper[test.Investment], len(investments))
	pa.Lookup = make(map[string]*Wrapper[test.Investment])
	for i, investment := range investments {
		pa.Positions[i] = Wrap[test.Investment](investment)
		switch it := investment.(type) {
		case *test.Stock:
			pa.Lookup[it.Symbol] = Wrap[test.Investment](investment)
		}
	}
}

//////////////////////////////////////////////////////////////////////////

type ProxyBond struct {
	// Can't embed test.Bond since we're changing its fields.

	Source *Wrapper[test.Borrower]
	Data   test.BondData
}

func (b *ProxyBond) CurrentValue() (float32, error) {
	return b.Data.Value, nil
}

func (b *ProxyBond) ClearPrivateFields() {
	b.Data.ClearPrivateFields()
}

func (b *ProxyBond) ConfigureTBill() {
	b.Source = Wrap[test.Borrower](test.TBillSource())
	b.Data = test.TBillBondData()
}

func (b *ProxyBond) ConfigureStateBond() {
	b.Source = Wrap[test.Borrower](test.StateBondSource())
	b.Data = test.TBillBondData()
}
