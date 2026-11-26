package money

import (
	"encoding/json"
	"testing"

	"github.com/shopspring/decimal"
)

func TestMoney_JSONRoundTrip(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"whole number", "12500", "12500.00"},
		{"one decimal", "0.1", "0.10"},
		{"negative", "-3.5", "-3.50"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			amount, err := NewMoney(tt.input)
			if err != nil {
				t.Fatal(err)
			}
			data, err := json.Marshal(amount)
			if err != nil {
				t.Fatal(err)
			}
			if got := string(data); got != `"`+tt.want+`"` {
				t.Fatalf("JSON = %s, want %q", got, tt.want)
			}
			var decoded Money
			if err := json.Unmarshal(data, &decoded); err != nil {
				t.Fatal(err)
			}
			if decoded.StringFixed(MoneyScale) != tt.want {
				t.Fatalf("round trip = %s, want %s", decoded.StringFixed(MoneyScale), tt.want)
			}
		})
	}
}

func TestMoney_RejectsJSONNumbers(t *testing.T) {
	for _, input := range []string{`12500.00`, `12500`, `1e3`, `null`} {
		var amount Money
		if err := json.Unmarshal([]byte(input), &amount); err == nil {
			t.Errorf("Unmarshal(%s) succeeded, want error", input)
		}
	}
}

func TestMoney_Rounding(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"half up", "10.005", "10.01"},
		{"below half", "10.004", "10.00"},
		{"negative half away from zero", "-10.005", "-10.01"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			value, err := decimal.NewFromString(tt.input)
			if err != nil {
				t.Fatal(err)
			}
			if got := RoundMoney(value).StringFixed(MoneyScale); got != tt.want {
				t.Fatalf("RoundMoney(%s) = %s, want %s", tt.input, got, tt.want)
			}
		})
	}
}

func TestQuantity_RoundAndJSON(t *testing.T) {
	value, err := NewQuantity("1.23445")
	if err != nil {
		t.Fatal(err)
	}
	rounded := RoundQuantity(value.Decimal)
	if got := rounded.StringFixed(QuantityScale); got != "1.2345" {
		t.Fatalf("RoundQuantity = %s, want 1.2345", got)
	}
	data, err := json.Marshal(rounded)
	if err != nil {
		t.Fatal(err)
	}
	if got := string(data); got != `"1.2345"` {
		t.Fatalf("JSON = %s, want \"1.2345\"", got)
	}
}
