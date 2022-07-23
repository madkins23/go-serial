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
	showAccount bool
}

func (suite *JsonTestSuite) SetupSuite() {
	if showAccount, found := os.LookupEnv("GO-TYPE-SHOW-ACCOUNT"); found {
		var err error
		suite.showAccount, err = strconv.ParseBool(showAccount)
		suite.Require().NoError(err)
	}
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
//////////////////////////////////////////////////////////////////////////
// This section tests the origial "manual" mechanism.
// This requires more coding and extra structs.

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
	suite.Assert().Contains(string(marshaled), "type\":")
	suite.Assert().Contains(string(marshaled), "data\":")
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
		Favorite  *wrapper[test.Investment]
		Positions []*wrapper[test.Investment]
		Lookup    map[string]*wrapper[test.Investment]
	}
}

func (a *Account) MarshalJSON() ([]byte, error) {
	xfer := &xferAccount{AccountData: a.AccountData}

	// Wrap objects referenced by interface fields.
	if a.Favorite != nil {
		xfer.Account.Favorite = Wrap[test.Investment](a.Favorite)
		if err := xfer.Account.Favorite.Wrap(); err != nil {
			return nil, fmt.Errorf("wrap favorite: %w", err)
		}
	}
	if a.Positions != nil {
		fixed := make([]*wrapper[test.Investment], len(a.Positions))
		for i, pos := range a.Positions {
			fixed[i] = Wrap[test.Investment](pos)
			if err := fixed[i].Wrap(); err != nil {
				return nil, fmt.Errorf("wrap Positions item: %w", err)
			}
		}
		xfer.Account.Positions = fixed
	}
	if a.Lookup != nil {
		fixed := make(map[string]*wrapper[test.Investment], len(a.Lookup))
		for k, pos := range a.Lookup {
			fixed[k] = Wrap[test.Investment](pos)
			if err := fixed[k].Wrap(); err != nil {
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
		return fmt.Errorf("unmarshal to transfer account: %w", err)
	}

	a.AccountData = xfer.AccountData

	if err := xfer.Account.Favorite.Unwrap(); err != nil {
		return fmt.Errorf("unwrap account favorite: %w", err)
	} else {
		a.Favorite = xfer.Account.Favorite.Get()
	}

	if xfer.Account.Positions != nil {
		fixed := make([]test.Investment, len(xfer.Account.Positions))
		for i, wPos := range xfer.Account.Positions {
			if err := wPos.Unwrap(); err != nil {
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
			if err := wPos.Unwrap(); err != nil {
				return fmt.Errorf("get Investment from Lookup: %w", err)
			} else {
				fixed[key] = wPos.Get()
			}
		}
		a.Lookup = fixed
	}

	return nil
}

func (a *Account) getInvestment(w *wrapper[test.Investment]) (test.Investment, error) {
	var ok bool
	var investment test.Investment
	if w != nil {
		if err := w.Unwrap(); err != nil {
			return nil, fmt.Errorf("unwrap item: %w", err)
		} else if investment, ok = w.Get().(test.Investment); !ok {
			return nil, fmt.Errorf("item %#v not Investment", w.Get())
		}
	}

	return investment, nil
}

//////////////////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////////
// This section tests the newer "proxy" mechanism.
// This requires all interface objects to be wrapped in proxy.Wrapper.

// TestMarshalCycle verifies the JSON Marshal/Unmarshal works as expected objects.
func (suite *JsonTestSuite) TestProxyMarshalCycle() {
	account := MakeAccount()

	marshaled, err := json.Marshal(account)
	suite.Require().NoError(err)
	if suite.showAccount {
		var buf bytes.Buffer
		suite.Require().NoError(json.Indent(&buf, marshaled, "", "  "))
		fmt.Println(buf.String())
	}
	suite.Assert().Contains(string(marshaled), "type\":")
	suite.Assert().Contains(string(marshaled), "data\":")
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
