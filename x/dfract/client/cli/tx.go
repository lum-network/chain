package cli

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/spf13/cobra"

	"github.com/lum-network/chain/x/dfract/types"
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

	cmd.AddCommand(CmdDeposit(), CmdWithdrawAndMint())

	return cmd
}

func CmdDeposit() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "deposit <amount>",
		Short: "Deposit funds to the dFract module",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Acquire the client context
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			txf, err := tx.NewFactoryCLI(clientCtx, cmd.Flags())
			if err != nil {
				return err
			}
			txf = txf.WithTxConfig(clientCtx.TxConfig).WithAccountRetriever(clientCtx.AccountRetriever)

			// Acquire the command arguments
			argsAmount, err := sdk.ParseCoinNormalized(args[0])
			if err != nil {
				return err
			}

			// Build the message
			msg := types.NewMsgDeposit(clientCtx.GetFromAddress().String(), argsAmount)
			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			// Generate the transaction
			return tx.GenerateOrBroadcastTxWithFactory(clientCtx, txf, msg)
		},
	}
	flags.AddTxFlagsToCmd(cmd)
	_ = cmd.MarkFlagRequired(flags.FlagFrom)
	return cmd
}

func CmdWithdrawAndMint() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "withdraw-and-mint <micro_mint_rate>",
		Short: "Withdraw and mint udfr",
		Long: strings.TrimSpace(
			fmt.Sprintf(`The withdrawal address specified in the dFract module parameters is the one authorized to withdraw and mint udfr tokens based on the micro mint rate.

Examples:
To create a withdraw-and-mint tx
$ %s tx %s withdraw-and-mint <micro_min_rate>
`,
				version.AppName, types.ModuleName),
		),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Acquire the client context
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			txf, err := tx.NewFactoryCLI(clientCtx, cmd.Flags())
			if err != nil {
				return err
			}
			txf = txf.WithTxConfig(clientCtx.TxConfig).WithAccountRetriever(clientCtx.AccountRetriever)

			microMintRate, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return err
			}

			// Build the message
			msg := types.NewMsgWithdrawAndMint(clientCtx.GetFromAddress().String(), microMintRate)

			// Generate the transaction
			return tx.GenerateOrBroadcastTxWithFactory(clientCtx, txf, msg)
		},
	}
	flags.AddTxFlagsToCmd(cmd)
	_ = cmd.MarkFlagRequired(flags.FlagFrom)
	return cmd
}
