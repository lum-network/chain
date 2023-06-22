package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/version"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	"github.com/spf13/cobra"

	govcli "github.com/cosmos/cosmos-sdk/x/gov/client/cli"
	"github.com/lum-network/chain/x/dfract/types"
)

func parseUpdateParamsProposalFile(cdc codec.JSONCodec, proposalFile string) (proposal types.ProposalUpdateParams, err error) {
	contents, err := os.ReadFile(proposalFile)
	if err != nil {
		return proposal, err
	}

	if err = cdc.UnmarshalJSON(contents, &proposal); err != nil {
		return proposal, err
	}
	return proposal, nil
}

func CmdProposalUpdateParams() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "dfract-update-params [proposal-file]",
		Short: "Submit a dfract update params proposal",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Submit an UpdateParams proposal along with an initial deposit.
The proposal details must be supplied via a JSON file.

Example:
$ %s tx gov submit-legacy-proposal dfract-update-params <path/to/proposal.json> --from=<key_or_address>

Where proposal.json contains:
{
    "title": "Update dfract params",
    "description": "Update management address and deposit enablement",
    "management_address": "lum1qx2dts3tglxcu0jh47k7ghstsn4nactukljgyj",
	"is_deposit_enabled": false
}
`, version.AppName),
		),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Acquire the client context
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			// Parse the proposal file
			proposal, err := parseUpdateParamsProposalFile(clientCtx.Codec, args[0])
			if err != nil {
				return err
			}

			if err := proposal.ValidateBasic(); err != nil {
				return err
			}

			// Grab the parameters
			from := clientCtx.GetFromAddress()

			// Grab the deposit
			depositStr, err := cmd.Flags().GetString(govcli.FlagDeposit)
			if err != nil {
				return err
			}

			deposit, err := sdk.ParseCoinsNormalized(depositStr)
			if err != nil {
				return err
			}

			msg, err := govtypes.NewMsgSubmitProposal(&proposal, deposit, from)
			if err != nil {
				return err
			}

			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			// Generate the transaction
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	cmd.Flags().String(govcli.FlagDeposit, "1ulum", "deposit of proposal")
	if err := cmd.MarkFlagRequired(govcli.FlagDeposit); err != nil {
		panic(err)
	}
	return cmd
}
