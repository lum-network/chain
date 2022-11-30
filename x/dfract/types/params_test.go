package types

import (
	"testing"
)

func TestParams_Validate(t *testing.T) {

	const (
		DefaultDenom            = "ibc/05554A9BFDD28894D7F18F4C707AA0930D778751A437A9FE1F4684A3E1199728"
		DefaultMinDepositAmount = 1000000
	)

	var (
		validDenomArr   = []string{DefaultDenom}
		invalidDenomArr = []string{"aevmos"}
	)

	type fields struct {
		DepositDenom     []string
		MinDepositAmount uint32
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			"Valid - take array of strings as deposit denom",
			fields{validDenomArr, DefaultMinDepositAmount},
			false,
		},
		{
			"Invalid - take invalid array of strings as deposit denom",
			fields{invalidDenomArr, DefaultMinDepositAmount},
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Params{
				DepositDenom:     tt.fields.DepositDenom,
				MinDepositAmount: tt.fields.MinDepositAmount,
			}
			if err := p.Validate(); (err != nil) != tt.wantErr {
				t.Errorf("Params.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
