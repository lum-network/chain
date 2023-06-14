package cli

import (
	"context"
	"fmt"
	"strconv"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/lum-network/chain/x/dfract/types"
	"github.com/spf13/cobra"
)

// GetQueryCmd returns the cli query commands for this module.
func GetQueryCmd(queryRoute string) *cobra.Command {
	// Group chain queries under a subcommand
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      fmt.Sprintf("Querying commands for the %s module", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		GetCmdQueryModuleAccountBalance(),
		GetCmdQueryParams(),
		CmdGetDepositsForAddress(),
		CmdFetchDeposits(),
	)

	return cmd
}

func GetCmdQueryModuleAccountBalance() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "module-account-balance",
		Short: "Query the dfract module account balance",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			req := &types.QueryModuleAccountBalanceRequest{}
			res, err := queryClient.ModuleAccountBalance(context.Background(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

func GetCmdQueryParams() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "params",
		Short: "Query the dfract parameters",
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

func CmdGetDepositsForAddress() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "deposits-for-address <address>",
		Short: "Fetch all the deposits for a given address",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Acquire the client instance
			clientCtx := client.GetClientContextFromCmd(cmd)

			// Acquire the query client from the context
			queryClient := types.NewQueryClient(clientCtx)

			// Construct the params payload
			params := &types.QueryGetDepositsForAddressRequest{
				Address: args[0],
			}

			// Construct the query
			res, err := queryClient.GetDepositsForAddress(cmd.Context(), params)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}
	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

func CmdFetchDeposits() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "fetch <type>",
		Short: "Fetch all the deposits for a given type",
		Args:  cobra.ExactArgs(1),
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

			// Acquire the type to fetch
			depositType, err := strconv.ParseInt(args[0], 10, 32)
			if err != nil {
				return err
			}

			// Construct the params payload
			params := &types.QueryFetchDepositsRequest{
				Type:       types.DepositsQueryType(depositType),
				Pagination: pageReq,
			}

			// Construct the query
			res, err := queryClient.FetchDeposits(cmd.Context(), params)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}
	flags.AddQueryFlagsToCmd(cmd)
	flags.AddPaginationFlagsToCmd(cmd, "all beams")
	return cmd
}
