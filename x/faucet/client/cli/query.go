package cli

import (
	"fmt"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"time"

	// "strings"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	// "github.com/cosmos/cosmos-sdk/client/flags"
	// sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/lum-network/chain/x/faucet/types"
)

// GetQueryCmd returns the cli query commands for this module
func GetQueryCmd(queryRoute string) *cobra.Command {
	// Group faucet queries under a subcommand
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      fmt.Sprintf("Querying commands for the %s module", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}
	cmd.AddCommand(CmdMintAndSend())
	return cmd
}

// CmdMintAndSend mint and send magic
func CmdMintAndSend() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "mint-and-send [minter]",
		Short: "Mint and send coins to a given address",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Acquire the client instance
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			// Acquire the query client
			queryClient := types.NewQueryClient(clientCtx)

			// Construct the payload
			params := &types.QueryMintAndSendRequest{
				Minter:   args[0],
				MintTime: time.Now(),
			}

			fmt.Println(args[0])

			// Post and acquire the response
			res, err := queryClient.MintAndSend(cmd.Context(), params)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}
	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}
