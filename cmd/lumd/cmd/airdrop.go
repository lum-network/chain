package cmd

import (
	"encoding/json"
	"fmt"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/server"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/bech32"
	"github.com/spf13/cobra"
	"io/ioutil"
	"os"

	v038genaccounts "github.com/cosmos/cosmos-sdk/x/auth/legacy/v038"
	v038distribution "github.com/cosmos/cosmos-sdk/x/distribution/legacy/v038"
	v038staking "github.com/cosmos/cosmos-sdk/x/staking/legacy/v038"
)

// AppStateV038 is app state structure for app state
type AppStateV038 struct {
	Accounts     []v038genaccounts.GenesisAccount `json:"accounts"`
	Staking      v038staking.GenesisState         `json:"staking"`
	Distribution v038distribution.GenesisState    `json:"distribution"`
}

// GenesisStateV038 is minimum structure to import airdrop accounts
type GenesisStateV038 struct {
	AppState AppStateV038 `json:"app_state"`
}

type Snapshot struct {
	TotalAtomAmount       sdk.Int `json:"total_atom_amount"`
	TotalLumAirdropAmount sdk.Int `json:"total_lum_amount"`
	NumberAccounts        uint64  `json:"num_accounts"`

	Accounts map[string]SnapshotAccount `json:"accounts"`
}

// SnapshotAccount provide fields of snapshot per account
type SnapshotAccount struct {
	AtomAddress string `json:"atom_address"` // Atom Balance = AtomStakedBalance + AtomUnstakedBalance

	AtomBalance          sdk.Int `json:"atom_balance"`
	AtomOwnershipPercent sdk.Dec `json:"atom_ownership_percent"`

	AtomStakedBalance   sdk.Int `json:"atom_staked_balance"`
	AtomUnstakedBalance sdk.Int `json:"atom_unstaked_balance"` // AtomStakedPercent = AtomStakedBalance / AtomBalance
	AtomStakedPercent   sdk.Dec `json:"atom_staked_percent"`

	LumBalance      sdk.Int `json:"lum_balance"`           // LumBalance = sqrt( AtomBalance ) * (1 + 1.5 * atom staked percent)
	LumBalanceBase  sdk.Int `json:"lum_balance_base"`      // LumBalanceBase = sqrt(atom balance)
	LumBalanceBonus sdk.Int `json:"lum_balance_bonus"`     // LumBalanceBonus = LumBalanceBase * (1.5 * atom staked percent)
	LumPercent      sdk.Dec `json:"lum_ownership_percent"` // LumPercent = LumNormalizedBalance / TotalLumSupply
}

// setCosmosBech32Prefixes set config for cosmos address system
func setCosmosBech32Prefixes() {
	defaultConfig := sdk.NewConfig()
	config := sdk.GetConfig()
	config.SetBech32PrefixForAccount(defaultConfig.GetBech32AccountAddrPrefix(), defaultConfig.GetBech32AccountPubPrefix())
	config.SetBech32PrefixForValidator(defaultConfig.GetBech32ValidatorAddrPrefix(), defaultConfig.GetBech32ValidatorPubPrefix())
	config.SetBech32PrefixForConsensusNode(defaultConfig.GetBech32ConsensusAddrPrefix(), defaultConfig.GetBech32ConsensusPubPrefix())
}

func ExportAirdropSnapshotCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "export-airdrop-snapshot [airdrop-to-denom] [input-genesis-file] [output-snapshot-json] --lum-supply=[lum-genesis-supply]",
		Short: "Export a quadratic fairdrop snapshot from a provided cosmos-sdk v0.36 genesis export",
		Long: `Export a quadratic fairdrop snapshot from a provided cosmos-sdk v0.36 genesis export
Sample genesis file:
	https://raw.githubusercontent.com/cephalopodequipment/cosmoshub-3/master/genesis.json
Example:
	lumd export-airdrop-genesis uatom ~/.lumd/config/genesis.json ../snapshot.json --lum-supply=100000000000000
	- Check input genesis:
		file is at ~/.lumd/config/genesis.json
	- Snapshot
		file is at "../snapshot.json"
`,
		Args: cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			aminoCodec := clientCtx.LegacyAmino.Amino

			serverCtx := server.GetServerContextFromCmd(cmd)
			config := serverCtx.Config

			config.SetRoot(clientCtx.HomeDir)

			denom := args[0]
			genesisFile := args[1]
			snapshotOutput := args[2]

			// Read genesis file
			genesisJson, err := os.Open(genesisFile)
			if err != nil {
				return err
			}
			defer genesisJson.Close()

			byteValue, _ := ioutil.ReadAll(genesisJson)

			var genStateV036 GenesisStateV038

			setCosmosBech32Prefixes()
			err = aminoCodec.UnmarshalJSON(byteValue, &genStateV036)
			if err != nil {
				return err
			}

			// Produce the map of address to total atom balance, both staked and unstaked
			snapshotAccs := make(map[string]SnapshotAccount)

			totalAtomBalance := sdk.NewInt(0)
			for _, account := range genStateV036.AppState.Accounts {

				balance := account.GetCoins().AmountOf(denom)
				totalAtomBalance = totalAtomBalance.Add(balance)

				snapshotAccs[account.GetAddress().String()] = SnapshotAccount{
					AtomAddress:         account.GetAddress().String(),
					AtomBalance:         balance,
					AtomUnstakedBalance: balance,
					AtomStakedBalance:   sdk.ZeroInt(),
				}
			}

			for _, unbonding := range genStateV036.AppState.Staking.UnbondingDelegations {
				address := unbonding.DelegatorAddress.String()
				acc, ok := snapshotAccs[address]
				if !ok {
					panic("no account found for unbonding")
				}

				unbondingAtoms := sdk.NewInt(0)
				for _, entry := range unbonding.Entries {
					unbondingAtoms = unbondingAtoms.Add(entry.Balance)
				}

				acc.AtomBalance = acc.AtomBalance.Add(unbondingAtoms)
				acc.AtomUnstakedBalance = acc.AtomUnstakedBalance.Add(unbondingAtoms)

				snapshotAccs[address] = acc
			}

			// Make a map from validator operator address to the v036 validator type
			validators := make(map[string]v038staking.Validator)
			for _, validator := range genStateV036.AppState.Staking.Validators {
				validators[validator.OperatorAddress.String()] = validator
			}

			for _, delegation := range genStateV036.AppState.Staking.Delegations {
				address := delegation.DelegatorAddress.String()

				acc, ok := snapshotAccs[address]
				if !ok {
					panic("no account found for delegation")
				}

				val := validators[delegation.ValidatorAddress.String()]
				stakedAtoms := delegation.Shares.MulInt(val.Tokens).Quo(val.DelegatorShares).RoundInt()

				acc.AtomBalance = acc.AtomBalance.Add(stakedAtoms)
				acc.AtomStakedBalance = acc.AtomStakedBalance.Add(stakedAtoms)

				snapshotAccs[address] = acc
			}

			totalLumBalance := sdk.NewInt(0)
			onePointFive := sdk.MustNewDecFromStr("1.5")

			for address, acc := range snapshotAccs {
				allAtoms := acc.AtomBalance.ToDec()

				acc.AtomOwnershipPercent = allAtoms.QuoInt(totalAtomBalance)

				if allAtoms.IsZero() {
					acc.AtomStakedPercent = sdk.ZeroDec()
					acc.LumBalanceBase = sdk.ZeroInt()
					acc.LumBalanceBonus = sdk.ZeroInt()
					acc.LumBalance = sdk.ZeroInt()
					snapshotAccs[address] = acc
					continue
				}

				stakedAtoms := acc.AtomStakedBalance.ToDec()
				stakedPercent := stakedAtoms.Quo(allAtoms)
				acc.AtomStakedPercent = stakedPercent

				baseLum, err := allAtoms.ApproxSqrt()
				if err != nil {
					panic(fmt.Sprintf("failed to root atom balance: %s", err))
				}
				acc.LumBalanceBase = baseLum.RoundInt()

				bonusLum := baseLum.Mul(onePointFive).Mul(stakedPercent)
				acc.LumBalanceBonus = bonusLum.RoundInt()

				allLum := baseLum.Add(bonusLum)
				// LumBalance = sqrt( all atoms) * (1 + 1.5) * (staked atom percent) =
				acc.LumBalance = allLum.RoundInt()

				if allAtoms.LTE(sdk.NewDec(1000000)) {
					acc.LumBalanceBase = sdk.ZeroInt()
					acc.LumBalanceBonus = sdk.ZeroInt()
					acc.LumBalance = sdk.ZeroInt()
				}

				totalLumBalance = totalLumBalance.Add(acc.LumBalance)

				snapshotAccs[address] = acc
			}

			// iterate to find Osmo ownership percentage per account
			for address, acc := range snapshotAccs {
				acc.LumPercent = acc.LumBalance.ToDec().Quo(totalLumBalance.ToDec())
				snapshotAccs[address] = acc
			}

			snapshot := Snapshot{
				TotalAtomAmount:       totalAtomBalance,
				TotalLumAirdropAmount: totalLumBalance,
				NumberAccounts:        uint64(len(snapshotAccs)),
				Accounts:              snapshotAccs,
			}

			fmt.Printf("# accounts: %d\n", len(snapshotAccs))
			fmt.Printf("atomTotalSupply: %s\n", totalAtomBalance.String())
			fmt.Printf("lumTotalSupply: %s\n", totalLumBalance.String())

			// export snapshot json
			snapshotJSON, err := json.MarshalIndent(snapshot, "", "    ")
			if err != nil {
				return fmt.Errorf("failed to marshal snapshot: %w", err)
			}

			err = ioutil.WriteFile(snapshotOutput, snapshotJSON, 0644)
			return err
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// ConvertCosmosAddressToLum convert cosmos1 address to juno1 address
func ConvertCosmosAddressToLum(address string) (sdk.AccAddress, error) {

	config := sdk.GetConfig()
	lumPrefix := config.GetBech32AccountAddrPrefix()

	_, bytes, err := bech32.DecodeAndConvert(address)
	if err != nil {
		return nil, err
	}

	newAddr, err := bech32.ConvertAndEncode(lumPrefix, bytes)
	if err != nil {
		return nil, err
	}

	sdkAddr, err := sdk.AccAddressFromBech32(newAddr)
	if err != nil {
		return nil, err
	}

	return sdkAddr, nil
}
