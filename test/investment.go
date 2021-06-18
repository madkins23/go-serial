package test

import (
	"time"
)

var _ Investment = &Stock{}
var _ Investment = &Bond{}

//////////////////////////////////////////////////////////////////////////

type Account struct {
	Favorite  Investment
	Positions []Investment
	Lookup    map[string]Investment
}

//////////////////////////////////////////////////////////////////////////

func (a *Account) MakeFake() {
	a.MakeFakeUsing(MakeCostco(), MakeWalmart(), MakeTBill())
}

func (a *Account) MakeFakeUsing(costco, walmart *Stock, tBill *Bond) {
	a.Favorite = costco
	a.Positions = []Investment{
		costco,
		walmart,
		tBill,
	}
	a.Lookup = map[string]Investment{
		"COST": costco,
		"WMT":  walmart,
	}
}

//////////////////////////////////////////////////////////////////////////

type Investment interface {
	CurrentValue() (float32, error)
	ClearPrivateFields()
}

//////////////////////////////////////////////////////////////////////////

type Stock struct {
	Market   string
	Symbol   string
	Name     string
	Position float32
	Value    float32
	notes    string
}

//////////////////////////////////////////////////////////////////////////

func (s *Stock) CurrentValue() (float32, error) {
	return s.Position * s.Value, nil
}

func (s *Stock) ClearPrivateFields() {
	s.notes = ""
}

func (s *Stock) ConfigureCostco() *Stock {
	s.Market = "NASDAQ"
	s.Symbol = "COST"
	s.Name = "Costco"
	s.Position = 10
	s.Value = 400
	s.notes = "Lorem ipsum dolor sit amet"
	return s
}

func MakeCostco() *Stock {
	return (&Stock{}).ConfigureCostco()
}

func (s *Stock) ConfigureWalmart() *Stock {
	s.Market = "NYSE"
	s.Symbol = "WMT"
	s.Name = "Walmart"
	s.Position = 20
	s.Value = 150
	s.notes = "consectetur adipiscing elit"
	return s
}

func MakeWalmart() *Stock {
	return (&Stock{}).ConfigureWalmart()
}

//////////////////////////////////////////////////////////////////////////

type Bond struct {
	Source   string
	Name     string
	Value    float32
	Interest float32
	Duration time.Duration
	notes    string
}

func (b *Bond) CurrentValue() (float32, error) {
	return b.Value, nil
}

func (b *Bond) ClearPrivateFields() {
	b.notes = ""
}

func (b *Bond) ConfigureTBill() *Bond {
	b.Source = "Treasury"
	b.Name = "T-Bill"
	b.Value = 1000
	b.Interest = 0.75
	b.Duration = 365 * 24 * time.Hour
	b.notes = "sed do eiusmod tempor incididunt ut labore et dolore magna aliqua"
	return b
}

func MakeTBill() *Bond {
	return (&Bond{}).ConfigureTBill()
}
