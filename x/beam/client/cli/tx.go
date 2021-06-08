package cli

import (
	"encoding/json"
	"fmt"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
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
		CmdClaimBeam(),
	)
	return cmd
}

// CmdOpenBeam Command definition for beam opening dispatch
func CmdOpenBeam() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "open <amount> <secret> <schema> [data] [closes-at-block] [claim-expires-at-block]",
		Short: "Open a new beam",
		Args:  cobra.RangeArgs(3, 6),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Trying to acquire the amount
			coin, err := sdk.ParseCoinNormalized(args[0])
			if err != nil {
				return err
			}

			// Acquire the command arguments
			argsSecret := args[1]
			argsSchema := args[2]
			if err != nil {
				return err
			}

			// Acquire the client context
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			// Try to acquire the data arg
			argsData, err := cmd.Flags().GetString(FlagData)
			if err != nil {
				return err
			}
			var data *types.BeamData
			if len(argsData) > 0 && argsData != "" {
				if err = json.Unmarshal([]byte(argsData), &data); err != nil {
					return err
				}
			}

			// Trying to acquire the owner flag
			argsOwner, err := cmd.Flags().GetString(FlagOwner)
			if err != nil {
				return err
			}

			argsClosesAtBlock, err := cmd.Flags().GetInt32(FlagClosesAtBlock)
			if err != nil {
				return err
			}

			argsClaimExpiresAtBlock, err := cmd.Flags().GetInt32(FlagClaimExpiresAtBlock)
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
			msg := types.NewMsgOpenBeam(id, clientCtx.GetFromAddress().String(), argsOwner, coin, hashedSecret, argsSchema, data, argsClosesAtBlock, argsClaimExpiresAtBlock)
			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			// Broadcast the message
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}
	cmd.Flags().AddFlagSet(flagSetBeamMetadata())
	cmd.Flags().AddFlagSet(flagSetOwner())
	flags.AddTxFlagsToCmd(cmd)
	_ = cmd.MarkFlagRequired(flags.FlagFrom)
	return cmd
}

// CmdUpdateBeam Command definition for beam increase dispatch
func CmdUpdateBeam() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id> <amount> [data] [cancel-reason] [hide-content] [closes-at-block] [claim-expires-at-block]",
		Short: "Update a given beam",
		Args:  cobra.RangeArgs(2, 5),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Acquire the command arguments
			argsId := args[0]

			// Trying to acquire the amount
			coin, err := sdk.ParseCoinNormalized(args[1])
			if err != nil {
				return err
			}

			// Try to acquire the data arg
			argsData, err := cmd.Flags().GetString(FlagData)
			if err != nil {
				return err
			}
			var data *types.BeamData
			if len(argsData) > 0 && argsData != "" {
				if err = json.Unmarshal([]byte(argsData), &data); err != nil {
					return err
				}
			}

			// Try to acquire the cancel reason
			argsCancelReason, err := cmd.Flags().GetString(FlagCancelReason)
			if err != nil {
				return err
			}

			// Try to acquire the hide content param
			argsHideContent, err := cmd.Flags().GetBool(FlagHideContent)
			if err != nil {
				return err
			}

			argsStatus, err := cmd.Flags().GetInt32(FlagStatus)
			if err != nil {
				return err
			}

			argsClosesAtBlock, err := cmd.Flags().GetInt32(FlagClosesAtBlock)
			if err != nil {
				return err
			}

			argsClaimExpiresAtBlock, err := cmd.Flags().GetInt32(FlagClaimExpiresAtBlock)
			if err != nil {
				return err
			}

			// Acquire the client context
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			// Construct the message and validate
			msg := types.NewMsgUpdateBeam(clientCtx.GetFromAddress().String(), argsId, coin, types.BeamState(argsStatus), data, argsCancelReason, argsHideContent, argsClosesAtBlock, argsClaimExpiresAtBlock)
			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}
	cmd.Flags().AddFlagSet(flagSetBeamMetadata())
	flags.AddTxFlagsToCmd(cmd)
	_ = cmd.MarkFlagRequired(flags.FlagFrom)
	return cmd
}

// CmdClaimBeam Command definition for beam claim dispatch
func CmdClaimBeam() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "claim <id> <secret>",
		Short: "Claim a given beam",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			argsId := args[0]
			argsSecret := args[1]

			// Acquire the client context
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			// Construct the message and validate
			msg := types.NewMsgClaimBeam(clientCtx.GetFromAddress().String(), argsId, argsSecret)
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
