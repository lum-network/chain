package cli

import (
	"fmt"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/lum-network/chain/x/dfract/types"
	"github.com/spf13/cobra"
)

// GetQueryCmd returns the cli query commands for this module
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
		CmdGetDepositsForAddress(),
		CmdFetchDeposits(),
	)

	return cmd
}

func CmdGetDepositsForAddress() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "deposits-for-address [address]",
		Short: "Fetch all the deposits for a given address",
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
		Use:   "fetch [type]",
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

			// Construct the params payload
			params := &types.QueryFetchDepositsRequest{
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
