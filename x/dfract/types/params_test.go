package types

import (
	"testing"
)

func TestParamsValidate(t *testing.T) {
	type fields struct {
		DepositDenoms    []string
		MinDepositAmount uint32
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			"Validate - Should take an array of strings as deposit denom and 1000000 deposit amount",
			fields{[]string{"ibc/05554A9BFDD28894D7F18F4C707AA0930D778751A437A9FE1F4684A3E1199728"}, 1000000},
			false,
		},
		{
			"Invalidate - Should take an array of strings as deposit denom and 0 deposit amount",
			fields{[]string{"ibc/05554A9BFDD28894D7F18F4C707AA0930D778751A437A9FE1F4684A3E1199728"}, 0},
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Params{
				DepositDenoms:    tt.fields.DepositDenoms,
				MinDepositAmount: tt.fields.MinDepositAmount,
			}
			if err := p.Validate(); (err != nil) != tt.wantErr {
				t.Errorf("Params.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
