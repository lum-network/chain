package cli

import (
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/gov/client/cli"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/lum-network/chain/x/dfract/types"
	"github.com/spf13/cobra"
	"strconv"
)

func NewSubmitSpendAndAdjustProposal() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "spend-and-adjust [destinationAddress] [mintInfo] [flags]",
		Args:  cobra.ExactArgs(2),
		Short: "Submit a spend and adjust proposal",
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
			mintRate, err := strconv.ParseInt(args[1], 10, 64)
			if err != nil {
				return err
			}

			content := types.NewSpendAndAdjustProposal(title, description, destinationAddress, mintRate)
			msg, err := govtypes.NewMsgSubmitProposal(content, deposit, from)
			if err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}
	cmd.Flags().String(cli.FlagTitle, "", "title of proposal")
	cmd.Flags().String(cli.FlagDescription, "", "description of proposal")
	cmd.Flags().String(cli.FlagDeposit, "", "deposit of proposal")

	cmd.MarkFlagRequired(cli.FlagTitle)
	cmd.MarkFlagRequired(cli.FlagDescription)
	return cmd
}
