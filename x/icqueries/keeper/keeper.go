package keeper

import (
	"fmt"
	"sort"
	"strings"
	"time"

	ibcconnectiontypes "github.com/cosmos/ibc-go/v7/modules/core/03-connection/types"

	errorsmod "cosmossdk.io/errors"
	"github.com/cometbft/cometbft/libs/log"
	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	clienttypes "github.com/cosmos/ibc-go/v7/modules/core/02-client/types"
	commitmenttypes "github.com/cosmos/ibc-go/v7/modules/core/23-commitment/types"
	"github.com/cosmos/ibc-go/v7/modules/core/exported"
	ibckeeper "github.com/cosmos/ibc-go/v7/modules/core/keeper"
	tendermint "github.com/cosmos/ibc-go/v7/modules/light-clients/07-tendermint"
	ics23 "github.com/cosmos/ics23/go"
	"github.com/spf13/cast"

	"github.com/lum-network/chain/x/icqueries/types"
)

type Keeper struct {
	cdc       codec.Codec
	storeKey  storetypes.StoreKey
	callbacks map[string]types.QueryCallbacks
	IBCKeeper *ibckeeper.Keeper
}

func NewKeeper(cdc codec.Codec, storeKey storetypes.StoreKey, ibckeeper *ibckeeper.Keeper) *Keeper {
	return &Keeper{
		cdc:       cdc,
		storeKey:  storeKey,
		callbacks: make(map[string]types.QueryCallbacks),
		IBCKeeper: ibckeeper,
	}
}

func (k *Keeper) SetCallbackHandler(module string, handler types.QueryCallbacks) error {
	_, found := k.callbacks[module]
	if found {
		return fmt.Errorf("callback handler already set for %s", module)
	}
	k.callbacks[module] = handler.RegisterICQCallbacks()
	return nil
}

// Logger returns a module-specific logger.
func (k *Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

func (k *Keeper) MakeRequest(ctx sdk.Context, module string, callbackId string, chainId string, connectionId string, extraId string, queryType string, request []byte, timeoutPolicy types.TimeoutPolicy, timeoutDuration time.Duration) error {
	k.Logger(ctx).Info(fmt.Sprintf("Submitting ICQ Request - module=%s, callbackId=%s, connectionId=%s, queryType=%s", module, callbackId, connectionId, queryType))

	// Compute our timestamp ttl
	timeoutTimestamp := uint64(ctx.BlockTime().UnixNano() + timeoutDuration.Nanoseconds())

	// Set the submission height on the Query to the latest light client height
	// In the query response, this will be used to verify that the query wasn't historical
	connection, found := k.IBCKeeper.ConnectionKeeper.GetConnection(ctx, connectionId)
	if !found {
		return errorsmod.Wrapf(ibcconnectiontypes.ErrConnectionNotFound, connectionId)
	}
	clientState, found := k.IBCKeeper.ClientKeeper.GetClientState(ctx, connection.ClientId)
	if !found {
		return errorsmod.Wrapf(clienttypes.ErrClientNotFound, connection.ClientId)
	}

	// Construct our query payload and validate it
	query := k.NewQuery(ctx, module, callbackId, chainId, connectionId, extraId, queryType, request, timeoutTimestamp, timeoutPolicy, timeoutDuration, clientState.GetLatestHeight().GetRevisionHeight())
	if err := k.ValidateQuery(ctx, query); err != nil {
		return err
	}

	// Store the actual query
	k.SetQuery(ctx, *query)
	return nil
}

// RetryRequest Re-submit an ICQ, generally used after a timeout
func (k *Keeper) RetryRequest(ctx sdk.Context, queryId string) error {
	query, found := k.GetQuery(ctx, queryId)
	if !found {
		return errorsmod.Wrapf(types.ErrInvalidQuery, "Query %s not found", queryId)
	}

	k.Logger(ctx).Info(fmt.Sprintf("Queuing ICQ Retry - Query Type: %s, Query ID: %s", query.CallbackId, query.Id))

	// Patch our query for re-scheduling by patching the flag, and resetting the timeout timestamp to now + duration (avoid instant timeout of it)
	query.RequestSent = false
	query.TimeoutTimestamp = uint64(ctx.BlockTime().UnixNano() + query.TimeoutDuration.Nanoseconds())
	k.SetQuery(ctx, query)
	return nil
}

// VerifyKeyProof check if the query requires proving; if it does, verify it!
func (k Keeper) VerifyKeyProof(ctx sdk.Context, msg *types.MsgSubmitQueryResponse, query types.Query) error {
	pathParts := strings.Split(query.QueryType, "/")

	// the query does NOT have an associated proof, so no need to verify it.
	if pathParts[len(pathParts)-1] != "key" {
		return nil
	}

	// If the query is a "key" proof query, verify the results are valid by checking the poof
	if msg.ProofOps == nil {
		return errorsmod.Wrapf(types.ErrInvalidICQProof, "Unable to validate proof. No proof submitted")
	}

	// Get the client consensus state at the height 1 block above the message height
	proofHeight, err := cast.ToUint64E(msg.Height)
	if err != nil {
		return err
	}
	height := clienttypes.NewHeight(clienttypes.ParseChainID(query.ChainId), proofHeight+1)

	// Confirm the query proof height occurred after the submission height
	if proofHeight <= query.SubmissionHeight {
		return errorsmod.Wrapf(types.ErrInvalidICQProof, "Query proof height (%d) is older than the submission height (%d)", proofHeight, query.SubmissionHeight)
	}

	// Get the client state and consensus state from the ConnectionId
	connection, found := k.IBCKeeper.ConnectionKeeper.GetConnection(ctx, query.ConnectionId)
	if !found {
		return errorsmod.Wrapf(types.ErrInvalidICQProof, "ConnectionId %s does not exist", query.ConnectionId)
	}
	consensusState, found := k.IBCKeeper.ClientKeeper.GetClientConsensusState(ctx, connection.ClientId, height)
	if !found {
		return errorsmod.Wrapf(types.ErrInvalidICQProof, "Consensus state not found for client %s and height %d", connection.ClientId, height)
	}
	clientState, found := k.IBCKeeper.ClientKeeper.GetClientState(ctx, connection.ClientId)
	if !found {
		return errorsmod.Wrapf(types.ErrInvalidICQProof, "Unable to fetch client state for client %s", connection.ClientId)
	}

	// Cast the client and consensus state to tendermint type
	tendermintConsensusState, ok := consensusState.(*tendermint.ConsensusState)
	if !ok {
		return errorsmod.Wrapf(types.ErrInvalidICQProof, "Only tendermint consensus state is supported (%s provided)", consensusState.ClientType())
	}
	tendermintClientState, ok := clientState.(*tendermint.ClientState)
	if !ok {
		return errorsmod.Wrapf(types.ErrInvalidICQProof, "Only tendermint client state is supported (%s provided)", clientState.ClientType())
	}
	var stateRoot exported.Root = tendermintConsensusState.Root
	var clientStateProof []*ics23.ProofSpec = tendermintClientState.ProofSpecs

	// Get the merkle path and merkle proof
	path := commitmenttypes.NewMerklePath([]string{pathParts[1], string(query.Request)}...)
	merkleProof, err := commitmenttypes.ConvertProofs(msg.ProofOps)
	if err != nil {
		return errorsmod.Wrapf(types.ErrInvalidICQProof, "Error converting proofs: %s", err.Error())
	}

	// If we got a non-nil response, verify inclusion proof
	if len(msg.Result) != 0 {
		if err := merkleProof.VerifyMembership(clientStateProof, stateRoot, path, msg.Result); err != nil {
			return errorsmod.Wrapf(types.ErrInvalidICQProof, "Unable to verify membership proof: %s", err.Error())
		}
		k.Logger(ctx).Info(fmt.Sprintf("Inclusion proof validated - QueryId %s", query.Id))

	} else {
		// if we got a nil query response, verify non inclusion proof.
		if err := merkleProof.VerifyNonMembership(clientStateProof, stateRoot, path); err != nil {
			return errorsmod.Wrapf(types.ErrInvalidICQProof, "Unable to verify non-membership proof: %s", err.Error())
		}
		k.Logger(ctx).Info(fmt.Sprintf("Non-inclusion proof validated - QueryId %s", query.Id))
	}

	return nil
}

// HandleQueryTimeout Handles a query timeout based on the timeout policy
func (k Keeper) HandleQueryTimeout(ctx sdk.Context, msg *types.MsgSubmitQueryResponse, query types.Query) error {
	k.Logger(ctx).Error(fmt.Sprintf("QUERY TIMEOUT - QueryId: %s, TTL: %d, BlockTime: %d", query.Id, query.TimeoutTimestamp, ctx.BlockHeader().Time.UnixNano()))

	switch query.TimeoutPolicy {
	case types.TimeoutPolicy_REJECT_QUERY_RESPONSE:
		k.Logger(ctx).Info(fmt.Sprintf("Rejecting query %s", query.GetId()))
		return nil

	case types.TimeoutPolicy_RETRY_QUERY_REQUEST:
		k.Logger(ctx).Info(fmt.Sprintf("Retrying query %s", query.GetId()))
		return k.RetryRequest(ctx, query.GetId())

	case types.TimeoutPolicy_EXECUTE_QUERY_CALLBACK:
		k.Logger(ctx).Info(fmt.Sprintf("Invoking query %s callback", query.GetId()))
		return k.InvokeCallback(ctx, msg, query, types.QueryResponseStatus_TIMEOUT)

	default:
		return fmt.Errorf("unsupported query timeout policy: %s", query.TimeoutPolicy.String())
	}
}

func (k Keeper) InvokeCallback(ctx sdk.Context, msg *types.MsgSubmitQueryResponse, query types.Query, status types.QueryResponseStatus) error {
	// Extract the result payload
	var result []byte
	if msg != nil {
		result = msg.Result
	}

	// get all the callback handlers and sort them for determinism
	// (each module has their own callback handler)
	moduleNames := []string{}
	for moduleName := range k.callbacks {
		moduleNames = append(moduleNames, moduleName)
	}
	sort.Strings(moduleNames)

	// Loop through each module until the callbackId is found in one of the module handlers
	for _, moduleName := range moduleNames {
		moduleCallbackHandler := k.callbacks[moduleName]

		// Once the callback is found, invoke the function
		if moduleCallbackHandler.HasICQCallback(query.CallbackId) {
			return moduleCallbackHandler.CallICQCallback(ctx, query.CallbackId, result, query, status)
		}
	}

	// If no callback was found, return an error
	return types.ErrICQCallbackNotFound
}
