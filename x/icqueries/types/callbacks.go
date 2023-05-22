package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type QueryResponseStatus int

const (
	QueryResponseStatus_SUCCESS QueryResponseStatus = iota
	QueryResponseStatus_TIMEOUT
	QueryResponseStatus_FAILURE
)

type QueryCallbacks interface {
	AddICQCallback(id string, fn interface{}) QueryCallbacks
	RegisterICQCallbacks() QueryCallbacks
	CallICQCallback(ctx sdk.Context, id string, args []byte, query Query, status QueryResponseStatus) error
	HasICQCallback(id string) bool
}
