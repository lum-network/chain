package keeper

import (
	"encoding/hex"
	"fmt"
	"time"

	"github.com/lum-network/chain/x/icqueries/types"

	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// EndBlocker of icqueries module.
func (k *Keeper) EndBlocker(ctx sdk.Context) {
	defer telemetry.ModuleMeasureSince(types.ModuleName, time.Now(), telemetry.MetricKeyEndBlocker)
	logger := k.Logger(ctx).With("ctx", "blocker_icq")

	// Loop through queries and emit them
	k.IterateQueries(ctx, func(_ int64, query types.Query) (stop bool) {
		if !query.RequestSent {
			events := sdk.Events{}
			logger.Info(fmt.Sprintf("InterchainQuery event emitted %s", query.Id))
			// QUESTION: Do we need to emit this event twice?
			event := sdk.NewEvent(
				sdk.EventTypeMessage,
				sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
				sdk.NewAttribute(sdk.AttributeKeyAction, types.AttributeValueQuery),
				sdk.NewAttribute(types.AttributeKeyQueryID, query.Id),
				sdk.NewAttribute(types.AttributeKeyChainID, query.ChainId),
				sdk.NewAttribute(types.AttributeKeyConnectionID, query.ConnectionId),
				sdk.NewAttribute(types.AttributeKeyType, query.QueryType),
				sdk.NewAttribute(types.AttributeKeyHeight, "0"),
				sdk.NewAttribute(types.AttributeKeyRequest, hex.EncodeToString(query.Request)),
			)
			events = append(events, event)

			event.Type = "query_request"
			events = append(events, event)

			query.RequestSent = true
			k.SetQuery(ctx, query)

			// Emit events, it will append
			ctx.EventManager().EmitEvents(events)
			logger.Info(fmt.Sprintf("[ICQ] Emitted a total of %d events, on block %d for query %s", len(events), ctx.BlockHeight(), query.Id))
		}
		return false
	})
}
