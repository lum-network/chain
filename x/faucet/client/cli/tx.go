package cli

import (
	"fmt"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"time"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	// "github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/sandblockio/chain/x/faucet/types"
)

// GetTxCmd returns the transaction commands for this module
func GetTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      fmt.Sprintf("%s transactions subcommands", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(CmdMintAndSend())

	return cmd
}

// CmdMintAndSend Command definition for faucet minting
func CmdMintAndSend() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "mint",
		Short: "Ask the faucet to mint a fixed amount of coin",
		Args:  cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Acquire the client instance
			clientCtx, err := client.ReadTxCommandFlags(client.GetClientContextFromCmd(cmd), cmd.Flags())
			if err != nil {
				return err
			}

			// Construct the message and validate
			msg := types.NewMsgMintAndSend(clientCtx.GetFromAddress().String(), time.Now())
			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}
	flags.AddTxFlagsToCmd(cmd)
	_ = cmd.MarkFlagRequired(flags.FlagFrom)
	return cmd
}
