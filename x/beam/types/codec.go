package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// RegisterCodec Register the codec for  the message passing
func RegisterCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(&MsgOpenBeam{}, "lum-network/OpenBeam", nil)
	cdc.RegisterConcrete(&MsgUpdateBeam{}, "lum-network/UpdateBeam", nil)
	cdc.RegisterConcrete(&MsgClaimBeam{}, "lum-network/ClaimBeam", nil)
	cdc.RegisterConcrete(&MsgCancelBeam{}, "lum-network/CancelBeam", nil)
	cdc.RegisterConcrete(&MsgCloseBeam{}, "lum-network/CloseBeam", nil)
}

// RegisterInterfaces Register the implementations for the given codecs
func RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	registry.RegisterImplementations((*sdk.Msg)(nil), &MsgOpenBeam{}, &MsgUpdateBeam{}, &MsgClaimBeam{}, &MsgCancelBeam{}, &MsgCloseBeam{})
}

var (
	amino     = codec.NewLegacyAmino()
	ModuleCdc = codec.NewAminoCodec(amino)
)
