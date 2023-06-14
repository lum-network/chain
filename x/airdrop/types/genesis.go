package types

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type Actions []Action

// DefaultIndex is the default capability global index.
const DefaultIndex uint64 = 1

// DefaultGenesis returns the default Capability genesis state.
func DefaultGenesis() *GenesisState {
	return &GenesisState{
		ModuleAccountBalance: sdk.NewCoin(DefaultClaimDenom, sdk.ZeroInt()),
		Params: Params{
			AirdropStartTime:   time.Time{},
			DurationUntilDecay: DefaultDurationUntilDecay, // 2 month
			DurationOfDecay:    DefaultDurationOfDecay,    // 4 months
			ClaimDenom:         DefaultClaimDenom,         // ulum
		},
		ClaimRecords: []ClaimRecord{},
	}
}

// GetGenesisStateFromAppState returns x/airdrop GenesisState given raw application genesis state.
func GetGenesisStateFromAppState(cdc codec.JSONCodec, appState map[string]json.RawMessage) *GenesisState {
	var genesisState GenesisState

	if appState[ModuleName] != nil {
		cdc.MustUnmarshalJSON(appState[ModuleName], &genesisState)
	}

	return &genesisState
}

// Validate performs basic genesis state validation returning an error upon any failure.
func (gs GenesisState) Validate() error {
	if err := gs.ModuleAccountBalance.Validate(); err != nil {
		return err
	}

	totalClaimable := sdk.NewCoin(gs.Params.ClaimDenom, sdk.ZeroInt())
	for _, claimRecord := range gs.ClaimRecords {
		for _, claimable := range claimRecord.InitialClaimableAmount {
			if err := claimable.Validate(); err != nil {
				return fmt.Errorf(fmt.Sprintf("Claimable invalid for address %s", claimRecord.Address))
			}

			if claimable.Denom != gs.Params.ClaimDenom {
				return fmt.Errorf(fmt.Sprintf("Tried to commit invalid denom %s", claimable.Denom))
			}

			totalClaimable = totalClaimable.Add(claimable)
		}
	}

	if gs.Params.ClaimDenom != gs.ModuleAccountBalance.Denom {
		return fmt.Errorf(fmt.Sprintf("Denom for module and claim does not match %s != %s", gs.Params.ClaimDenom, gs.ModuleAccountBalance.Denom))
	}

	if !totalClaimable.IsEqual(gs.ModuleAccountBalance) {
		return fmt.Errorf(fmt.Sprintf("Balance is different from total of claimable: %s != %s", totalClaimable.String(), gs.ModuleAccountBalance.String()))
	}

	return nil
}
