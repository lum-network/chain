package keeper

import (
	"github.com/lum-network/chain/x/faucet/types"
)

var _ types.QueryServer = Keeper{}
