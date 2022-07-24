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
	showAccount bool
}

func (suite *YamlTestSuite) SetupSuite() {
	if showAccount, found := os.LookupEnv("GO-TYPE-SHOW-ACCOUNT"); found {
		var err error
		suite.showAccount, err = strconv.ParseBool(showAccount)
		suite.Require().NoError(err)
	}
	suite.Require().NoError(reg.AddAlias("test", test.Account{}), "creating test alias")
	suite.Require().NoError(reg.Register(&test.Stock{}))
	suite.Require().NoError(reg.Register(&test.Federal{}))
	suite.Require().NoError(reg.Register(&test.State{}))
	suite.Require().NoError(reg.AddAlias("yamlTest", Account{}), "creating test alias")
	suite.Require().NoError(reg.Register(&Account{}))
	suite.Require().NoError(reg.Register(&Bond{}))
}

func TestYamlSuite(t *testing.T) {
	suite.Run(t, new(YamlTestSuite))
}

//////////////////////////////////////////////////////////////////////////

// TestMarshalCycle verifies the JSON Marshal/Unmarshal works as expected.
func (suite *YamlTestSuite) TestMarshalCycle() {
	account := MakeAccount()

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
	suite.Assert().Contains(string(marshaled), "[yamlTest]Bond")

	var newAccount Account
	suite.Require().NoError(yaml.Unmarshal(marshaled, &newAccount))
	if suite.showAccount {
		fmt.Println("---------------------------")
		spew.Dump(newAccount)
	}

	suite.Assert().NotEqual(account, newAccount)
	account.Favorite.ClearPrivateFields()
	for _, position := range account.Positions {
		position.ClearPrivateFields()
	}
	for _, position := range account.Lookup {
		position.ClearPrivateFields()
	}
	// Succeeds now that unexported (private) fields are gone.
	suite.Assert().Equal(account, &newAccount)
}

//////////////////////////////////////////////////////////////////////////

type Account struct {
	test.Account
}

func MakeAccount() *Account {
	account := &Account{}
	tBill := &Bond{}
	tBill.ConfigureTBill()
	state := &Bond{}
	state.ConfigureStateBond()
	account.MakeFake(test.MakeCostco(), test.MakeWalmart(), tBill, state)
	return account
}

type xferAccount struct {
	Account struct {
		Favorite  *wrapper[test.Investment]
		Positions []*wrapper[test.Investment]
		Lookup    map[string]*wrapper[test.Investment]
	}
	test.AccountData
}

func (a *Account) MarshalYAML() (interface{}, error) {
	xfer := &xferAccount{
		AccountData: a.AccountData,
	}

	// Pack objects referenced by interface fields.
	if a.Favorite != nil {
		xfer.Account.Favorite = Wrap[test.Investment](a.Favorite)
		if err := xfer.Account.Favorite.Pack(); err != nil {
			return nil, fmt.Errorf("wrap favorite: %w", err)
		}
	}
	if a.Positions != nil {
		fixed := make([]*wrapper[test.Investment], len(a.Positions))
		for i, pos := range a.Positions {
			fixed[i] = Wrap[test.Investment](pos)
			if err := fixed[i].Pack(); err != nil {
				return nil, fmt.Errorf("wrap Positions item: %w", err)
			}
		}
		xfer.Account.Positions = fixed
	}
	if a.Lookup != nil {
		fixed := make(map[string]*wrapper[test.Investment], len(a.Lookup))
		for k, pos := range a.Lookup {
			fixed[k] = Wrap[test.Investment](pos)
			if err := fixed[k].Pack(); err != nil {
				return nil, fmt.Errorf("wrap Lookup item: %w", err)
			}
		}
		xfer.Account.Lookup = fixed
	}

	return xfer, nil
}

func (a *Account) UnmarshalYAML(node *yaml.Node) error {
	xfer := &xferAccount{}
	if err := node.Decode(&xfer); err != nil {
		return fmt.Errorf("unmarshal to transfer account: %w", err)
	}

	a.AccountData = xfer.AccountData

	if err := xfer.Account.Favorite.Unpack(); err != nil {
		return fmt.Errorf("unwrap account favorite: %w", err)
	} else {
		a.Favorite = xfer.Account.Favorite.Get()
	}

	if xfer.Account.Positions != nil {
		fixed := make([]test.Investment, len(xfer.Account.Positions))
		for i, wPos := range xfer.Account.Positions {
			if err := wPos.Unpack(); err != nil {
				return fmt.Errorf("get Investment from Positions: %w", err)
			} else {
				fixed[i] = wPos.Get()
			}
		}
		a.Positions = fixed
	}

	if xfer.Account.Lookup != nil {
		fixed := make(map[string]test.Investment, len(xfer.Account.Lookup))
		for key, wPos := range xfer.Account.Lookup {
			if err := wPos.Unpack(); err != nil {
				return fmt.Errorf("get Investment from Lookup: %w", err)
			} else {
				fixed[key] = wPos.Get()
			}
		}
		a.Lookup = fixed
	}

	return nil
}

//////////////////////////////////////////////////////////////////////////

type Bond struct {
	test.Bond
}

type xferBond struct {
	Source   *wrapper[test.Borrower]
	BondData test.BondData
}

func (b *Bond) MarshalYAML() (interface{}, error) {
	xfer := &xferBond{BondData: b.Data}

	// Pack objects referenced by interface fields.
	if b.Source != nil {
		xfer.Source = Wrap[test.Borrower](b.Source)
		if err := xfer.Source.Pack(); err != nil {
			return nil, fmt.Errorf("pack borrower: %w", err)
		}
	}

	return xfer, nil
}

func (b *Bond) UnmarshalYAML(node *yaml.Node) error {
	xfer := &xferBond{}
	if err := node.Decode(&xfer); err != nil {
		return fmt.Errorf("unmarshal to transfer bond: %w", err)
	}

	b.Data = xfer.BondData

	if err := xfer.Source.Unpack(); err != nil {
		return fmt.Errorf("unpack borrower: %w", err)
	} else {
		b.Source = xfer.Source.Get()
	}

	return nil
}
