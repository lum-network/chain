package cli

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/version"

	gogotypes "github.com/cosmos/gogoproto/types"

	"github.com/lum-network/chain/x/millions/types"
)

func GetTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      fmt.Sprintf("%s transaction subcommands", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		CmdTxDeposit(),
		CmdTxDepositRetry(),
		CmdTxClaimPrize(),
		CmdTxDrawRetry(),
		CmdTxWithdrawDeposit(),
		CmdTxWithdrawDepositRetry(),
		CmdTxRestoreInterchainAccounts(),
		CmdTxDepositEdit(),
	)
	return cmd
}

func CmdTxDeposit() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "deposit <pool_id> <amount>",
		Short: "Deposit funds into a millions pool",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Deposit funds into a millions pool.

Examples:
To create a classic deposit
$ %s tx %s deposit <pool_id> <amount>

To create a delegated deposit (delegate drawing chances to another address)
$ %s tx %s deposit <pool_id> <amount> --winner_address=<address>

To create a sponsorship deposit (no drawing chances at all)
$ %s tx %s deposit <pool_id> <amount> --sponsor=true`,
				version.AppName, types.ModuleName, version.AppName, types.ModuleName, version.AppName, types.ModuleName),
		),
		Args: cobra.ExactArgs(2),
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
			poolId, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return err
			}
			amount, err := sdk.ParseCoinNormalized(args[1])
			if err != nil {
				return err
			}

			// Acquire optional arguments
			winnerAddress, err := cmd.Flags().GetString("winner_address")
			if err != nil {
				return err
			}
			isSponsor, err := cmd.Flags().GetBool("sponsor")
			if err != nil {
				return err
			}

			// Build the message
			msg := types.NewMsgDeposit(clientCtx.GetFromAddress().String(), amount, poolId)
			msg.WinnerAddress = winnerAddress
			msg.IsSponsor = isSponsor

			// Generate the transaction
			return tx.GenerateOrBroadcastTxWithFactory(clientCtx, txf, msg)
		},
	}
	cmd.Flags().String("winner_address", "", "(optional) winner address to direct the draw prizes to")
	cmd.Flags().Bool("sponsor", false, "(optional) active sponsor mode and waive this deposit draw chances")
	flags.AddTxFlagsToCmd(cmd)
	_ = cmd.MarkFlagRequired(flags.FlagFrom)
	return cmd
}

func CmdTxDepositRetry() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "deposit-retry <pool_id> <deposit_id>",
		Short: "Retry a failed deposit",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Retry a deposit stuck in a faulty state (ex: interchain tx issues).

Example:
$ %s tx %s deposit-retry <pool_id> <deposit_id>`,
				version.AppName, types.ModuleName),
		),
		Args: cobra.ExactArgs(2),
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
			poolID, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return err
			}

			depositID, err := strconv.ParseUint(args[1], 10, 64)
			if err != nil {
				return err
			}

			// Build the message
			msg := types.NewMsgDepositRetry(clientCtx.GetFromAddress().String(), poolID, depositID)

			// Generate the transaction
			return tx.GenerateOrBroadcastTxWithFactory(clientCtx, txf, msg)
		},
	}
	flags.AddTxFlagsToCmd(cmd)
	_ = cmd.MarkFlagRequired(flags.FlagFrom)
	return cmd
}

func CmdTxDepositEdit() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "deposit-edit <pool_id> <deposit_id> [winner_address] [sponsor]",
		Short: "Edit a deposit",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Edit a deposit's winnerAddress to redirect the draw prize and/or sponsor mode that waives this deposit draw chances.
Example:
To edit a delegated deposit (delegate drawing chances to another address)
$ %s tx %s deposit-edit <pool_id> <deposit_id> --winner_address=<address>
To edit a sponsorship deposit (no drawing chances at all)
$ %s tx %s deposit-edit <pool_id> <deposit_id> --sponsor=true
To edit a delegated deposit and sponsorship
$ %s tx %s deposit-edit <pool_id> <deposit_id> --winner_address=<address> --sponsor=true`,
				version.AppName, types.ModuleName, version.AppName, types.ModuleName, version.AppName, types.ModuleName),
		),
		Args: cobra.ExactArgs(2),
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
			poolID, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return err
			}

			depositID, err := strconv.ParseUint(args[1], 10, 64)
			if err != nil {
				return err
			}

			// Acquire optional arguments
			winnerAddress, err := cmd.Flags().GetString("winner_address")
			if err != nil {
				return err
			}

			isSponsor, err := cmd.Flags().GetString("sponsor")
			if err != nil {
				return err
			}

			// Parse the boolean value from the string
			var isSponsorWrapper *gogotypes.BoolValue
			if isSponsor != "" {
				isSponsorBool, err := strconv.ParseBool(isSponsor)
				if err != nil {
					return err
				}
				isSponsorWrapper = &gogotypes.BoolValue{
					Value: isSponsorBool,
				}
			}

			// Build the message
			msg := types.NewMsgDepositEdit(clientCtx.GetFromAddress().String(), poolID, depositID)
			msg.WinnerAddress = winnerAddress
			msg.IsSponsor = isSponsorWrapper

			// Generate the transaction
			return tx.GenerateOrBroadcastTxWithFactory(clientCtx, txf, msg)
		},
	}

	cmd.Flags().String("winner_address", "", "(optional) winner address to direct the draw prizes to")
	cmd.Flags().String("sponsor", "", "(optional) active sponsor mode and waive this deposit draw chances")
	flags.AddTxFlagsToCmd(cmd)
	_ = cmd.MarkFlagRequired(flags.FlagFrom)
	return cmd
}

func CmdTxClaimPrize() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "claim-prize <pool_id> <draw_id> <prize_id>",
		Short: "Claim a millions prize",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Claim a millions prize to send the funds to the prize winner address.

Example:
$ %s tx %s claim-prize <pool_id> <draw_id> <prize_id>`,
				version.AppName, types.ModuleName),
		),
		Args: cobra.ExactArgs(3),
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
			poolID, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return err
			}
			drawID, err := strconv.ParseUint(args[1], 10, 64)
			if err != nil {
				return err
			}
			prizeID, err := strconv.ParseUint(args[2], 10, 64)
			if err != nil {
				return err
			}

			// Build the message
			msg := types.NewMsgMsgClaimPrize(clientCtx.GetFromAddress().String(), poolID, drawID, prizeID)

			// Generate the transaction
			return tx.GenerateOrBroadcastTxWithFactory(clientCtx, txf, msg)
		},
	}
	flags.AddTxFlagsToCmd(cmd)
	_ = cmd.MarkFlagRequired(flags.FlagFrom)
	return cmd
}

func CmdTxWithdrawDeposit() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "withdraw-deposit <pool_id> <deposit_id>",
		Short: "Withdraw a deposit",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Withdraw a deposit.
Launches the unbonding of the deposit and wait for its completion in order to send the funds to the specified withdrawal address.

Examples:
To withdraw a deposit to the depositor address
$ %s tx %s withdraw-deposit <pool_id> <deposit_id>

To withdraw a deposit to another address
$ %s tx %s withdraw-deposit <pool_id> <deposit_id> --to_address=<address>`,
				version.AppName, types.ModuleName, version.AppName, types.ModuleName),
		),
		Args: cobra.ExactArgs(2),
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
			poolID, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return err
			}

			depositID, err := strconv.ParseUint(args[1], 10, 64)
			if err != nil {
				return err
			}

			destAddr, err := cmd.Flags().GetString("to_address")
			if err != nil {
				return err
			}
			if destAddr == "" {
				destAddr = clientCtx.GetFromAddress().String()
			}

			// Build the message
			msg := types.NewMsgWithdrawDeposit(clientCtx.GetFromAddress().String(), destAddr, poolID, depositID)
			msg.ToAddress = destAddr

			// Generate the transaction
			return tx.GenerateOrBroadcastTxWithFactory(clientCtx, txf, msg)
		},
	}
	cmd.Flags().String("to_address", "", "(optional) address to send the deposit back")
	flags.AddTxFlagsToCmd(cmd)
	_ = cmd.MarkFlagRequired(flags.FlagFrom)
	return cmd
}

func CmdTxWithdrawDepositRetry() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "withdraw-deposit-retry <pool_id> <withdrawal_id>",
		Short: "Retry a failed withdrawal",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Retry a withdrawal stuck in a faulty state (ex: interchain tx issues).

Example:
$ %s tx %s withdraw-deposit-retry <pool_id> <withdrawal_id>`,
				version.AppName, types.ModuleName),
		),
		Args: cobra.ExactArgs(2),
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
			poolID, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return err
			}
			withdrawalID, err := strconv.ParseUint(args[1], 10, 64)
			if err != nil {
				return err
			}

			// Build the message
			msg := types.NewMsgWithdrawDepositRetry(clientCtx.GetFromAddress().String(), poolID, withdrawalID)

			// Generate the transaction
			return tx.GenerateOrBroadcastTxWithFactory(clientCtx, txf, msg)
		},
	}
	flags.AddTxFlagsToCmd(cmd)
	_ = cmd.MarkFlagRequired(flags.FlagFrom)
	return cmd
}

func CmdTxDrawRetry() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "draw-retry <pool_id> <draw_id>",
		Short: "Retry a failed draw",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Retry a draw stuck in a faulty state (ex: interchain tx issues).
Draws are only launched by the protocol itself but can be retried (if in a faulty state) by anyone.

Example:
$ %s tx %s draw-retry <pool_id> <draw_id>`,
				version.AppName, types.ModuleName),
		),
		Args: cobra.ExactArgs(2),
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
			poolID, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return err
			}

			drawID, err := strconv.ParseUint(args[1], 10, 64)
			if err != nil {
				return err
			}

			// Build the message
			msg := types.NewMsgDrawRetry(clientCtx.GetFromAddress().String(), poolID, drawID)

			// Generate the transaction
			return tx.GenerateOrBroadcastTxWithFactory(clientCtx, txf, msg)
		},
	}
	flags.AddTxFlagsToCmd(cmd)
	_ = cmd.MarkFlagRequired(flags.FlagFrom)
	return cmd
}

func CmdTxRestoreInterchainAccounts() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "restore-interchain-accounts <pool_id>",
		Short: "Try to restore the interchain accounts of a given pool after one closed",
		Args:  cobra.MinimumNArgs(1),
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
			poolId, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return err
			}

			// Build the message
			msg := types.NewMsgRestoreInterchainAccounts(clientCtx.GetFromAddress().String(), poolId)
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
