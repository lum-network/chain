package cli

import (
	"strconv"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/gov/client/cli"
	govv1betatypes "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	"github.com/lum-network/chain/x/dfract/types"
	"github.com/spf13/cobra"
)

func NewSubmitWithdrawAndMintProposal() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "withdraw-and-mint [withdrawal-ddress] [micro-mint-rate] [flags]",
		Args:  cobra.ExactArgs(2),
		Short: "Submit a withdraw and mint proposal",
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			// Extract title and description
			title, err := cmd.Flags().GetString(cli.FlagTitle)
			if err != nil {
				return err
			}
			description, err := cmd.Flags().GetString(cli.FlagDescription)
			if err != nil {
				return err
			}

			// Extract the deposit value
			depositStr, err := cmd.Flags().GetString(cli.FlagDeposit)
			if err != nil {
				return err
			}

			deposit, err := sdk.ParseCoinsNormalized(depositStr)
			if err != nil {
				return err
			}

			// Extract the from address
			from := clientCtx.GetFromAddress()

			// Extract the other parameters
			destinationAddress := args[0]
			microMintRate, err := strconv.ParseInt(args[1], 10, 64)
			if err != nil {
				return err
			}

			content := types.NewWithdrawAndMintProposal(title, description, destinationAddress, microMintRate)
			msg, err := govv1betatypes.NewMsgSubmitProposal(content, deposit, from)
			if err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}
	cmd.Flags().String(cli.FlagTitle, "", "title of proposal")
	cmd.Flags().String(cli.FlagDescription, "", "description of proposal")
	cmd.Flags().String(cli.FlagDeposit, "", "deposit of proposal")

	_ = cmd.MarkFlagRequired(cli.FlagTitle)
	_ = cmd.MarkFlagRequired(cli.FlagDescription)
	return cmd
}
