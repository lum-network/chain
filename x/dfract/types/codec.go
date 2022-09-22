package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
)

var (
	amino     = codec.NewLegacyAmino()
	ModuleCdc = codec.NewProtoCodec(cdctypes.NewInterfaceRegistry())
)

// RegisterCodec Register the codec for  the message passing
func RegisterCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(&MsgDeposit{}, "lum-network/Deposit", nil)
	cdc.RegisterConcrete(&WithdrawAndMintProposal{}, "lum-network/WithdrawAndMintProposal", nil)
}

// RegisterInterfaces Register the implementations for the given codecs
func RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	registry.RegisterImplementations((*sdk.Msg)(nil), &MsgDeposit{})
	registry.RegisterImplementations((*govtypes.Content)(nil), &WithdrawAndMintProposal{})

	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}
