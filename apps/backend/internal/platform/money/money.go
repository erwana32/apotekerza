// Package money holds decimal wrappers for money and item quantities.
package money

import (
	"encoding/json"
	"fmt"

	"github.com/shopspring/decimal"
)

const (
	// MoneyScale is the number of decimal places used for money (numeric(18,2)).
	MoneyScale int32 = 2
	// QuantityScale is the number of decimal places used for quantities (numeric(18,4)).
	QuantityScale int32 = 4
)

// Money is an amount of money. It never uses floating point.
type Money struct{ decimal.Decimal }

// Quantity is an item quantity in the product's smallest unit.
type Quantity struct{ decimal.Decimal }

// NewMoney parses a decimal string into Money.
func NewMoney(value string) (Money, error) {
	amount, err := decimal.NewFromString(value)
	if err != nil {
		return Money{}, fmt.Errorf("money: parse %q: %w", value, err)
	}
	return Money{amount}, nil
}

// NewQuantity parses a decimal string into Quantity.
func NewQuantity(value string) (Quantity, error) {
	amount, err := decimal.NewFromString(value)
	if err != nil {
		return Quantity{}, fmt.Errorf("money: parse quantity %q: %w", value, err)
	}
	return Quantity{amount}, nil
}

// RoundMoney rounds to two decimals, half away from zero. Callers choose where
// this happens; finance documents the exact rounding points when implemented.
func RoundMoney(value decimal.Decimal) Money {
	return Money{value.Round(MoneyScale)}
}

// RoundQuantity rounds to four decimals, half away from zero.
func RoundQuantity(value decimal.Decimal) Quantity {
	return Quantity{value.Round(QuantityScale)}
}

func (m Money) MarshalJSON() ([]byte, error) {
	return []byte(`"` + m.StringFixed(MoneyScale) + `"`), nil
}

func (q Quantity) MarshalJSON() ([]byte, error) {
	return []byte(`"` + q.StringFixed(QuantityScale) + `"`), nil
}

func (m *Money) UnmarshalJSON(data []byte) error {
	amount, err := parseDecimalString(data)
	if err != nil {
		return err
	}
	*m = Money{amount}
	return nil
}

func (q *Quantity) UnmarshalJSON(data []byte) error {
	amount, err := parseDecimalString(data)
	if err != nil {
		return err
	}
	*q = Quantity{amount}
	return nil
}

// parseDecimalString accepts only a quoted decimal string. JSON numbers are
// rejected so precision cannot be lost before parsing.
func parseDecimalString(data []byte) (decimal.Decimal, error) {
	if len(data) < 2 || data[0] != '"' || data[len(data)-1] != '"' {
		return decimal.Decimal{}, fmt.Errorf("money: expected a decimal string, got %s", data)
	}
	var value string
	if err := json.Unmarshal(data, &value); err != nil {
		return decimal.Decimal{}, fmt.Errorf("money: decode decimal string: %w", err)
	}
	amount, err := decimal.NewFromString(value)
	if err != nil {
		return decimal.Decimal{}, fmt.Errorf("money: parse %q: %w", value, err)
	}
	return amount, nil
}
