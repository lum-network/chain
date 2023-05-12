package cli

import (
	"context"
	"fmt"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/lum-network/chain/x/icqueries/types"
	"github.com/spf13/cobra"
	"strings"
)

// GetQueryCmd returns the cli query commands for this module.
func GetQueryCmd() *cobra.Command {
	// Group lockup queries under a subcommand
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      fmt.Sprintf("Querying commands for the %s module", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		GetCmdListQueries(),
		GetCmdListPendingQueries(),
	)

	return cmd
}

// GetCmdListQueries provides a list of all pending queries (queries that have not have been requested but have not received a response)
func GetCmdListQueries() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list-queries",
		Short: "Query all queries",
		Example: strings.TrimSpace(
			fmt.Sprintf(`$ %s query icqueries list-queries`,
				version.AppName,
			),
		),
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryServiceClient(clientCtx)

			req := &types.QueryQueriesRequest{}

			res, err := queryClient.Queries(context.Background(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// GetCmdListPendingQueries provides a list of all pending queries (queries that have not have been requested but have not received a response)
func GetCmdListPendingQueries() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list-pending-queries",
		Short: "Query all pending queries",
		Example: strings.TrimSpace(
			fmt.Sprintf(`$ %s query icqueries list-pending-queries`,
				version.AppName,
			),
		),
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryServiceClient(clientCtx)

			req := &types.QueryPendingQueriesRequest{}

			res, err := queryClient.PendingQueries(context.Background(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}
