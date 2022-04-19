package cli

import (
	"fmt"
	"github.com/cosmos/cosmos-sdk/client/flags"

	// "strings"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	// "github.com/cosmos/cosmos-sdk/client/flags"
	// sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/lum-network/chain/x/beam/types"
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
		CmdFetchOpenBeamsQueue(),
		CmdFetchClosedBeamsQueue(),
		CmdFetchOldOpenBeamsQueue(),
		CmdGetBeam(),
	)

	return cmd
}

// CmdFetchBeams Query the blockchain and print the list of beams from the store
func CmdFetchBeams() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "fetch",
		Short: "Fetch all the beams from the store",
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
				Old:        false,
				State:      types.BeamState_StateUnspecified,
				Pagination: pageReq,
			}

			// Post and acquire response
			res, err := queryClient.Beams(cmd.Context(), params)
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

// CmdFetchOpenBeamsQueue Query the blockchain and print the list of beams from the open queue system
func CmdFetchOpenBeamsQueue() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "fetch-open-queue",
		Short: "Fetch all the beams from the open queue",
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
				Old:        false,
				State:      types.BeamState_StateOpen,
				Pagination: pageReq,
			}

			// Post and acquire response
			res, err := queryClient.Beams(cmd.Context(), params)
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

// CmdFetchClosedBeamsQueue Query the blockchain and print the list of beams from the closed queue system
func CmdFetchClosedBeamsQueue() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "fetch-closed-queue",
		Short: "Fetch all the beams from the closed queue",
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
				Old:        false,
				State:      types.BeamState_StateClosed,
				Pagination: pageReq,
			}

			// Post and acquire response
			res, err := queryClient.Beams(cmd.Context(), params)
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

// CmdFetchOldOpenBeamsQueue Query the blockchain and print the list of beams from the old open queue system
func CmdFetchOldOpenBeamsQueue() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "fetch-open-old-queue",
		Short: "Fetch all beams from the old open queue",
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
				Old:        true,
				State:      types.BeamState_StateOpen,
				Pagination: pageReq,
			}

			// Post and acquire response
			res, err := queryClient.Beams(cmd.Context(), params)
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
		Use:   "get [id]",
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
			res, err := queryClient.Beam(cmd.Context(), params)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}
	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}
