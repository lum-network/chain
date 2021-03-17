package keeper

import (
	"github.com/lum-network/chain/x/chain/types"
)

var _ types.QueryServer = Keeper{}
