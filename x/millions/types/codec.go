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
	cdc.RegisterConcrete(&MsgDeposit{}, "lum-network/millions/MsgDeposit", nil)
	cdc.RegisterConcrete(&MsgDepositRetry{}, "lum-network/millions/MsgDepositRetry", nil)
	cdc.RegisterConcrete(&MsgClaimPrize{}, "lum-network/millions/MsgClaimPrize", nil)
	cdc.RegisterConcrete(&MsgDrawRetry{}, "lum-network/millions/MsgDrawRetry", nil)
	cdc.RegisterConcrete(&MsgWithdrawDeposit{}, "lum-network/millions/MsgWithdrawDeposit", nil)
	cdc.RegisterConcrete(&MsgWithdrawDepositRetry{}, "lum-network/millions/MsgWithdrawDepositRetry", nil)
	cdc.RegisterConcrete(&MsgRestoreInterchainAccounts{}, "lum-network/millions/MsgRestoreInterchainAccounts", nil)
	cdc.RegisterConcrete(&MsgRedelegateRetry{}, "lum-network/millions/MsgRedelegateRetry", nil)

	// Proposals
	cdc.RegisterConcrete(&ProposalUpdateParams{}, "lum-network/millions/ProposalUpdateParams", nil)
	cdc.RegisterConcrete(&ProposalRegisterPool{}, "lum-network/millions/ProposalRegisterPool", nil)
	cdc.RegisterConcrete(&ProposalUpdatePool{}, "lum-network/millions/ProposalUpdatePool", nil)
	cdc.RegisterConcrete(&ProposalDisableValidator{}, "lum-network/millions/ProposalDisableValidator", nil)
}

// RegisterInterfaces Register the implementations for the given codecs
func RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	// Messages
	registry.RegisterImplementations((*sdk.Msg)(nil), &MsgDeposit{})
	registry.RegisterImplementations((*sdk.Msg)(nil), &MsgDepositRetry{})
	registry.RegisterImplementations((*sdk.Msg)(nil), &MsgClaimPrize{})
	registry.RegisterImplementations((*sdk.Msg)(nil), &MsgDrawRetry{})
	registry.RegisterImplementations((*sdk.Msg)(nil), &MsgWithdrawDeposit{})
	registry.RegisterImplementations((*sdk.Msg)(nil), &MsgWithdrawDepositRetry{})
	registry.RegisterImplementations((*sdk.Msg)(nil), &MsgRestoreInterchainAccounts{})
	registry.RegisterImplementations((*sdk.Msg)(nil), &MsgRedelegateRetry{})

	// Proposals
	registry.RegisterImplementations((*govtypes.Content)(nil), &ProposalUpdateParams{})
	registry.RegisterImplementations((*govtypes.Content)(nil), &ProposalRegisterPool{})
	registry.RegisterImplementations((*govtypes.Content)(nil), &ProposalUpdatePool{})
	registry.RegisterImplementations((*govtypes.Content)(nil), &ProposalDisableValidator{})

	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}

func init() {
	RegisterLegacyAminoCodec(amino)
	cryptocodec.RegisterCrypto(amino)
	amino.Seal()

	RegisterLegacyAminoCodec(legacy.Cdc)
}
