package wallet

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"golang.org/x/crypto/ed25519"
)

// ensures a set of currencies are valid
//   - value and decimal is positive
//   - decimal == decimal % 100
func (s Currencies) isValid() bool {
	for _, curr := range s {
		if !curr.isPositive() {
			return false
		}
		if curr.Decimal != curr.Decimal%100 {
			return false
		}
	}

	return true
}

func (s Currencies) Subtract(s2 Currencies) Currencies {
	out := Currencies{}

	// iterate over keys in 1st
	for ident, sAmount := range s {
		if s2Amount, ok := s2[ident]; ok {
			out[ident] = sAmount.Subtract(s2Amount)
		} else {
			out[ident] = sAmount
		}
	}

	// iterate over missed keys in 2nd
	for ident, s2Amount := range s2 {
		if _, ok := out[ident]; !ok {
			out[ident] = s2Amount.Inverse()
		}
	}

	return out
}

func (s Currencies) Inverse() Currencies {
	out := Currencies{}

	for ident, sAmount := range s {
		out[ident] = sAmount.Inverse()
	}

	return out
}

func (d DecimalAmount) Subtract(d2 DecimalAmount) DecimalAmount {
	newDecimal := (d.Decimal - d2.Decimal) % 100
	newValue := d.Value - d2.Value
	// subtract one if the decimal rolled over
	if newDecimal > d.Decimal {
		newValue -= 1
	}
	return DecimalAmount{Decimal: newDecimal, Value: newValue}
}

func NewTransaction(src ed25519.PublicKey, dest ed25519.PublicKey, curr Currency, amount DecimalAmount, create bool) transaction {
	return transaction{src, dest, [64]byte{}, curr, amount, create}
}

// returns the inverse under addition
func (d DecimalAmount) Inverse() DecimalAmount {
	return DecimalAmount{Value: -d.Value, Decimal: -d.Decimal}
}

func (d DecimalAmount) isPositive() bool {
	return d.Value >= 0 && d.Decimal >= 0
}

func (t transaction) ToBytes() []byte {
	type sign struct {
		Src    ed25519.PublicKey
		Dest   ed25519.PublicKey
		Curr   Currency
		Amount DecimalAmount
	}

	var buf bytes.Buffer
	if err := gob.NewEncoder(&buf).Encode(sign{t.Src, t.Dest, t.Curr, t.Amount}); err != nil {
		panic(err)
	}

	return buf.Bytes()
}

// validates a request to pay somebody
func (t transaction) IsVerified() bool {
	return t.Create || ed25519.Verify(t.Src, t.ToBytes(), t.Sig[:])
}

func (t *transaction) Sign(key ed25519.PrivateKey) {
	if !t.Create {
		copy(t.Sig[:], ed25519.Sign(key, t.ToBytes()))
	}
}

// returns true if the wallet belonging to key
// has more of the given decimalAmount than requestedAmount
func (s *WalletStore) CheckBalance(key ed25519.PublicKey, curr Currency, amount DecimalAmount) error {
	wallet, ok := s.Lookup(key)
	if !ok {
		return fmt.Errorf("wallet does not exist")
	}
	if !wallet[curr].Subtract(DecimalAmount(amount)).isPositive() {
		return fmt.Errorf("not enough cash: need %v, only have %v, would end up with %v", amount, wallet[curr], wallet[curr].Subtract(DecimalAmount(amount)))
	}

	return nil
}
