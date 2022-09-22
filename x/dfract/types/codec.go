package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
)

// RegisterCodec Register the codec for  the message passing
func RegisterCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(&WithdrawAndMintProposal{}, "lum-network/WithdrawAndMintProposal", nil)
}

// RegisterInterfaces Register the implementations for the given codecs
func RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	registry.RegisterImplementations((*govtypes.Content)(nil), &WithdrawAndMintProposal{})
}

var (
	amino     = codec.NewLegacyAmino()
	ModuleCdc = codec.NewAminoCodec(amino)
)
