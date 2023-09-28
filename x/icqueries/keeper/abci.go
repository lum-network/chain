package keeper

import (
	"encoding/hex"
	"fmt"
	"time"

	"github.com/lum-network/chain/x/icqueries/types"

	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// EndBlocker of icqueries module
func (k *Keeper) EndBlocker(ctx sdk.Context) {
	defer telemetry.ModuleMeasureSince(types.ModuleName, time.Now(), telemetry.MetricKeyEndBlocker)
	logger := k.Logger(ctx).With("ctx", "blocker_icq")

	events := sdk.Events{}

	// We can loop on each query since when response is received, we delete them, there will never be much to process
	for _, query := range k.AllQueries(ctx) {
		if query.RequestSent {
			// If the query has timed out here, it means we still haven't received the relayer timeout notification.
			// This is fault-tolerant piece of code, since if the relayer confirmation finally arrives, it will just get ignored
			if query.HasTimedOut(ctx.BlockTime()) {
				// Try to invoke the timeout query, but continue in case of failure (panicking here is heavily dangerous)
				if err := k.HandleQueryTimeout(ctx, nil, query); err != nil {
					logger.Error(fmt.Sprintf("Error encountered while handling query timeout : %v", err.Error()))
					continue
				}
				// If everything went fine, just delete the query data
				k.DeleteQuery(ctx, query.Id)
			}
			continue
		}
		logger.Info(fmt.Sprintf("InterchainQuery event emitted %s", query.Id))
		event := sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeyAction, types.AttributeValueQuery),
			sdk.NewAttribute(types.AttributeKeyQueryId, query.Id),
			sdk.NewAttribute(types.AttributeKeyChainId, query.ChainId),
			sdk.NewAttribute(types.AttributeKeyConnectionId, query.ConnectionId),
			sdk.NewAttribute(types.AttributeKeyType, query.QueryType),
			sdk.NewAttribute(types.AttributeKeyHeight, "0"),
			sdk.NewAttribute(types.AttributeKeyRequest, hex.EncodeToString(query.Request)),
		)

		// Emit our event twice
		events = append(events, event)
		event.Type = "query_request"
		events = append(events, event)

		// Patch our query
		query.RequestSent = true
		k.SetQuery(ctx, query)
	}

	if len(events) > 0 {
		ctx.EventManager().EmitEvents(events)
	}
}
