package simulation

import (
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/simapp/helpers"
	"github.com/cosmos/cosmos-sdk/simapp/params"
	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	account "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	bank "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	"github.com/cosmos/cosmos-sdk/x/simulation"
	"github.com/lum-network/chain/utils"
	"github.com/lum-network/chain/x/beam/keeper"
	"github.com/lum-network/chain/x/beam/types"
	"math/rand"
)

const (
	OpWeightMsgOpenBeam   = "op_weight_msg_open_beam"
	OpWeightMsgUpdateBeam = "op_weight_msg_update_beam"
	OpWeightMsgClaimBeam  = "op_weight_msg_claim_beam"
)

func WeightedOperations(appParams simtypes.AppParams, cdc codec.JSONCodec, ak account.AccountKeeper,
	bk bank.Keeper, k keeper.Keeper, wContents []simtypes.WeightedProposalContent) simulation.WeightedOperations {
	var (
		weightMsgOpenBeam   int
		weightMsgUpdateBeam int
		weightMsgClaimBeam  int
	)

	appParams.GetOrGenerate(cdc, OpWeightMsgOpenBeam, &weightMsgOpenBeam, nil, func(_ *rand.Rand) {
		weightMsgOpenBeam = 50
	})

	appParams.GetOrGenerate(cdc, OpWeightMsgUpdateBeam, &weightMsgUpdateBeam, nil, func(_ *rand.Rand) {
		weightMsgUpdateBeam = 100
	})

	appParams.GetOrGenerate(cdc, OpWeightMsgClaimBeam, &weightMsgClaimBeam, nil, func(_ *rand.Rand) {
		weightMsgClaimBeam = 30
	})

	return simulation.WeightedOperations{
		simulation.NewWeightedOperation(weightMsgOpenBeam, SimulateMsgOpenBeam(k, ak, bk)),
		simulation.NewWeightedOperation(weightMsgUpdateBeam, SimulateMsgUpdateBeam(k, ak, bk)),
		simulation.NewWeightedOperation(weightMsgClaimBeam, SimulateMsgClaimBeam(k, ak, bk)),
	}
}

func operationSimulateMsgClaimBeam(k keeper.Keeper, ak account.AccountKeeper, bk bank.Keeper, beamID string) simtypes.Operation {
	return func(r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		simAccount, _ := simtypes.RandomAcc(r, accs)

		if len(beamID) <= 0 {
			beamID = utils.GenerateSecureToken(8)
		}

		msg := types.NewMsgClaimBeam(simAccount.Address.String(), beamID, "")

		account := ak.GetAccount(ctx, simAccount.Address)
		spendable := bk.SpendableCoins(ctx, account.GetAddress())

		fees, err := simtypes.RandomFees(r, ctx, spendable)
		if err != nil {
			return simtypes.NoOpMsg(types.ModuleName, msg.Type(), "unable to generate fees"), nil, err
		}

		txGenerator := params.MakeTestEncodingConfig().TxConfig
		tx, err := helpers.GenSignedMockTx(
			r,
			txGenerator,
			[]sdk.Msg{msg},
			fees,
			helpers.DefaultGenTxGas,
			chainID,
			[]uint64{account.GetAccountNumber()}, []uint64{account.GetSequence()}, simAccount.PrivKey,
		)

		if err != nil {
			return simtypes.NoOpMsg(types.ModuleName, msg.Type(), "unable to generate mock tx"), nil, err
		}

		_, _, err = app.SimDeliver(txGenerator.TxEncoder(), tx)
		if err != nil {
			return simtypes.NoOpMsg(types.ModuleName, msg.Type(), "unable to deliver tx"), nil, err
		}

		return simtypes.NewOperationMsg(msg, true, "", nil), nil, nil
	}
}

func SimulateMsgClaimBeam(k keeper.Keeper, ak account.AccountKeeper, bk bank.Keeper) simtypes.Operation {
	return operationSimulateMsgClaimBeam(k, ak, bk, "")
}

func SimulateMsgOpenBeam(k keeper.Keeper, ak account.AccountKeeper, bk bank.Keeper) simtypes.Operation {
	return func(r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		// Generate fields
		msgId := utils.GenerateSecureToken(8)
		msgSecret := utils.GenerateSecureToken(10)
		msgSecretHashed := utils.GenerateHashFromString(msgSecret)
		from, to, coins, skip := randomSendFields(r, ctx, accs, bk, ak)

		// If anything failed, just skip
		if skip {
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgOpenBeam, "skip open beam"), nil, nil
		}

		msg := types.NewMsgOpenBeam(msgId, from.Address.String(), to.Address.String(), &coins[0], string(msgSecretHashed), "lum-network/reward", nil, 0, 0)
		err := sendMsgOpenBeam(r, app, bk, ak, msg, ctx, chainID, []cryptotypes.PrivKey{from.PrivKey})
		if err != nil {
			return simtypes.NoOpMsg(types.ModuleName, msg.Type(), "invalid transfers"), nil, err
		}

		return simtypes.NewOperationMsg(msg, true, "", nil), nil, nil
	}
}

func operationSimulateMsgUpdateBeam(k keeper.Keeper, ak account.AccountKeeper, bk bank.Keeper, beamID string) simtypes.Operation {
	return func(r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		simAccount, _ := simtypes.RandomAcc(r, accs)

		if len(beamID) <= 0 {
			beamID = utils.GenerateSecureToken(8)
		}

		_, _, coins, skip := randomSendFields(r, ctx, accs, bk, ak)
		if skip {
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgUpdateBeam, "skip update beam"), nil, nil
		}

		msg := types.NewMsgUpdateBeam(simAccount.Address.String(), beamID, &coins[0], types.BeamState_StateUnspecified, nil, "", false, 0, 0)

		account := ak.GetAccount(ctx, simAccount.Address)
		spendable := bk.SpendableCoins(ctx, account.GetAddress())

		fees, err := simtypes.RandomFees(r, ctx, spendable)
		if err != nil {
			return simtypes.NoOpMsg(types.ModuleName, msg.Type(), "unable to generate fees"), nil, err
		}

		txGenerator := params.MakeTestEncodingConfig().TxConfig
		tx, err := helpers.GenSignedMockTx(
			r,
			txGenerator,
			[]sdk.Msg{msg},
			fees,
			helpers.DefaultGenTxGas,
			chainID,
			[]uint64{account.GetAccountNumber()}, []uint64{account.GetSequence()}, simAccount.PrivKey,
		)

		if err != nil {
			return simtypes.NoOpMsg(types.ModuleName, msg.Type(), "unable to generate mock tx"), nil, err
		}

		_, _, err = app.SimDeliver(txGenerator.TxEncoder(), tx)
		if err != nil {
			return simtypes.NoOpMsg(types.ModuleName, msg.Type(), "unable to deliver tx"), nil, err
		}

		return simtypes.NewOperationMsg(msg, true, "", nil), nil, nil
	}
}

func SimulateMsgUpdateBeam(k keeper.Keeper, ak account.AccountKeeper, bk bank.Keeper) simtypes.Operation {
	return operationSimulateMsgUpdateBeam(k, ak, bk, "")
}

func sendMsgOpenBeam(r *rand.Rand, app *baseapp.BaseApp, bk bank.Keeper, ak account.AccountKeeper, msg *types.MsgOpenBeam, ctx sdk.Context, chainID string, privkeys []cryptotypes.PrivKey) error {
	var (
		fees sdk.Coins
		err  error
	)

	from, err := sdk.AccAddressFromBech32(msg.GetCreatorAddress())
	if err != nil {
		return err
	}

	fromAccount := ak.GetAccount(ctx, from)
	fromAccountSpendable := bk.SpendableCoins(ctx, fromAccount.GetAddress())

	coins, hasNeg := fromAccountSpendable.SafeSub(*msg.GetAmount())
	if !hasNeg {
		fees, err = simtypes.RandomFees(r, ctx, coins)
		if err != nil {
			return err
		}
	}

	txGenerator := params.MakeTestEncodingConfig().TxConfig
	tx, err := helpers.GenSignedMockTx(r, txGenerator, []sdk.Msg{msg}, fees, helpers.DefaultGenTxGas, chainID, []uint64{fromAccount.GetAccountNumber()}, []uint64{fromAccount.GetSequence()}, privkeys...)
	if err != nil {
		return err
	}

	_, _, err = app.SimDeliver(txGenerator.TxEncoder(), tx)
	if err != nil {
		return err
	}

	return nil
}

func randomSendFields(r *rand.Rand, ctx sdk.Context, accs []simtypes.Account, bk bank.Keeper, ak account.AccountKeeper) (simtypes.Account, simtypes.Account, sdk.Coins, bool) {
	from, _ := simtypes.RandomAcc(r, accs)
	to, _ := simtypes.RandomAcc(r, accs)

	// disallow sending money to yourself
	for from.PubKey.Equals(to.PubKey) {
		to, _ = simtypes.RandomAcc(r, accs)
	}

	acc := ak.GetAccount(ctx, from.Address)
	if acc == nil {
		return from, to, nil, true
	}

	spendable := bk.SpendableCoins(ctx, acc.GetAddress())

	sendCoins := simtypes.RandSubsetCoins(r, spendable)
	if sendCoins.Empty() {
		return from, to, nil, true
	}

	return from, to, sendCoins, false
}
