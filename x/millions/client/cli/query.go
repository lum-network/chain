package cli

import (
	"context"
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"

	"github.com/lum-network/chain/x/millions/types"
)

func GetQueryCmd(queryRoute string) *cobra.Command {
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      fmt.Sprintf("Querying commands for the %s module", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	groupAll := &cobra.Group{Title: "All Commands:", ID: fmt.Sprintf("%s/%s", cmd.Use, "all")}
	groupPool := &cobra.Group{Title: "Pool Commands:", ID: fmt.Sprintf("%s/%s", cmd.Use, "pool")}
	groupDraw := &cobra.Group{Title: "Draw Commands:", ID: fmt.Sprintf("%s/%s", cmd.Use, "draw")}
	groupAccount := &cobra.Group{Title: "Account Commands:", ID: fmt.Sprintf("%s/%s", cmd.Use, "account")}
	groupEpoch := &cobra.Group{Title: "Epoch Commands:", ID: fmt.Sprintf("%s/%s", cmd.Use, "epoch")}
	cmd.AddGroup(
		groupAll,
		groupPool,
		groupDraw,
		groupAccount,
		groupEpoch,
	)

	cmd.AddCommand(
		GetCmdParams(),
		GetCmdPools(groupAll.ID),
		GetCmdPool(groupPool.ID),
		GetCmdDeposits(groupAll.ID),
		GetCmdPoolDeposits(groupPool.ID),
		GetCmdPoolDeposit(groupPool.ID),
		GetCmdAccountDeposits(groupAccount.ID),
		GetCmdAccountPoolDeposits(groupAccount.ID),
		GetCmdDraws(groupAll.ID),
		GetCmdPoolDraws(groupPool.ID),
		GetCmdPoolDraw(groupDraw.ID),
		GetCmdPrizes(groupAll.ID),
		GetCmdPoolPrizes(groupPool.ID),
		GetCmdPoolDrawPrizes(groupDraw.ID),
		GetCmdPoolDrawPrize(groupDraw.ID),
		GetCmdAccountPrizes(groupAccount.ID),
		GetCmdAccountPoolPrizes(groupAccount.ID),
		GetCmdAccountPoolDrawPrizes(groupAccount.ID),
		GetCmdWithdrawals(groupAll.ID),
		GetCmdPoolWithdrawals(groupPool.ID),
		GetCmdPoolWithdrawal(groupPool.ID),
		GetCmdAccountWithdrawals(groupAccount.ID),
		GetCmdAccountPoolWithdrawals(groupAccount.ID),
		GetCmdEpochUnbondings(groupEpoch.ID),
		GetCmdEpochPoolUnbonding(groupEpoch.ID),
	)

	return cmd
}

func GetCmdParams() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "params",
		Short: "Query the millions parameters",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			params := &types.QueryParamsRequest{}
			res, err := queryClient.Params(context.Background(), params)

			if err != nil {
				return err
			}

			return clientCtx.PrintProto(&res.Params)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

func GetCmdPools(groupID string) *cobra.Command {
	cmd := &cobra.Command{
		GroupID: groupID,
		Use:     "pools",
		Short:   "Query the millions pools",
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Acquire the client instance
			clientCtx := client.GetClientContextFromCmd(cmd)

			// Acquire the pagination
			pageReq, err := client.ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}

			// Acquire the query client from the context
			queryClient := types.NewQueryClient(clientCtx)

			// Construct the params payload
			params := &types.QueryPoolsRequest{
				Pagination: pageReq,
			}

			// Construct the query
			res, err := queryClient.Pools(cmd.Context(), params)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}
	flags.AddQueryFlagsToCmd(cmd)
	flags.AddPaginationFlagsToCmd(cmd, "all pools")
	return cmd
}

func GetCmdPool(groupID string) *cobra.Command {
	cmd := &cobra.Command{
		GroupID: groupID,
		Use:     "pool <pool_id>",
		Short:   "Query a millions pool",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Acquire the client instance
			clientCtx := client.GetClientContextFromCmd(cmd)

			// Acquire the query client
			queryClient := types.NewQueryClient(clientCtx)

			// Construct the payload
			poolID, err := strconv.ParseUint(args[0], 0, 64)
			if err != nil {
				return err
			}

			params := &types.QueryPoolRequest{
				PoolId: poolID,
			}

			// Post and acquire the response
			res, err := queryClient.Pool(cmd.Context(), params)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}
	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

func GetCmdDeposits(groupID string) *cobra.Command {
	cmd := &cobra.Command{
		GroupID: groupID,
		Use:     "deposits",
		Short:   "Query the deposits from all the pools",
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Acquire the client instance
			clientCtx := client.GetClientContextFromCmd(cmd)

			// Acquire the pagination
			pageReq, err := client.ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}

			// Acquire the query client from the context
			queryClient := types.NewQueryClient(clientCtx)

			// Construct the params payload
			params := &types.QueryDepositsRequest{
				Pagination: pageReq,
			}

			// Construct the query
			res, err := queryClient.Deposits(cmd.Context(), params)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}
	flags.AddQueryFlagsToCmd(cmd)
	flags.AddPaginationFlagsToCmd(cmd, "all deposits")
	return cmd
}

func GetCmdPoolDeposits(groupID string) *cobra.Command {
	cmd := &cobra.Command{
		GroupID: groupID,
		Use:     "pool-deposits <pool_id>",
		Short:   "Query the deposits for a given poolID",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Acquire the client instance
			clientCtx := client.GetClientContextFromCmd(cmd)

			// Acquire the pagination
			pageReq, err := client.ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}

			// Acquire the query client from the context
			queryClient := types.NewQueryClient(clientCtx)

			poolId, err := strconv.ParseUint(args[0], 0, 64)
			if err != nil {
				return err
			}

			// Construct the params payload
			params := &types.QueryPoolDepositsRequest{
				PoolId:     poolId,
				Pagination: pageReq,
			}

			// Construct the query
			res, err := queryClient.PoolDeposits(cmd.Context(), params)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}
	flags.AddQueryFlagsToCmd(cmd)
	flags.AddPaginationFlagsToCmd(cmd, "all deposits by pool")
	return cmd
}

func GetCmdPoolDeposit(groupID string) *cobra.Command {
	cmd := &cobra.Command{
		GroupID: groupID,
		Use:     "pool-deposit <pool_id> <deposit_id>",
		Short:   "Query a deposit by ID for a given poolID",
		Args:    cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Acquire the client instance
			clientCtx := client.GetClientContextFromCmd(cmd)

			// Acquire the query client from the context
			queryClient := types.NewQueryClient(clientCtx)

			poolID, err := strconv.ParseUint(args[0], 0, 64)
			if err != nil {
				return err
			}
			depositID, err := strconv.ParseUint(args[1], 0, 64)
			if err != nil {
				return err
			}

			// Construct the params payload
			params := &types.QueryPoolDepositRequest{
				PoolId:    poolID,
				DepositId: depositID,
			}

			// Construct the query
			res, err := queryClient.PoolDeposit(cmd.Context(), params)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}
	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

func GetCmdAccountDeposits(groupID string) *cobra.Command {
	cmd := &cobra.Command{
		GroupID: groupID,
		Use:     "account-deposits <address>",
		Short:   "Query deposits for a given account address",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Acquire the client instance
			clientCtx := client.GetClientContextFromCmd(cmd)

			// Acquire the pagination
			pageReq, err := client.ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}

			// Acquire the query client from the context
			queryClient := types.NewQueryClient(clientCtx)

			// Construct the params payload
			params := &types.QueryAccountDepositsRequest{
				DepositorAddress: args[0],
				Pagination:       pageReq,
			}

			// Construct the query
			res, err := queryClient.AccountDeposits(cmd.Context(), params)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}
	flags.AddQueryFlagsToCmd(cmd)
	flags.AddPaginationFlagsToCmd(cmd, "all deposits by pool")
	return cmd
}

func GetCmdAccountPoolDeposits(groupID string) *cobra.Command {
	cmd := &cobra.Command{
		GroupID: groupID,
		Use:     "account-pool-deposits <address> <pool_id>",
		Short:   "Query the pool deposits for a given account address",
		Args:    cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Acquire the client instance
			clientCtx := client.GetClientContextFromCmd(cmd)

			// Acquire the pagination
			pageReq, err := client.ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}

			// Acquire the query client from the context
			queryClient := types.NewQueryClient(clientCtx)

			poolID, err := strconv.ParseUint(args[1], 0, 64)
			if err != nil {
				return err
			}

			// Construct the params payload
			params := &types.QueryAccountPoolDepositsRequest{
				DepositorAddress: args[0],
				PoolId:           poolID,
				Pagination:       pageReq,
			}

			// Construct the query
			res, err := queryClient.AccountPoolDeposits(cmd.Context(), params)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}
	flags.AddQueryFlagsToCmd(cmd)
	flags.AddPaginationFlagsToCmd(cmd, "all deposits by address and pool")
	return cmd
}

func GetCmdDraws(groupID string) *cobra.Command {
	cmd := &cobra.Command{
		GroupID: groupID,
		Use:     "draws",
		Short:   "Query the draws from all pools",
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Acquire the client instance
			clientCtx := client.GetClientContextFromCmd(cmd)

			// Acquire the pagination
			pageReq, err := client.ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}

			// Acquire the query client from the context
			queryClient := types.NewQueryClient(clientCtx)

			// Construct the params payload
			params := &types.QueryDrawsRequest{
				Pagination: pageReq,
			}

			// Construct the query
			res, err := queryClient.Draws(cmd.Context(), params)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}
	flags.AddQueryFlagsToCmd(cmd)
	flags.AddPaginationFlagsToCmd(cmd, "all draws")
	return cmd
}

func GetCmdPoolDraws(groupID string) *cobra.Command {
	cmd := &cobra.Command{
		GroupID: groupID,
		Use:     "pool-draws <pool_id>",
		Short:   "Query the draws for a given pool",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Acquire the client instance
			clientCtx := client.GetClientContextFromCmd(cmd)

			// Acquire the pagination
			pageReq, err := client.ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}

			poolID, err := strconv.ParseUint(args[0], 0, 64)
			if err != nil {
				return err
			}

			// Acquire the query client from the context
			queryClient := types.NewQueryClient(clientCtx)

			// Construct the params payload
			params := &types.QueryPoolDrawsRequest{
				PoolId:     poolID,
				Pagination: pageReq,
			}

			// Construct the query
			res, err := queryClient.PoolDraws(cmd.Context(), params)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}
	flags.AddQueryFlagsToCmd(cmd)
	flags.AddPaginationFlagsToCmd(cmd, "all draws by pool")
	return cmd
}

func GetCmdPoolDraw(groupID string) *cobra.Command {
	cmd := &cobra.Command{
		GroupID: groupID,
		Use:     "draw <pool_id> <draw_id>",
		Short:   "Query a pool draw by ID",
		Args:    cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Acquire the client instance
			clientCtx := client.GetClientContextFromCmd(cmd)

			poolID, err := strconv.ParseUint(args[0], 0, 64)
			if err != nil {
				return err
			}

			drawID, err := strconv.ParseUint(args[1], 0, 64)
			if err != nil {
				return err
			}

			// Acquire the query client from the context
			queryClient := types.NewQueryClient(clientCtx)

			// Construct the params payload
			params := &types.QueryPoolDrawRequest{
				PoolId: poolID,
				DrawId: drawID,
			}

			// Construct the query
			res, err := queryClient.PoolDraw(cmd.Context(), params)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}
	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

func GetCmdPrizes(groupID string) *cobra.Command {
	cmd := &cobra.Command{
		GroupID: groupID,
		Use:     "prizes",
		Short:   "Query all prizes from all draws and all pools",
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Acquire the client instance
			clientCtx := client.GetClientContextFromCmd(cmd)

			// Acquire the pagination
			pageReq, err := client.ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}

			// Acquire the query client from the context
			queryClient := types.NewQueryClient(clientCtx)

			// Construct the params payload
			params := &types.QueryPrizesRequest{
				Pagination: pageReq,
			}

			// Construct the query
			res, err := queryClient.Prizes(cmd.Context(), params)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}
	flags.AddQueryFlagsToCmd(cmd)
	flags.AddPaginationFlagsToCmd(cmd, "all prizes")
	return cmd
}

func GetCmdPoolPrizes(groupID string) *cobra.Command {
	cmd := &cobra.Command{
		GroupID: groupID,
		Use:     "pool-prizes <pool_id>",
		Short:   "Query all prizes from all draws for a given pool",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Acquire the client instance
			clientCtx := client.GetClientContextFromCmd(cmd)

			// Acquire the pagination
			pageReq, err := client.ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}

			// Acquire the query client from the context
			queryClient := types.NewQueryClient(clientCtx)

			poolID, err := strconv.ParseUint(args[0], 0, 64)
			if err != nil {
				return err
			}

			// Construct the params payload
			params := &types.QueryPoolPrizesRequest{
				PoolId:     poolID,
				Pagination: pageReq,
			}

			// Construct the query
			res, err := queryClient.PoolPrizes(cmd.Context(), params)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}
	flags.AddQueryFlagsToCmd(cmd)
	flags.AddPaginationFlagsToCmd(cmd, "all prizes by pool")
	return cmd
}

func GetCmdPoolDrawPrizes(groupID string) *cobra.Command {
	cmd := &cobra.Command{
		GroupID: groupID,
		Use:     "draw-prizes <pool_id> <draw_id>",
		Short:   "Query all prizes from a pool draw",
		Args:    cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Acquire the client instance
			clientCtx := client.GetClientContextFromCmd(cmd)

			// Acquire the pagination
			pageReq, err := client.ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}

			// Acquire the query client from the context
			queryClient := types.NewQueryClient(clientCtx)

			poolID, err := strconv.ParseUint(args[0], 0, 64)
			if err != nil {
				return err
			}
			drawID, err := strconv.ParseUint(args[1], 0, 64)
			if err != nil {
				return err
			}

			// Construct the params payload
			params := &types.QueryPoolDrawPrizesRequest{
				PoolId:     poolID,
				DrawId:     drawID,
				Pagination: pageReq,
			}

			// Construct the query
			res, err := queryClient.PoolDrawPrizes(cmd.Context(), params)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}
	flags.AddQueryFlagsToCmd(cmd)
	flags.AddPaginationFlagsToCmd(cmd, "all prizes by draw")
	return cmd
}

func GetCmdPoolDrawPrize(groupID string) *cobra.Command {
	cmd := &cobra.Command{
		GroupID: groupID,
		Use:     "draw-prize <pool_id> <draw_id> <prize_id>",
		Short:   "Query a prize from a pool draw",
		Args:    cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Acquire the client instance
			clientCtx := client.GetClientContextFromCmd(cmd)

			// Acquire the query client from the context
			queryClient := types.NewQueryClient(clientCtx)

			poolID, err := strconv.ParseUint(args[0], 0, 64)
			if err != nil {
				return err
			}
			drawID, err := strconv.ParseUint(args[1], 0, 64)
			if err != nil {
				return err
			}
			prizeID, err := strconv.ParseUint(args[2], 0, 64)
			if err != nil {
				return err
			}

			// Construct the params payload
			params := &types.QueryPoolDrawPrizeRequest{
				PoolId:  poolID,
				DrawId:  drawID,
				PrizeId: prizeID,
			}

			// Construct the query
			res, err := queryClient.PoolDrawPrize(cmd.Context(), params)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}
	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

func GetCmdAccountPrizes(groupID string) *cobra.Command {
	cmd := &cobra.Command{
		GroupID: groupID,
		Use:     "account-prizes <address>",
		Short:   "Query all prizes for a given account address",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Acquire the client instance
			clientCtx := client.GetClientContextFromCmd(cmd)

			// Acquire the pagination
			pageReq, err := client.ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}

			// Acquire the query client from the context
			queryClient := types.NewQueryClient(clientCtx)

			// Construct the params payload
			params := &types.QueryAccountPrizesRequest{
				WinnerAddress: args[0],
				Pagination:    pageReq,
			}

			// Construct the query
			res, err := queryClient.AccountPrizes(cmd.Context(), params)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}
	flags.AddQueryFlagsToCmd(cmd)
	flags.AddPaginationFlagsToCmd(cmd, "all prizes for a given address")
	return cmd
}

func GetCmdAccountPoolPrizes(groupID string) *cobra.Command {
	cmd := &cobra.Command{
		GroupID: groupID,
		Use:     "account-pool-prizes <address> <pool_id>",
		Short:   "Query the pool prizes for a given account address",
		Args:    cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Acquire the client instance
			clientCtx := client.GetClientContextFromCmd(cmd)

			// Acquire the pagination
			pageReq, err := client.ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}

			// Acquire the query client from the context
			queryClient := types.NewQueryClient(clientCtx)

			poolID, err := strconv.ParseUint(args[1], 0, 64)
			if err != nil {
				return err
			}

			// Construct the params payload
			params := &types.QueryAccountPoolPrizesRequest{
				WinnerAddress: args[0],
				PoolId:        poolID,
				Pagination:    pageReq,
			}

			// Construct the query
			res, err := queryClient.AccountPoolPrizes(cmd.Context(), params)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}
	flags.AddQueryFlagsToCmd(cmd)
	flags.AddPaginationFlagsToCmd(cmd, "all pool prizes by address")
	return cmd
}

func GetCmdAccountPoolDrawPrizes(groupID string) *cobra.Command {
	cmd := &cobra.Command{
		GroupID: groupID,
		Use:     "account-pool-draw-prizes <address> <pool_id> <draw_id>",
		Short:   "Query the draw prizes for a pool and a given account address",
		Args:    cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Acquire the client instance
			clientCtx := client.GetClientContextFromCmd(cmd)

			// Acquire the pagination
			pageReq, err := client.ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}

			// Acquire the query client from the context
			queryClient := types.NewQueryClient(clientCtx)

			poolID, err := strconv.ParseUint(args[1], 0, 64)
			if err != nil {
				return err
			}
			drawID, err := strconv.ParseUint(args[2], 0, 64)
			if err != nil {
				return err
			}

			// Construct the params payload
			params := &types.QueryAccountPoolDrawPrizesRequest{
				WinnerAddress: args[0],
				PoolId:        poolID,
				DrawId:        drawID,
				Pagination:    pageReq,
			}

			// Construct the query
			res, err := queryClient.AccountPoolDrawPrizes(cmd.Context(), params)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}
	flags.AddQueryFlagsToCmd(cmd)
	flags.AddPaginationFlagsToCmd(cmd, "all pool draw prizes by address")
	return cmd
}

func GetCmdWithdrawals(groupID string) *cobra.Command {
	cmd := &cobra.Command{
		GroupID: groupID,
		Use:     "withdrawals",
		Short:   "Query all withdrawals from all pools",
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Acquire the client instance
			clientCtx := client.GetClientContextFromCmd(cmd)

			// Acquire the pagination
			pageReq, err := client.ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}

			// Acquire the query client from the context
			queryClient := types.NewQueryClient(clientCtx)

			// Construct the params payload
			params := &types.QueryWithdrawalsRequest{
				Pagination: pageReq,
			}

			// Construct the query
			res, err := queryClient.Withdrawals(cmd.Context(), params)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}
	flags.AddQueryFlagsToCmd(cmd)
	flags.AddPaginationFlagsToCmd(cmd, "all withdrawals")
	return cmd
}

func GetCmdPoolWithdrawals(groupID string) *cobra.Command {
	cmd := &cobra.Command{
		GroupID: groupID,
		Use:     "pool-withdrawals <pool_id>",
		Short:   "Query all withdrawals for a given pool",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Acquire the client instance
			clientCtx := client.GetClientContextFromCmd(cmd)

			// Acquire the pagination
			pageReq, err := client.ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}

			// Acquire the query client from the context
			queryClient := types.NewQueryClient(clientCtx)

			poolID, err := strconv.ParseUint(args[0], 0, 64)
			if err != nil {
				return err
			}

			// Construct the params payload
			params := &types.QueryPoolWithdrawalsRequest{
				PoolId:     poolID,
				Pagination: pageReq,
			}

			// Construct the query
			res, err := queryClient.PoolWithdrawals(cmd.Context(), params)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}
	flags.AddQueryFlagsToCmd(cmd)
	flags.AddPaginationFlagsToCmd(cmd, "all withdrawals by pool")
	return cmd
}

func GetCmdPoolWithdrawal(groupID string) *cobra.Command {
	cmd := &cobra.Command{
		GroupID: groupID,
		Use:     "pool-withdrawal <pool_id> <withdrawal_id>",
		Short:   "Query withdrawal by ID for a given pool",
		Args:    cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Acquire the client instance
			clientCtx := client.GetClientContextFromCmd(cmd)

			// Acquire the query client from the context
			queryClient := types.NewQueryClient(clientCtx)

			poolID, err := strconv.ParseUint(args[0], 0, 64)
			if err != nil {
				return err
			}
			withdrawalID, err := strconv.ParseUint(args[1], 0, 64)
			if err != nil {
				return err
			}

			// Construct the params payload
			params := &types.QueryPoolWithdrawalRequest{
				PoolId:       poolID,
				WithdrawalId: withdrawalID,
			}

			// Construct the query
			res, err := queryClient.PoolWithdrawal(cmd.Context(), params)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}
	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

func GetCmdAccountWithdrawals(groupID string) *cobra.Command {
	cmd := &cobra.Command{
		GroupID: groupID,
		Use:     "account-withdrawals <address>",
		Short:   "Query withdrawals for a given account address",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Acquire the client instance
			clientCtx := client.GetClientContextFromCmd(cmd)

			// Acquire the pagination
			pageReq, err := client.ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}

			// Acquire the query client from the context
			queryClient := types.NewQueryClient(clientCtx)

			// Construct the params payload
			params := &types.QueryAccountWithdrawalsRequest{
				DepositorAddress: args[0],
				Pagination:       pageReq,
			}

			// Construct the query
			res, err := queryClient.AccountWithdrawals(cmd.Context(), params)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}
	flags.AddQueryFlagsToCmd(cmd)
	flags.AddPaginationFlagsToCmd(cmd, "all withdrawals by address")
	return cmd
}

func GetCmdAccountPoolWithdrawals(groupID string) *cobra.Command {
	cmd := &cobra.Command{
		GroupID: groupID,
		Use:     "account-pool-withdrawals <address> <pool_id>",
		Short:   "Query the pool withdrawals for a given account address",
		Args:    cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Acquire the client instance
			clientCtx := client.GetClientContextFromCmd(cmd)

			// Acquire the pagination
			pageReq, err := client.ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}

			// Acquire the query client from the context
			queryClient := types.NewQueryClient(clientCtx)

			poolID, err := strconv.ParseUint(args[1], 0, 64)
			if err != nil {
				return err
			}

			// Construct the params payload
			params := &types.QueryAccountPoolWithdrawalsRequest{
				DepositorAddress: args[0],
				PoolId:           poolID,
				Pagination:       pageReq,
			}

			// Construct the query
			res, err := queryClient.AccountPoolWithdrawals(cmd.Context(), params)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}
	flags.AddQueryFlagsToCmd(cmd)
	flags.AddPaginationFlagsToCmd(cmd, "all withdrawals by address and pool")
	return cmd
}

func GetCmdEpochUnbondings(groupID string) *cobra.Command {
	cmd := &cobra.Command{
		GroupID: groupID,
		Use:     "epoch-unbondings <epoch_id>",
		Short:   "Query all epoch unbondings for a given epoch",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Acquire the client instance
			clientCtx := client.GetClientContextFromCmd(cmd)

			// Acquire the pagination
			pageReq, err := client.ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}

			// Acquire the query client from the context
			queryClient := types.NewQueryClient(clientCtx)

			epochID, err := strconv.ParseUint(args[0], 0, 64)
			if err != nil {
				return err
			}

			// Construct the params payload
			epochUnbondings := &types.QueryEpochUnbondingsRequest{
				EpochNumber: epochID,
				Pagination:  pageReq,
			}

			// Construct the query
			res, err := queryClient.EpochUnbondings(cmd.Context(), epochUnbondings)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}
	flags.AddQueryFlagsToCmd(cmd)
	flags.AddPaginationFlagsToCmd(cmd, "all epochUnbondings by epoch")
	return cmd
}

func GetCmdEpochPoolUnbonding(groupID string) *cobra.Command {
	cmd := &cobra.Command{
		GroupID: groupID,
		Use:     "epoch-pool-unbonding <epoch_id> <pool_id>",
		Short:   "Query epoch unbonding for a given epoch and poolID",
		Args:    cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Acquire the client instance
			clientCtx := client.GetClientContextFromCmd(cmd)

			// Acquire the query client from the context
			queryClient := types.NewQueryClient(clientCtx)

			epochID, err := strconv.ParseUint(args[0], 0, 64)
			if err != nil {
				return err
			}

			poolID, err := strconv.ParseUint(args[1], 0, 64)
			if err != nil {
				return err
			}

			// Construct the params payload
			epochPoolUnbonding := &types.QueryEpochPoolUnbondingRequest{
				EpochNumber: epochID,
				PoolId:      poolID,
			}

			// Construct the query
			res, err := queryClient.EpochPoolUnbonding(cmd.Context(), epochPoolUnbonding)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}
	flags.AddQueryFlagsToCmd(cmd)
	flags.AddPaginationFlagsToCmd(cmd, "epoch pool unbonding")
	return cmd
}
