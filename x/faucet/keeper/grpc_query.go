package keeper

import (
	"github.com/sandblockio/chain/x/faucet/types"
)

var _ types.QueryServer = Keeper{}
