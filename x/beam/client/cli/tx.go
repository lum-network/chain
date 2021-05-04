package cli

import (
	"fmt"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/lum-network/chain/x/beam/types"
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
		CmdUpdateBeam(),
		CmdCloseBeam(),
		CmdClaimBeam(),
	)
	return cmd
}

// CmdOpenBeam Command definition for beam opening dispatch
func CmdOpenBeam() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "open [amount] [secret]",
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
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			// Generate the random id
			id := types.GenerateSecureToken(10)

			// Encode the secret
			hashedSecret, err := types.GenerateHashFromString(argsSecret)
			if err != nil {
				return err
			}

			// Construct the message and validate
			msg := types.NewMsgOpenBeam(id, clientCtx.GetFromAddress().String(), int64(argsAmount), hashedSecret)
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

// CmdUpdateBeam Command definition for beam increase dispatch
func CmdUpdateBeam() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id> <amount> [reward] [review]",
		Short: "Update a given beam",
		Args:  cobra.RangeArgs(2, 4),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Acquire the command arguments
			argsAmount, err := strconv.ParseInt(args[1], 10, 32)
			argsId := args[0]
			if err != nil {
				return err
			}

			// Try to acquire the reward and review arguments
			argsReward := &args[2]
			argsReview := &args[3]

			// Acquire the client context
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			// Construct the message and validate
			msg := types.NewMsgUpdateBeam(clientCtx.GetFromAddress().String(), argsId, int64(argsAmount), argsReward, argsReview)
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
		Use:   "close [id]",
		Short: "Close a given beam",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			argsId := args[0]

			// Acquire the client context
			clientCtx, err := client.GetClientTxContext(cmd)
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
		Use:   "claim [id]",
		Short: "Claim a given beam",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			argsId := args[0]

			// Acquire the client context
			clientCtx, err := client.GetClientTxContext(cmd)
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
