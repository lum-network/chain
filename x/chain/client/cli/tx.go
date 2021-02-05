package cli

import (
	"fmt"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	// "github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/sandblockio/chain/x/chain/types"
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

	cmd.AddCommand(CmdOpenBeam())
	return cmd
}

// CmdOpenBeam Command definition for beam opening dispatch
func CmdOpenBeam() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "open-beam [amount] [secret]",
		Short: "Open a new beam",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Acquire the command arguments
			argsAmount, err := strconv.ParseInt(args[0], 10, 32)
			argsSecret := args[1]
			if err != nil {
				return err
			}

			// Acquire the client context
			clientCtx := client.GetClientContextFromCmd(cmd)

			// Construct the message and validate
			msg := types.NewMsgOpenBeam(clientCtx.GetFromAddress().String(), int32(argsAmount), argsSecret)
			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			// Broadcast the message
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}
	return cmd
}
