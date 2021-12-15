package rest

import (
	"context"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/types/rest"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/tendermint/tendermint/p2p"
	ctypes "github.com/tendermint/tendermint/rpc/core/types"
	"net/http"
)

func getNodeStatus(clientCtx client.Context) (*ctypes.ResultStatus, error) {
	node, err := clientCtx.GetNode()
	if err != nil {
		return &ctypes.ResultStatus{}, err
	}

	return node.Status(context.Background())
}

// NodeInfoResponse defines a response type that contains node status and version information.
type NodeInfoResponse struct {
	p2p.DefaultNodeInfo `json:"node_info"`

	ApplicationVersion version.Info `json:"application_version"`
}

// REST handler for node info
func NodeInfoRequestHandlerFn(clientCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		status, err := getNodeStatus(clientCtx)
		if rest.CheckInternalServerError(w, err) {
			return
		}

		resp := NodeInfoResponse{
			DefaultNodeInfo:    status.NodeInfo,
			ApplicationVersion: version.NewInfo(),
		}

		rest.PostProcessResponseBare(w, clientCtx, resp)
	}
}

// SyncingResponse defines a response type that contains node syncing information.
type SyncingResponse struct {
	Syncing bool `json:"syncing"`
}

// REST handler for node syncing
func NodeSyncingRequestHandlerFn(clientCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		status, err := getNodeStatus(clientCtx)
		if rest.CheckInternalServerError(w, err) {
			return
		}

		rest.PostProcessResponseBare(w, clientCtx, SyncingResponse{Syncing: status.SyncInfo.CatchingUp})
	}
}
