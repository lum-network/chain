package keeper

import (
	"github.com/sandblockio/chain/x/chain/types"
)

var _ types.QueryServer = Keeper{}
