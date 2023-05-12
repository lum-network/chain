package keeper

import (
	"github.com/lum-network/chain/x/icacallbacks/types"
)

var _ types.QueryServer = Keeper{}
