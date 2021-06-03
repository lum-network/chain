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
		Use:   "open <amount> <secret> <schema>",
		Short: "Open a new beam",
		Args:  cobra.ExactArgs(3),
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

			// Try to acquire the reward flag
			argsReward, err := cmd.Flags().GetString(FlagReward)
			if err != nil {
				return err
			}
			var rew types.BeamSchemeReward
			if len(argsReward) > 0 {
				if err = json.Unmarshal([]byte(argsReward), &rew); err != nil {
					return err
				}
			}

			// Trying to acquire the review flag
			argsReview, err := cmd.Flags().GetString(FlagReview)
			if err != nil {
				return err
			}
			var rev types.BeamSchemeReview
			if len(argsReview) > 0 {
				if err = json.Unmarshal([]byte(argsReview), &rev); err != nil {
					return err
				}
			}

			// Trying to acquire the owner flag
			argsOwner, err := cmd.Flags().GetString(FlagOwner)
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
			msg := types.NewMsgOpenBeam(id, clientCtx.GetFromAddress().String(), argsOwner, &coin, hashedSecret, argsSchema, &rew, &rev)
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
		Use:   "update <id> <amount> [reward] [review]",
		Short: "Update a given beam",
		Args:  cobra.RangeArgs(2, 4),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Acquire the command arguments
			argsId := args[0]

			// Trying to acquire the amount
			coin, err := sdk.ParseCoinNormalized(args[1])
			if err != nil {
				return err
			}

			// Try to acquire the reward arg
			argsReward, err := cmd.Flags().GetString(FlagReward)
			if err != nil {
				return err
			}
			var rew types.BeamSchemeReward
			if len(argsReward) > 0 {
				if err = json.Unmarshal([]byte(argsReward), &rew); err != nil {
					return err
				}
			}

			// Trying to acquire the review arg
			argsReview, err := cmd.Flags().GetString(FlagReview)
			if err != nil {
				return err
			}
			var rev types.BeamSchemeReview
			if len(argsReview) > 0 {
				if err = json.Unmarshal([]byte(argsReview), &rev); err != nil {
					return err
				}
			}

			argsStatus, err := cmd.Flags().GetInt32(FlagStatus)
			if err != nil {
				return err
			}

			// Acquire the client context
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			// Construct the message and validate
			msg := types.NewMsgUpdateBeam(clientCtx.GetFromAddress().String(), argsId, &coin, types.BeamState(argsStatus), &rew, &rev)
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
		Use:   "claim [id]",
		Short: "Claim a given beam",
		Args:  cobra.ExactArgs(1),
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
