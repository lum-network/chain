package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// RegisterCodec Register the codec for  the message passing
func RegisterCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(&MsgOpenBeam{}, "sandblock/OpenBeam", nil)
	cdc.RegisterConcrete(&MsgIncreaseBeam{}, "sandblock/IncreaseBeam", nil)
	cdc.RegisterConcrete(&MsgClaimBeam{}, "sandblock/ClaimBeam", nil)
	cdc.RegisterConcrete(&MsgCancelBeam{}, "sandblock/CancelBeam", nil)
	cdc.RegisterConcrete(&MsgCloseBeam{}, "sandblock/CloseBeam", nil)
}

// RegisterInterfaces Register the implementations for the given codecs
func RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	registry.RegisterImplementations((*sdk.Msg)(nil), &MsgOpenBeam{}, &MsgIncreaseBeam{}, &MsgClaimBeam{}, &MsgCancelBeam{}, &MsgCloseBeam{})
}

var (
	amino     = codec.NewLegacyAmino()
	ModuleCdc = codec.NewAminoCodec(amino)
)
