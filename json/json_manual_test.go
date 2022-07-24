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

// The (original) "manual" mechanism for JSON.
// The manual way requires more coding and extra structs during de/serialization.
// On the other hand, the objects are directly accessible without constant dereferences.
type JsonManualTestSuite struct {
	suite.Suite
	showAccount bool
}

func (suite *JsonManualTestSuite) SetupSuite() {
	if showAccount, found := os.LookupEnv("GO-TYPE-SHOW-ACCOUNT"); found {
		var err error
		suite.showAccount, err = strconv.ParseBool(showAccount)
		suite.Require().NoError(err)
	}
	suite.Require().NoError(test.Registration())
	suite.Require().NoError(reg.AddAlias("jsonManualTest", ManualAccount{}),
		"creating json manual test alias")
	suite.Require().NoError(reg.Register(&ManualAccount{}))
	suite.Require().NoError(reg.Register(&ManualBond{}))
}

func TestManualJsonSuite(t *testing.T) {
	suite.Run(t, new(JsonManualTestSuite))
}

//////////////////////////////////////////////////////////////////////////

// TestMarshalCycle verifies the JSON Marshal/Unmarshal works as expected.
func (suite *JsonManualTestSuite) TestMarshalCycle() {
	account := MakeManualAccount()

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
	suite.Assert().Contains(string(marshaled), "[test]Federal")
	suite.Assert().Contains(string(marshaled), "[test]State")
	suite.Assert().Contains(string(marshaled), "[jsonManualTest]ManualBond")

	var newAccount ManualAccount
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

type ManualAccount struct {
	test.Account
}

func MakeManualAccount() *ManualAccount {
	account := &ManualAccount{}
	tBill := &ManualBond{}
	tBill.ConfigureTBill()
	state := &ManualBond{}
	state.ConfigureStateBond()
	account.MakeFake(test.MakeCostco(), test.MakeWalmart(), tBill, state)
	return account
}

type xferManualAccount struct {
	AccountData test.AccountData
	Account     struct {
		Favorite  *Wrapper[test.Investment]
		Positions []*Wrapper[test.Investment]
		Lookup    map[string]*Wrapper[test.Investment]
	}
}

func (a *ManualAccount) MarshalJSON() ([]byte, error) {
	xfer := &xferManualAccount{AccountData: a.AccountData}

	// Pack objects referenced by interface fields.
	if a.Favorite != nil {
		xfer.Account.Favorite = Wrap[test.Investment](a.Favorite)
	}
	if a.Positions != nil {
		fixed := make([]*Wrapper[test.Investment], len(a.Positions))
		for i, pos := range a.Positions {
			fixed[i] = Wrap[test.Investment](pos)
		}
		xfer.Account.Positions = fixed
	}
	if a.Lookup != nil {
		fixed := make(map[string]*Wrapper[test.Investment], len(a.Lookup))
		for k, pos := range a.Lookup {
			fixed[k] = Wrap[test.Investment](pos)
		}
		xfer.Account.Lookup = fixed
	}

	return json.Marshal(xfer)
}

func (a *ManualAccount) UnmarshalJSON(marshaled []byte) error {
	xfer := &xferManualAccount{}
	if err := json.Unmarshal(marshaled, xfer); err != nil {
		return fmt.Errorf("unmarshal to transfer account: %w", err)
	}

	a.AccountData = xfer.AccountData
	a.Favorite = xfer.Account.Favorite.Get()

	if xfer.Account.Positions != nil {
		fixed := make([]test.Investment, len(xfer.Account.Positions))
		for i, wPos := range xfer.Account.Positions {
			fixed[i] = wPos.Get()
		}
		a.Positions = fixed
	}

	if xfer.Account.Lookup != nil {
		fixed := make(map[string]test.Investment, len(xfer.Account.Lookup))
		for key, wPos := range xfer.Account.Lookup {
			fixed[key] = wPos.Get()
		}
		a.Lookup = fixed
	}

	return nil
}

//////////////////////////////////////////////////////////////////////////

type ManualBond struct {
	test.Bond
}

type xferManualBond struct {
	Source   *Wrapper[test.Borrower]
	BondData test.BondData
}

func (b *ManualBond) MarshalJSON() ([]byte, error) {
	xfer := &xferManualBond{BondData: b.Data}

	// Pack objects referenced by interface fields.
	if b.Source != nil {
		xfer.Source = Wrap[test.Borrower](b.Source)
	}

	return json.Marshal(xfer)
}

func (b *ManualBond) UnmarshalJSON(marshaled []byte) error {
	xfer := &xferManualBond{}
	if err := json.Unmarshal(marshaled, xfer); err != nil {
		return fmt.Errorf("unmarshal to transfer account: %w", err)
	}

	b.Data = xfer.BondData
	b.Source = xfer.Source.Get()

	return nil
}
