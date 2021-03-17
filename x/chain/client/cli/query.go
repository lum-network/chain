package cli

import (
	"context"
	"fmt"
	"github.com/cosmos/cosmos-sdk/client/flags"

	// "strings"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	// "github.com/cosmos/cosmos-sdk/client/flags"
	// sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/lum-network/chain/x/chain/types"
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
		CmdFetchBeams(),
		CmdGetBeam(),
	)

	return cmd
}

// CmdFetchBeams Query the blockchain and print the list of beams
func CmdFetchBeams() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "fetch-beams",
		Short: "Fetch all beams",
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
			params := &types.QueryFetchBeamsRequest{
				Pagination: pageReq,
			}

			// Post and acquire response
			res, err := queryClient.Beams(context.Background(), params)
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

// CmdGetBeam Query the blockchain about a beam and print response
func CmdGetBeam() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get-beam [id]",
		Short: "Get a given beam",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Acquire the client instance
			clientCtx := client.GetClientContextFromCmd(cmd)

			// Acquire the query client
			queryClient := types.NewQueryClient(clientCtx)

			// Construct the payload
			params := &types.QueryGetBeamRequest{
				Id: args[0],
			}

			// Post and acquire the response
			res, err := queryClient.Beam(context.Background(), params)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}
	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}
