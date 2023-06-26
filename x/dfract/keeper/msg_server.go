package keeper

import (
	"context"
	"fmt"
	"strconv"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/lum-network/chain/x/dfract/types"
)

type msgServer struct {
	Keeper
}

// NewMsgServerImpl returns an implementation of the MsgServer interface
// for the provided Keeper.
func NewMsgServerImpl(keeper Keeper) types.MsgServer {
	return msgServer{Keeper: keeper}
}

var _ types.MsgServer = msgServer{}

// Deposit creates a deposit from the transaction message
func (k msgServer) Deposit(goCtx context.Context, msg *types.MsgDeposit) (*types.MsgDepositResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	// Acquire the parameters to get the denoms
	params := k.GetParams(ctx)

	// Verify params to see if deposit is enabled
	if !params.IsDepositEnabled {
		return nil, types.ErrDepositNotEnabled
	}

	// Make sure the deposit is made of allowed denom
	for _, denom := range params.DepositDenoms {
		if denom != msg.GetAmount().Denom {
			return nil, types.ErrUnauthorizedDepositDenom
		}
	}

	// Make sure we have an actual deposit to do
	if msg.GetAmount().IsNegative() || msg.GetAmount().IsZero() {
		return nil, types.ErrEmptyDepositAmount
	}

	// Make sure the deposit is sufficient
	if msg.GetAmount().Amount.LT(sdk.NewInt(int64(params.MinDepositAmount))) {
		return nil, types.ErrInsufficientDepositAmount
	}

	// Cast the depositor address
	depositorAddress, err := sdk.AccAddressFromBech32(msg.GetDepositorAddress())
	if err != nil {
		return nil, sdkerrors.ErrInvalidAddress
	}

	// Does the deposit exists ? If not, create it otherwise just append the amount
	deposit, found := k.GetDepositPendingWithdrawal(ctx, depositorAddress)
	if !found {
		deposit = types.Deposit{
			DepositorAddress: msg.GetDepositorAddress(),
			Amount:           msg.GetAmount(),
			CreatedAt:        ctx.BlockTime(),
		}
	} else {
		deposit.Amount = deposit.Amount.Add(msg.GetAmount())
	}

	// Move the funds
	if err := k.BankKeeper.SendCoinsFromAccountToModule(ctx, depositorAddress, types.ModuleName, sdk.NewCoins(msg.GetAmount())); err != nil {
		return nil, err
	}

	// Insert into queue
	k.SetDepositPendingWithdrawal(ctx, depositorAddress, deposit)

	// Trigger deposit event
	ctx.EventManager().Events().AppendEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeDeposit,
			sdk.NewAttribute(types.AttributeKeyDepositor, msg.GetDepositorAddress()),
			sdk.NewAttribute(types.AttributeKeyAmount, msg.GetAmount().String()),
		),
	})
	return &types.MsgDepositResponse{}, nil
}

// WithdrawAndMint process the withdraw and mint based on the MicroMintRate
func (k msgServer) WithdrawAndMint(goCtx context.Context, msg *types.MsgWithdrawAndMint) (*types.MsgWithdrawAndMintResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	// Acquire the parameters to get the denoms
	params := k.GetParams(ctx)

	// make sure the signer is the management address
	if msg.GetAddress() != params.GetManagementAddress() {
		return nil, types.ErrInvalidSignerAddress
	}

	// Iterate over the array of deposit denoms to get the module account balance for a particular denom
	for _, denom := range params.DepositDenoms {
		// Acquire the module account balance
		balance := k.GetModuleAccountBalanceForDenom(ctx, denom)

		// Acquire the list of deposits pending withdrawal and mint
		depositsPendingWithdrawal := k.ListDepositsPendingWithdrawal(ctx)
		depositsPendingMint := k.ListDepositsPendingMint(ctx)

		// Withdrawal the coins from the pending withdrawal deposits using the module account
		if balance.Amount.IsPositive() {
			destinationAddress, err := sdk.AccAddressFromBech32(params.GetManagementAddress())
			if err != nil {
				return nil, sdkerrors.ErrInvalidAddress
			}
			if err := k.BankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, destinationAddress, sdk.NewCoins(balance)); err != nil {
				return nil, err
			}
		}

		// Mint the coins for all the pending mint deposits and add them to the the minted deposits
		for _, deposit := range depositsPendingMint {
			depositorAddress, err := sdk.AccAddressFromBech32(deposit.GetDepositorAddress())
			if err != nil {
				return nil, sdkerrors.ErrInvalidAddress
			}

			toMint := sdk.NewCoin(types.MintDenom, deposit.GetAmount().Amount.MulRaw(msg.GetMicroMintRate()).QuoRaw(MicroPrecision))
			if err := k.BankKeeper.MintCoins(ctx, types.ModuleName, sdk.NewCoins(toMint)); err != nil {
				return nil, err
			}

			if err := k.BankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, depositorAddress, sdk.NewCoins(toMint)); err != nil {
				return nil, err
			}

			k.RemoveDepositPendingMint(ctx, depositorAddress)
			k.AddDepositMinted(ctx, depositorAddress, *deposit)
			// Trigger mint event
			ctx.EventManager().Events().AppendEvents(sdk.Events{
				sdk.NewEvent(
					types.EventTypeMint,
					sdk.NewAttribute(types.AttributeKeyDepositor, depositorAddress.String()),
					sdk.NewAttribute(types.AttributeKeyAmount, deposit.GetAmount().String()),
					sdk.NewAttribute(types.AttributeKeyMinted, toMint.String()),
					sdk.NewAttribute(types.AttributeKeyMicroMintRate, fmt.Sprintf("%d", msg.GetMicroMintRate())),
				),
			})
		}

		// Process the pending withdrawal deposits
		for _, deposit := range depositsPendingWithdrawal {
			depositorAddress, err := sdk.AccAddressFromBech32(deposit.GetDepositorAddress())
			if err != nil {
				return nil, sdkerrors.ErrInvalidAddress
			}
			k.RemoveDepositPendingWithdrawal(ctx, depositorAddress)
			k.SetDepositPendingMint(ctx, depositorAddress, *deposit)

			// Trigger the withdraw events
			ctx.EventManager().Events().AppendEvents(sdk.Events{
				sdk.NewEvent(
					types.EventTypeWithdraw,
					sdk.NewAttribute(types.AttributeKeyDepositor, depositorAddress.String()),
					sdk.NewAttribute(types.AttributeKeyAmount, deposit.GetAmount().String()),
				),
			})
		}

		// Trigger mint event
		ctx.EventManager().Events().AppendEvents(sdk.Events{
			sdk.NewEvent(types.EventTypeMint, sdk.NewAttribute(types.AttributeKeyMintBlock, strconv.FormatInt(ctx.BlockHeight(), 10))),
		})
	}

	return &types.MsgWithdrawAndMintResponse{}, nil
}
