package wallet

import (
	"reflect"
	"testing"
)

func TestCurrencies_Subtract(t *testing.T) {
	type args struct {
		s2 Account
	}
	tests := []struct {
		name string
		s    Account
		args args
		want Account
	}{
		{
			"Test",
			Account{"usd": {Value: 10, Decimal: 0}},
			args{s2: Account{"usd": {Value: 2, Decimal: 0}}},
			Account{"usd": {Value: 8, Decimal: 0}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.s.Subtract(tt.args.s2); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Account.Subtract() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDecimalAmount_Subtract(t *testing.T) {
	type fields struct {
		Value   int64
		Decimal int8
	}
	type args struct {
		d2 DecimalAmount
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   DecimalAmount
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := DecimalAmount{
				Value:   tt.fields.Value,
				Decimal: tt.fields.Decimal,
			}
			if got := d.Subtract(tt.args.d2); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("DecimalAmount.Subtract() = %v, want %v", got, tt.want)
			}
		})
	}
}
