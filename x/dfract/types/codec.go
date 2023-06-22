package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/legacy"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
)

var (
	amino     = codec.NewLegacyAmino()
	ModuleCdc = codec.NewAminoCodec(amino)
)

// RegisterLegacyAminoCodec Register the codec for the message passing
func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	// Messages
	cdc.RegisterConcrete(&MsgDeposit{}, "lum-network/dfract/MsgDeposit", nil)
	cdc.RegisterConcrete(&MsgWithdrawAndMint{}, "lum-network/dfract/MsgWithdrawAndMint", nil)
	// Proposals
	cdc.RegisterConcrete(&ProposalUpdateParams{}, "lum-network/dfract/ProposalUpdateParams", nil)
}

// RegisterInterfaces Register the implementations for the given codecs
func RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	// Messages
	registry.RegisterImplementations((*sdk.Msg)(nil), &MsgDeposit{})
	registry.RegisterImplementations((*sdk.Msg)(nil), &MsgWithdrawAndMint{})
	// Proposals
	registry.RegisterImplementations((*govtypes.Content)(nil), &ProposalUpdateParams{})

	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}

func init() {
	RegisterLegacyAminoCodec(amino)
	cryptocodec.RegisterCrypto(amino)
	amino.Seal()

	// Specific register call for Authz
	RegisterLegacyAminoCodec(legacy.Cdc)
}
