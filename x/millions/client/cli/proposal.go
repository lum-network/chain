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
	govcli "github.com/cosmos/cosmos-sdk/x/gov/client/cli"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	"github.com/lum-network/chain/x/millions/types"
	"github.com/spf13/cobra"
)

func parseRegisterPoolProposalFile(cdc codec.JSONCodec, proposalFile string) (proposal types.ProposalRegisterPool, err error) {
	contents, err := os.ReadFile(proposalFile)
	if err != nil {
		return proposal, err
	}

	if err = cdc.UnmarshalJSON(contents, &proposal); err != nil {
		return proposal, err
	}
	return proposal, nil
}

func parseUpdatePoolProposalFile(cdc codec.JSONCodec, proposalFile string) (proposal types.ProposalUpdatePool, err error) {
	contents, err := os.ReadFile(proposalFile)
	if err != nil {
		return proposal, err
	}

	if err = cdc.UnmarshalJSON(contents, &proposal); err != nil {
		return proposal, err
	}
	return proposal, nil
}

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

func parseDisableValidatorProposalFile(cdc codec.JSONCodec, proposalFile string) (proposal types.ProposalDisableValidator, err error) {
	contents, err := os.ReadFile(proposalFile)
	if err != nil {
		return proposal, err
	}

	if err = cdc.UnmarshalJSON(contents, &proposal); err != nil {
		return proposal, err
	}
	return proposal, nil
}

func CmdProposalRegisterPool() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "millions-register-pool [proposal-file]",
		Short: "Submit a millions register pool proposal",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Submit a RegisterPool proposal along with an initial deposit.
The proposal details must be supplied via a JSON file.

Example:
$ %s tx gov submit-legacy-proposal millions-register-pool <path/to/proposal.json> --from=<key_or_address>

Where proposal.json contains:
{
    "title": "My new pool",
    "description": "This is my new pool",
    "chain_id": "lumnetwork-testnet",
    "denom": "ulum",
    "native_denom": "ulum",
    "connection_id": "",
    "validators": ["lumvaloper1wf6alkrpjn4zhcnag3afqz34mlanplzwx6v8qz"],
    "min_deposit_amount": "1000000",
    "draw_schedule": {
        "draw_delta": "3600s",
        "initial_draw_at": "2023-04-19T00:23:41.242670441Z" 
    },
    "prize_strategy": {
        "prize_batches": [{
            "draw_probability": "1.000000000000000000",
            "pool_percent": "100",
            "quantity": "100"
        }]
    },
    "bech32_prefix_acc_addr": "lum",
    "bech32_prefix_val_addr": "lumvaloper",
    "transfer_channel_id": ""
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
			proposal, err := parseRegisterPoolProposalFile(clientCtx.Codec, args[0])
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

func CmdProposalUpdatePool() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "millions-update-pool [proposal-file]",
		Short: "Submit a millions update pool proposal",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Submit an UpdatePool proposal along with an initial deposit.
The proposal details must be supplied via a JSON file.

Example:
$ %s tx gov submit-legacy-proposal millions-update-pool <path/to/proposal.json> --from=<key_or_address>

Where proposal.json contains:
{
    "title": "Update my pool",
    "description": "This is my updated pool",
    "pool_id": 1,
    "validators": ["lumvaloper1wf6alkrpjn4zhcnag3afqz34mlanplzwx6v8qz"],
    "min_deposit_amount": "2000000",
    "draw_schedule": {
        "draw_delta": "3600s",
        "initial_draw_at": "2023-04-19T00:23:41.242670441Z" 
    },
    "prize_strategy": {
        "prize_batches": [{
            "draw_probability": "1.000000000000000000",
            "pool_percent": "100",
            "quantity": "100"
        }]
    }
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
			proposal, err := parseUpdatePoolProposalFile(clientCtx.Codec, args[0])
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

func CmdProposalUpdateParams() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "millions-update-params [proposal-file]",
		Short: "Submit a millions update params proposal",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Submit an UpdateParams proposal along with an initial deposit.
The proposal details must be supplied via a JSON file.

Example:
$ %s tx gov submit-legacy-proposal millions-update-params <path/to/proposal.json> --from=<key_or_address>

Where proposal.json contains:
{
    "title": "Update my params",
    "description": "This is my updated params",
    "min_deposit_amount": "1",
    "min_deposit_draw_delta": "60s",
    "max_prize_strategy_batches": "100",
    "max_prize_batch_quantity": "1000",
    "min_draw_schedule_delta": "3600s",
    "max_draw_schedule_delta": "31622400s",
    "prize_expiration_delta": "2592000s",
    "fees_stakers": "0.000000000000000000"
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

func CmdProposalDisableValidator() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "millions-disable-validator [proposal-file]",
		Short: "Submit a millions disable validator",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Submit an DisableValidator proposal along with an initial deposit.
The proposal details must be supplied via a JSON file.

Example:
$ %s tx gov submit-legacy-proposal millions-disable-validator <path/to/proposal.json> --from=<key_or_address>

Where proposal.json contains:
{
    "title": "Disable validator",
    "description": "This is the reason why we want to disable this validator",
	"pool_id": "1",
    "operator_address": "lumvaloper12x88yexvf6qexfjg9czp6jhpv7vpjdwwedhzhk"
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
			proposal, err := parseDisableValidatorProposalFile(clientCtx.Codec, args[0])
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
