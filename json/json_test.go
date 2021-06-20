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

	"github.com/madkins23/go-serial/test"
	"github.com/madkins23/go-type/reg"
)

type JsonTestSuite struct {
	suite.Suite
	showAccount bool
}

func (suite *JsonTestSuite) SetupSuite() {
	if showAccount, found := os.LookupEnv("GO-TYPE-SHOW-ACCOUNT"); found {
		var err error
		suite.showAccount, err = strconv.ParseBool(showAccount)
		suite.Require().NoError(err)
	}
	suite.showAccount = true
	suite.Require().NoError(reg.AddAlias("test", test.Account{}), "creating test alias")
	suite.Require().NoError(reg.Register(&test.Stock{}))
	suite.Require().NoError(reg.Register(&test.Bond{}))
	suite.Require().NoError(reg.AddAlias("jsonTest", Account{}), "creating test alias")
	suite.Require().NoError(reg.Register(&Account{}))
}

func TestJsonSuite(t *testing.T) {
	suite.Run(t, new(JsonTestSuite))
}

//////////////////////////////////////////////////////////////////////////

// TestMarshalCycle verifies the JSON Marshal/Unmarshal works as expected.
func (suite *JsonTestSuite) TestMarshalCycle() {
	account := MakeAccount()

	marshaled, err := json.Marshal(account)
	suite.Require().NoError(err)
	if suite.showAccount {
		var buf bytes.Buffer
		suite.Require().NoError(json.Indent(&buf, marshaled, "", "  "))
		fmt.Println(buf.String())
	}
	suite.Assert().Contains(string(marshaled), "TypeName")
	suite.Assert().Contains(string(marshaled), "[test]Stock")
	suite.Assert().Contains(string(marshaled), "[test]Bond")

	var newAccount Account
	suite.Require().NoError(json.Unmarshal(marshaled, &newAccount))
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
	account.MakeFake()
	return account
}

type xferAccount struct {
	AccountData test.AccountData
	Account     struct {
		Favorite  *Wrapper
		Positions []*Wrapper
		Lookup    map[string]*Wrapper
	}
}

func (a *Account) MarshalJSON() ([]byte, error) {
	xfer := &xferAccount{AccountData: a.AccountData}

	// Wrap objects referenced by interface fields.
	var err error
	if a.Favorite != nil {
		if xfer.Account.Favorite, err = WrapItem(a.Favorite); err != nil {
			return nil, fmt.Errorf("wrap favorite: %w", err)
		}
	}
	if a.Positions != nil {
		fixed := make([]*Wrapper, len(a.Positions))
		for i, pos := range a.Positions {
			if fixed[i], err = WrapItem(pos); err != nil {
				return nil, fmt.Errorf("wrap Positions item: %w", err)
			}
		}
		xfer.Account.Positions = fixed
	}
	if a.Lookup != nil {
		fixed := make(map[string]*Wrapper, len(a.Lookup))
		for k, pos := range a.Lookup {
			if fixed[k], err = WrapItem(pos); err != nil {
				return nil, fmt.Errorf("wrap Lookup item: %w", err)
			}
		}
		xfer.Account.Lookup = fixed
	}

	return json.Marshal(xfer)
}

func (a *Account) UnmarshalJSON(marshaled []byte) error {
	xfer := &xferAccount{}
	if err := json.Unmarshal(marshaled, xfer); err != nil {
		return fmt.Errorf("unmarshal to loader: %w", err)
	}

	a.AccountData = xfer.AccountData

	// Unwrap objects referenced by interface fields.
	var err error
	if a.Favorite, err = a.getInvestment(xfer.Account.Favorite); err != nil {
		return fmt.Errorf("get Investment from Favorite")
	}
	if xfer.Account.Positions != nil {
		fixed := make([]test.Investment, len(xfer.Account.Positions))
		for i, wPos := range xfer.Account.Positions {
			if fixed[i], err = a.getInvestment(wPos); err != nil {
				return fmt.Errorf("get Investment from Positions")
			}
		}
		a.Positions = fixed
	}
	if xfer.Account.Lookup != nil {
		fixed := make(map[string]test.Investment, len(xfer.Account.Lookup))
		for key, wPos := range xfer.Account.Lookup {
			if fixed[key], err = a.getInvestment(wPos); err != nil {
				return fmt.Errorf("get Investment from Lookup")
			}
		}
		a.Lookup = fixed
	}

	return nil
}

func (a *Account) getInvestment(w *Wrapper) (test.Investment, error) {
	var ok bool
	var investment test.Investment
	if w != nil {
		if item, err := w.Unwrap(); err != nil {
			return nil, fmt.Errorf("unwrap item: %w", err)
		} else if investment, ok = item.(test.Investment); !ok {
			return nil, fmt.Errorf("item %#v not Investment", item)
		}
	}

	return investment, nil
}
