package cli

import (
	"fmt"
	"github.com/cosmos/cosmos-sdk/client/flags"
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

	cmd.AddCommand(
		CmdOpenBeam(),
		CmdIncreaseBeam(),
		CmdCloseBeam(),
		CmdClaimBeam(),
	)
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
			clientCtx, err := client.ReadPersistentCommandFlags(client.GetClientContextFromCmd(cmd), cmd.Flags())
			if err != nil {
				return err
			}

			// Construct the message and validate
			msg := types.NewMsgOpenBeam(clientCtx.GetFromAddress().String(), int32(argsAmount), argsSecret)
			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			// Broadcast the message
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}
	flags.AddTxFlagsToCmd(cmd)
	_ = cmd.MarkFlagRequired(flags.FlagFrom)
	return cmd
}

// CmdIncreaseBeam Command definition for beam increase dispatch
func CmdIncreaseBeam() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "increase-beam [id] [amount]",
		Short: "Increase a given beam amount",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Acquire the command arguments
			argsAmount, err := strconv.ParseInt(args[1], 10, 32)
			argsId := args[0]
			if err != nil {
				return err
			}

			// Acquire the client context
			clientCtx, err := client.ReadPersistentCommandFlags(client.GetClientContextFromCmd(cmd), cmd.Flags())
			if err != nil {
				return err
			}

			// Construct the message and validate
			msg := types.NewMsgIncreaseBeam(clientCtx.GetFromAddress().String(), argsId, int32(argsAmount))
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

// CmdCloseBeam Command definition for beam close dispatch
func CmdCloseBeam() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "close-beam [id]",
		Short: "Close a given beam",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			argsId := args[0]

			// Acquire the client context
			clientCtx, err := client.ReadPersistentCommandFlags(client.GetClientContextFromCmd(cmd), cmd.Flags())
			if err != nil {
				return err
			}

			// Construct the message and validate
			msg := types.NewMsgCloseBeam(clientCtx.GetFromAddress().String(), argsId)
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

// CmdClaimBeam Command definition for beam claim dispatch
func CmdClaimBeam() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "claim-beam [id]",
		Short: "Claim a given beam",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			argsId := args[0]

			// Acquire the client context
			clientCtx, err := client.ReadPersistentCommandFlags(client.GetClientContextFromCmd(cmd), cmd.Flags())
			if err != nil {
				return err
			}

			// Construct the message and validate
			msg := types.NewMsgClaimBeam(clientCtx.GetFromAddress().String(), argsId)
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
