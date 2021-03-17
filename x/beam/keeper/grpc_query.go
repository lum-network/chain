package keeper

import (
	"github.com/lum-network/chain/x/beam/types"
)

var _ types.QueryServer = Keeper{}
