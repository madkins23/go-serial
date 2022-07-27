package test

import (
	"fmt"
	"time"

	"github.com/madkins23/go-type/reg"
)

var _ Investment = &Stock{}
var _ Investment = &Bond{}
var _ Borrower = &Federal{}
var _ Borrower = &State{}

// Registration adds the 'test' alias and registers several structs.
// Uses the github.com/madkins23/go-type library to register structs by name.
func Registration() error {
	if err := reg.AddAlias("test", &Account{}); err != nil {
		return fmt.Errorf("adding 'test' alias: %w", err)
	}
	if err := reg.Register(&Stock{}); err != nil {
		return fmt.Errorf("registering Stock struct: %w", err)
	}
	if err := reg.Register(&Federal{}); err != nil {
		return fmt.Errorf("registering Stock struct: %w", err)
	}
	if err := reg.Register(&State{}); err != nil {
		return fmt.Errorf("registering Stock struct: %w", err)
	}
	return nil
}

//////////////////////////////////////////////////////////////////////////

type Account struct {
	AccountData
	Favorite  Investment
	Positions []Investment
	Lookup    map[string]Investment
}

type AccountData struct {
	Name    string
	Age     uint
	Veteran bool
}

//------------------------------------------------------------------------

// MakeFake creates and initializes a fake account using the specified investements.
// The first investment is assumed to be the favorite.
func (a *Account) MakeFake(investments ...Investment) {
	a.AccountData = AccountData{
		Name:    "Goober Snoofus",
		Age:     23,
		Veteran: true,
	}
	a.Favorite = investments[0]
	a.Positions = make([]Investment, len(investments))
	a.Lookup = make(map[string]Investment)
	for i, investment := range investments {
		a.Positions[i] = investment
		switch it := investment.(type) {
		case *Stock:
			a.Lookup[it.Symbol] = investment
		}
	}
}

//////////////////////////////////////////////////////////////////////////

type Investment interface {
	CurrentValue() (float32, error)
	ClearPrivateFields()
}

//========================================================================

type Stock struct {
	Market   string
	Symbol   string
	Name     string
	Position float32
	Value    float32
	notes    string
}

func (s *Stock) CurrentValue() (float32, error) {
	return s.Position * s.Value, nil
}

func (s *Stock) ClearPrivateFields() {
	s.notes = ""
}

func MakeCostco() *Stock {
	return &Stock{
		Market:   "NASDAQ",
		Symbol:   "COST",
		Name:     "Costco",
		Position: 10,
		Value:    4000,
		notes:    "Lorem ipsum dolor sit amet",
	}
}

func MakeWalmart() *Stock {
	return &Stock{
		Market:   "NYSE",
		Symbol:   "WMT",
		Name:     "Walmart",
		Position: 20,
		Value:    150,
		notes:    "consectetur adipiscing elit",
	}
}

//========================================================================

type Bond struct {
	Source Borrower
	Data   BondData
}

type BondData struct {
	Name     string
	Value    float32
	Interest float32
	Duration time.Duration
	notes    string
}

func (bd *BondData) ClearPrivateFields() {
	bd.notes = ""
}

func (b *Bond) CurrentValue() (float32, error) {
	return b.Data.Value, nil
}

func (b *Bond) ClearPrivateFields() {
	b.Data.ClearPrivateFields()
}

func (b *Bond) ConfigureTBill() {
	b.Source = TBillSource()
	b.Data = TBillBondData()
}

func TBillBondData() BondData {
	return BondData{
		Name:     "T-Bill",
		Value:    1000,
		Interest: 0.75,
		Duration: 365 * 24 * time.Hour,
		notes:    "sed do eiusmod tempor incididunt ut labore et dolore magna aliqua",
	}
}

func TBillSource() Borrower {
	return &Federal{Class: "T-Bill"}
}

func (b *Bond) ConfigureStateBond() {
	b.Source = StateBondSource()
	b.Data = StateBondData()
}

func StateBondData() BondData {
	return BondData{
		Name:     "Roads",
		Value:    1000,
		Interest: 1.75,
		Duration: 10 * 365 * 24 * time.Hour,
		notes:    "vero eos et accusamus et iusto odio dignissimos ducimus qui blanditiis praesentium voluptatum",
	}
}

func StateBondSource() Borrower {
	return &State{State: "Confusion"}
}

//////////////////////////////////////////////////////////////////////////

// Borrower is broken out to test nesting of interface objects.
// Borrower is nested within Bond within Account.
type Borrower interface {
	Name() string
}

//========================================================================

type Federal struct {
	Class string
}

func (f *Federal) Name() string {
	return "Treasury"
}

//========================================================================

type State struct {
	State string
}

func (c *State) Name() string {
	return c.State
}
