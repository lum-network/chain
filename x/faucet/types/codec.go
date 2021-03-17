package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	// this line is used by starport scaffolding # 1
)

// RegisterCodec Register the GRPC codecs
func RegisterCodec(cdc *codec.LegacyAmino) {
}

// RegisterInterfaces Register the message interfaces for the previously registered codecs
func RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
}

var (
	amino = codec.NewLegacyAmino()
	ModuleCdc = codec.NewAminoCodec(amino)
)
