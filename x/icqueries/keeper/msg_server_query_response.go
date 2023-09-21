package keeper

import (
	"context"
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/lum-network/chain/x/icqueries/types"
)

// SubmitQueryResponse Handle ICQ query responses by validating the proof, and calling the query's corresponding callback
func (k msgServer) SubmitQueryResponse(goCtx context.Context, msg *types.MsgSubmitQueryResponse) (*types.MsgSubmitQueryResponseResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// check if the response has an associated query stored on Lum
	query, found := k.GetQuery(ctx, msg.QueryId)
	if !found {
		k.Logger(ctx).Info("ICQ RESPONSE  | Ignoring non-existent query response (note: duplicate responses are nonexistent)")
		return &types.MsgSubmitQueryResponseResponse{}, nil // technically this is an error, but will cause the entire tx to fail if we have one 'bad' message, so we can just no-op here.
	}

	defer ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(types.AttributeKeyQueryId, query.Id),
		),
		sdk.NewEvent(
			"query_response",
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(types.AttributeKeyQueryId, query.Id),
			sdk.NewAttribute(types.AttributeKeyChainId, query.ChainId),
		),
	})

	// Verify the response's proof, if one exists
	err := k.VerifyKeyProof(ctx, msg, query)
	if err != nil {
		k.Logger(ctx).Error(fmt.Sprintf("QUERY PROOF VERIFICATION FAILED - QueryId: %s, Error: %s", query.Id, err.Error()))
		if err = k.InvokeCallback(ctx, msg, query, types.QueryResponseStatus_FAILURE); err != nil {
			return nil, err
		}
		return nil, err
	}

	// Immediately delete the query so it cannot process again
	k.DeleteQuery(ctx, query.Id)

	// Verify the query hasn't expired (if the block time is greater than the TTL timestamp, the query is expired)
	if query.HasTimedOut(ctx.BlockTime()) {
		if err := k.HandleQueryTimeout(ctx, msg, query); err != nil {
			return nil, err
		}
		return &types.MsgSubmitQueryResponseResponse{}, nil
	}

	// Call the query's associated callback function
	if err = k.InvokeCallback(ctx, msg, query, types.QueryResponseStatus_SUCCESS); err != nil {
		return nil, err
	}

	return &types.MsgSubmitQueryResponseResponse{}, nil
}
