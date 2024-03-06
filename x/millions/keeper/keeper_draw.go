package keeper

import (
	"crypto/sha256"
	"fmt"
	"math"
	"math/rand"
	"sort"
	"strconv"
	"time"

	errorsmod "cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/lum-network/chain/x/millions/types"
)

type draw struct {
	PrizeIdx  int
	DrawValue sdk.Dec
}

type DepositTWB struct {
	Address string
	Amount  sdkmath.Int
}

type PrizeDraw struct {
	Amount sdkmath.Int
	Winner *DepositTWB
}

type DrawResult struct {
	PrizeDraws     []PrizeDraw
	TotalWinCount  uint64
	TotalWinAmount sdkmath.Int
}

// LaunchNewDraw initiates a new draw and triggers the ICA get reward phase
// See UpdateDrawAtStateICAOp for next phase
func (k Keeper) LaunchNewDraw(ctx sdk.Context, poolID uint64) (*types.Draw, error) {
	// Acquire Pool
	pool, err := k.GetPool(ctx, poolID)
	if err != nil {
		return nil, err
	}
	if pool.LastDrawState != types.DrawState_Unspecified && pool.LastDrawState != types.DrawState_Success {
		// This check is also performed using the last draw entity
		return nil, types.ErrPoolDrawNotDone
	}

	// Initiate new draw procedure
	drawID := pool.GetNextDrawId()
	draw := types.Draw{
		PoolId:          poolID,
		DrawId:          drawID,
		State:           types.DrawState_IcaWithdrawRewards,
		PrizePool:       sdk.NewCoin(pool.Denom, sdk.ZeroInt()),
		CreatedAtHeight: ctx.BlockHeight(),
		UpdatedAtHeight: ctx.BlockHeight(),
		CreatedAt:       ctx.BlockTime(),
		UpdatedAt:       ctx.BlockTime(),
	}
	k.SetPoolDraw(ctx, draw)

	// Update pool with latest draw info
	t := ctx.BlockTime()
	pool.NextDrawId++
	pool.LastDrawCreatedAt = &t
	pool.LastDrawState = draw.State
	k.updatePool(ctx, &pool)

	return k.ClaimYieldOnRemoteZone(ctx, poolID, drawID)
}

// ClaimYieldOnRemoteZone Claim staking rewards from the native chain validators
// - wait for the ICA callback to move to OnClaimYieldOnRemoteZoneCompleted
// - or go to OnClaimYieldOnRemoteZoneCompleted directly upon claim rewards success if local zone
func (k Keeper) ClaimYieldOnRemoteZone(ctx sdk.Context, poolID uint64, drawID uint64) (*types.Draw, error) {
	// Acquire pool config
	pool, err := k.GetPool(ctx, poolID)
	if err != nil {
		return nil, err
	}

	// Acquire Draw
	draw, err := k.GetPoolDraw(ctx, poolID, drawID)
	if err != nil {
		return nil, err
	}

	// Validate pool runner
	poolRunner := k.MustGetPoolRunner(pool.PoolType)
	if err := poolRunner.ClaimYieldOnRemoteZone(ctx, pool, draw); err != nil {
		return &draw, err
	}

	if pool.IsLocalZone(ctx) {
		return k.OnClaimYieldOnRemoteZoneCompleted(ctx, poolID, drawID, false)
	}

	// Special case - no bonded validator
	// Does not need to do any ICA call
	bondedActiveVals, _ := pool.BondedValidators()
	if len(bondedActiveVals) == 0 {
		return k.OnClaimYieldOnRemoteZoneCompleted(ctx, poolID, drawID, false)
	}

	return &draw, nil
}

// OnClaimYieldOnRemoteZoneCompleted Acknowledge the ICA claim rewards from the native chain validators response and trigger an ICQ if success
func (k Keeper) OnClaimYieldOnRemoteZoneCompleted(ctx sdk.Context, poolID uint64, drawID uint64, isError bool) (*types.Draw, error) {
	// Acquire pool config
	pool, err := k.GetPool(ctx, poolID)
	if err != nil {
		return nil, err
	}

	// Acquire Draw
	draw, err := k.GetPoolDraw(ctx, poolID, drawID)
	if err != nil {
		return nil, err
	}
	if draw.State != types.DrawState_IcaWithdrawRewards {
		return &draw, errorsmod.Wrapf(types.ErrIllegalStateOperation, "state should be %s but is %s", types.DrawState_IcaWithdrawRewards.String(), draw.State.String())
	}

	// Abort on errors
	if isError {
		draw.State = types.DrawState_Failure
		draw.ErrorState = types.DrawState_IcaWithdrawRewards
		draw.UpdatedAtHeight = ctx.BlockHeight()
		draw.UpdatedAt = ctx.BlockTime()
		k.SetPoolDraw(ctx, draw)
		pool.LastDrawState = draw.State
		k.updatePool(ctx, &pool)
		return &draw, nil
	}

	draw.State = types.DrawState_IcqBalance
	draw.ErrorState = types.DrawState_Unspecified
	draw.UpdatedAtHeight = ctx.BlockHeight()
	draw.UpdatedAt = ctx.BlockTime()
	k.SetPoolDraw(ctx, draw)

	// Update pool with latest draw info
	pool.LastDrawState = draw.State
	k.updatePool(ctx, &pool)

	// Trigger the balance query
	return k.QueryFreshPrizePoolCoinsOnRemoteZone(ctx, poolID, drawID)
}

// QueryFreshPrizePoolCoinsOnRemoteZone query the available rewards on the remote zone
// - wait for the ICQ callback to move to OnQueryFreshPrizePoolCoinsOnRemoteZoneCompleted
// - or go to OnQueryFreshPrizePoolCoinsOnRemoteZoneCompleted directly upon query rewards success if local zone
func (k Keeper) QueryFreshPrizePoolCoinsOnRemoteZone(ctx sdk.Context, poolID uint64, drawID uint64) (*types.Draw, error) {
	// Acquire pool config
	pool, err := k.GetPool(ctx, poolID)
	if err != nil {
		return nil, err
	}

	// Acquire Draw
	draw, err := k.GetPoolDraw(ctx, poolID, drawID)
	if err != nil {
		return nil, err
	}
	if draw.State != types.DrawState_IcqBalance {
		return &draw, errorsmod.Wrapf(types.ErrIllegalStateOperation, "state should be %s but is %s", types.DrawState_IcqBalance.String(), draw.State.String())
	}

	// If it's a local pool, proceed with local balance fetch and synchronously return
	if pool.IsLocalZone(ctx) {
		moduleAccAddress := sdk.MustAccAddressFromBech32(pool.GetIcaPrizepoolAddress())
		balance := k.BankKeeper.GetBalance(ctx, moduleAccAddress, pool.GetNativeDenom())
		return k.OnQueryFreshPrizePoolCoinsOnRemoteZoneCompleted(ctx, poolID, drawID, sdk.NewCoins(balance), false)
	}

	// Validate pool runner
	poolRunner := k.MustGetPoolRunner(pool.PoolType)
	if err := poolRunner.QueryFreshPrizePoolCoinsOnRemoteZone(ctx, pool, draw); err != nil {
		// Handle the error here and return it
		return k.OnQueryFreshPrizePoolCoinsOnRemoteZoneCompleted(ctx, poolID, drawID, nil, true)
	}

	return &draw, err
}

// OnQueryFreshPrizePoolCoinsOnRemoteZoneCompleted Acknowledge the ICQ query rewards from the remote zone and moves to next step
func (k Keeper) OnQueryFreshPrizePoolCoinsOnRemoteZoneCompleted(ctx sdk.Context, poolID uint64, drawID uint64, coins sdk.Coins, isError bool) (*types.Draw, error) {
	// Acquire pool config
	pool, err := k.GetPool(ctx, poolID)
	if err != nil {
		return nil, err
	}

	// Acquire Draw
	draw, err := k.GetPoolDraw(ctx, poolID, drawID)
	if err != nil {
		return nil, err
	}
	if draw.State != types.DrawState_IcqBalance {
		return &draw, errorsmod.Wrapf(types.ErrIllegalStateOperation, "state should be %s but is %s", types.DrawState_IcqBalance.String(), draw.State.String())
	}

	// Abort on errors
	if isError {
		draw.State = types.DrawState_Failure
		draw.ErrorState = types.DrawState_IcqBalance
		draw.UpdatedAtHeight = ctx.BlockHeight()
		draw.UpdatedAt = ctx.BlockTime()
		k.SetPoolDraw(ctx, draw)
		pool.LastDrawState = draw.State
		k.updatePool(ctx, &pool)
		return &draw, nil
	}

	// Save new draw state
	freshPrizePool := sdk.NewCoin(pool.Denom, sdk.ZeroInt())
	for _, c := range coins {
		if c.Denom == pool.NativeDenom {
			// Do not use Add directly here since we change the coin native denom into the local denom
			// This is due to the fact that the coins params comes from an ICA callback (except for local pools where Denom == NativeDenom)
			freshPrizePool = freshPrizePool.AddAmount(c.Amount)
		}
		// TODO: handle other denoms here ?
		// - in the case of Lum we will receive here the stakers fees as well (inception)
		// - in the case of other zones we might receive other tokens as well
		// - TBD: do something with it at some point
	}

	draw.State = types.DrawState_IbcTransfer
	draw.ErrorState = types.DrawState_Unspecified
	draw.PrizePoolFreshAmount = freshPrizePool.Amount
	draw.PrizePool = draw.PrizePool.Add(freshPrizePool)
	draw.UpdatedAtHeight = ctx.BlockHeight()
	draw.UpdatedAt = ctx.BlockTime()
	k.SetPoolDraw(ctx, draw)

	// Update pool with latest draw info
	pool.LastDrawState = draw.State
	k.updatePool(ctx, &pool)

	return k.TransferFreshPrizePoolCoinsToLocalZone(ctx, poolID, drawID)
}

// TransferFreshPrizePoolCoinsToLocalZone Transfer the claimed rewards to the local chain
// - wait for the ICA callback to move to OnTransferFreshPrizePoolCoinsToLocalZoneCompleted
// - or go to OnTransferFreshPrizePoolCoinsToLocalZoneCompleted directly if local zone
func (k Keeper) TransferFreshPrizePoolCoinsToLocalZone(ctx sdk.Context, poolID uint64, drawID uint64) (*types.Draw, error) {

	// Acquire Pool
	pool, err := k.GetPool(ctx, poolID)
	if err != nil {
		return nil, err
	}
	// Acquire Draw
	draw, err := k.GetPoolDraw(ctx, poolID, drawID)
	if err != nil {
		return nil, err
	}
	if draw.State != types.DrawState_IbcTransfer {
		return &draw, errorsmod.Wrapf(types.ErrIllegalStateOperation, "state should be %s but is %s", types.DrawState_IbcTransfer.String(), draw.State.String())
	}

	// Nothing to transfer
	if draw.PrizePoolFreshAmount.IsZero() {
		return k.OnTransferFreshPrizePoolCoinsToLocalZoneCompleted(ctx, poolID, drawID, false)
	}

	// Validate pool runner
	poolRunner := k.MustGetPoolRunner(pool.PoolType)
	if err := poolRunner.TransferFreshPrizePoolCoinsToLocalZone(ctx, pool, draw); err != nil {
		// Handle the error here and return it
		return k.OnTransferFreshPrizePoolCoinsToLocalZoneCompleted(ctx, poolID, drawID, true)
	}

	if pool.IsLocalZone(ctx) {
		return k.OnTransferFreshPrizePoolCoinsToLocalZoneCompleted(ctx, poolID, drawID, false)
	}

	return &draw, nil
}

// OnTransferFreshPrizePoolCoinsToLocalZoneCompleted Acknowledge the transfer of the claimed rewards
// finalises the Draw if success
func (k Keeper) OnTransferFreshPrizePoolCoinsToLocalZoneCompleted(ctx sdk.Context, poolID uint64, drawID uint64, isError bool) (*types.Draw, error) {
	logger := k.Logger(ctx).With("ctx", "draw_finalise")

	// Acquire Pool
	pool, err := k.GetPool(ctx, poolID)
	if err != nil {
		return nil, err
	}
	// Acquire Draw
	draw, err := k.GetPoolDraw(ctx, poolID, drawID)
	if err != nil {
		return nil, err
	}
	if draw.State != types.DrawState_IbcTransfer {
		return &draw, errorsmod.Wrapf(types.ErrIllegalStateOperation, "state should be %s but is %s", types.DrawState_IbcTransfer.String(), draw.State.String())
	}

	if isError {
		draw.State = types.DrawState_Failure
		draw.ErrorState = types.DrawState_IbcTransfer
		draw.UpdatedAtHeight = ctx.BlockHeight()
		draw.UpdatedAt = ctx.BlockTime()
		k.SetPoolDraw(ctx, draw)
		pool.LastDrawState = draw.State
		k.updatePool(ctx, &pool)
		return &draw, nil
	}

	draw.State = types.DrawState_Drawing
	draw.ErrorState = types.DrawState_Unspecified
	draw.UpdatedAtHeight = ctx.BlockHeight()
	draw.UpdatedAt = ctx.BlockTime()
	k.SetPoolDraw(ctx, draw)
	pool.LastDrawState = draw.State
	k.updatePool(ctx, &pool)

	// Voluntary exit the current cache context here since:
	// - the pool draw rewards have been transfered to the local chain
	// - the ExecuteDraw errors are not part of the transfer context
	// - we do not want to partially commit any state of the execute draw phase by not returning an error
	cacheCtx, writeCache := ctx.CacheContext()
	fDraw, err := k.ExecuteDraw(cacheCtx, poolID, drawID)
	if err != nil {
		// DO NOT commit ExecuteDraw changes in case of failures
		// Re-apply draw state changes (errors) and exit without error
		logger.Error(
			fmt.Sprintf("Failed to execute draw phase: %v", err),
			"pool_id", poolID,
			"draw_id", drawID,
		)
		if fDraw == nil {
			// Case not theoretically possible but necessary to respect implementation safety
			fDraw = &draw
		}
		//nolint:errcheck // error check is not necessary here since we handled it before
		k.OnExecuteDrawCompleted(ctx, &pool, fDraw, err)
		return fDraw, nil
	} else {
		// Commit ExecuteDraw changes in case of success
		logger.Debug(
			"Draw execution completed",
			"pool_id", poolID,
			"draw_id", drawID,
		)
		writeCache()
	}

	return fDraw, nil
}

func (k Keeper) ComputeRandomSeed(ctx sdk.Context) int64 {
	hashBytes := sha256.Sum256(append(ctx.BlockHeader().AppHash, []byte(strconv.Itoa(int(ctx.BlockTime().UnixNano())))...))
	bytesToInt64 := func(bytes []byte) int64 {
		var value int64
		for i := 0; i < 8; i++ {
			value = (value << 8) | int64(bytes[i])
		}
		for i := 8; i < 32; i++ {
			value = (value << 8) | int64(bytes[i]&0x7F)
			if bytes[i]&0x80 != 0 {
				value = -value
			}
		}
		return value
	}
	return bytesToInt64(hashBytes[:])
}

// ExecuteDraw completes the draw phases by effectively drawing prizes
// This is the last phase of a Draw
// WARNING: this method can eventually commit critical partial store updates if the caller does not return on error
func (k Keeper) ExecuteDraw(ctx sdk.Context, poolID uint64, drawID uint64) (*types.Draw, error) {
	// Acquire Pool
	pool, err := k.GetPool(ctx, poolID)
	if err != nil {
		return nil, err
	}
	// Acquire Draw
	draw, err := k.GetPoolDraw(ctx, poolID, drawID)
	if err != nil {
		return nil, err
	}
	if draw.State != types.DrawState_Drawing {
		return &draw, errorsmod.Wrapf(types.ErrIllegalStateOperation, "state should be %s but is %s", types.DrawState_Drawing.String(), draw.State.String())
	}

	// Generate draw random seed
	draw.RandSeed = k.ComputeRandomSeed(ctx)

	// Acquire TWB deposits
	depositorsTWB := k.ComputeDepositsTWB(
		ctx,
		ctx.BlockTime().Add(-pool.DrawSchedule.DrawDelta),
		ctx.BlockTime(),
		k.ListPoolDeposits(ctx, poolID),
	)

	// Acquire available amount from pool to compute final prize pool amount
	draw.PrizePoolRemainsAmount = pool.AvailablePrizePool.Amount
	draw.PrizePool = draw.PrizePool.Add(pool.AvailablePrizePool)

	prizeStrat := pool.PrizeStrategy

	if pool.State == types.PoolState_Closing {
		// Pool is unfortunately closing but some people will be lucky
		// Modify prize distribution to distribute the full prize pool in one draw
		// which is equal to making all batches not unique and with a 100% draw probability
		for i := range prizeStrat.PrizeBatches {
			prizeStrat.PrizeBatches[i].IsUnique = false
			prizeStrat.PrizeBatches[i].DrawProbability = sdk.OneDec()
		}
	}

	// Draw prizes
	dRes, err := k.RunDrawPrizes(
		ctx,
		draw.PrizePool,
		prizeStrat,
		depositorsTWB,
		draw.RandSeed,
	)
	if err != nil {
		return k.OnExecuteDrawCompleted(
			ctx,
			&pool,
			&draw,
			errorsmod.Wrapf(err, "failed to draw prizes for pool %d draw %d", poolID, drawID),
		)
	}

	// Update draw result
	draw.TotalWinCount = dRes.TotalWinCount
	draw.TotalWinAmount = dRes.TotalWinAmount

	// Update pool available prize pool
	pool.AvailablePrizePool = draw.PrizePool.SubAmount(draw.TotalWinAmount)

	// Save draw state before prize distrib in case we don't have prizeRefs
	k.SetPoolDraw(ctx, draw)

	// Distribute prizes and collect fees
	fc := k.NewFeeManager(ctx, pool)
	if err := k.DistributePrizes(ctx, fc, dRes, draw); err != nil {
		return k.OnExecuteDrawCompleted(
			ctx,
			&pool,
			&draw,
			errorsmod.Wrapf(err, "failed to distribute prizes for pool %d draw %d", poolID, drawID),
		)
	}

	// Get the updated draw prizeRefs after DistributePrizes if we have potential winners
	draw, err = k.GetPoolDraw(ctx, poolID, drawID)
	if err != nil {
		return nil, err
	}

	// Send collected fees
	if err := fc.SendCollectedFees(ctx); err != nil {
		return k.OnExecuteDrawCompleted(
			ctx,
			&pool,
			&draw,
			errorsmod.Wrapf(err, "failed to send collected fees for pool %d draw %d", poolID, drawID),
		)
	}

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
		),
		sdk.NewEvent(
			types.EventTypeDrawSuccess,
			sdk.NewAttribute(types.AttributeSeed, strconv.FormatInt(draw.RandSeed, 10)),
			sdk.NewAttribute(types.AttributeKeyPoolID, strconv.FormatUint(draw.PoolId, 10)),
			sdk.NewAttribute(types.AttributeKeyDrawID, strconv.FormatUint(draw.DrawId, 10)),
			sdk.NewAttribute(types.AttributeKeyPrizePool, draw.PrizePool.String()),
			sdk.NewAttribute(types.AttributeKeyTotalWinners, strconv.FormatUint(draw.TotalWinCount, 10)),
			sdk.NewAttribute(types.AttributeKeyTotalWinAmount, draw.TotalWinAmount.String()),
		),
	})

	return k.OnExecuteDrawCompleted(ctx, &pool, &draw, nil)
}

// OnExecuteDrawCompleted wrappers for draw state update upon drawing phase completion
// returns the error specified in parameters and does not produce any internal error
func (k Keeper) OnExecuteDrawCompleted(ctx sdk.Context, pool *types.Pool, draw *types.Draw, err error) (*types.Draw, error) {
	if err != nil {
		draw.State = types.DrawState_Failure
		draw.ErrorState = types.DrawState_Drawing
		draw.UpdatedAtHeight = ctx.BlockHeight()
		draw.UpdatedAt = ctx.BlockTime()
		k.SetPoolDraw(ctx, *draw)
		pool.LastDrawState = draw.State
		k.updatePool(ctx, pool)
		return draw, err
	}

	draw.State = types.DrawState_Success
	draw.ErrorState = types.DrawState_Unspecified
	draw.UpdatedAtHeight = ctx.BlockHeight()
	draw.UpdatedAt = ctx.BlockTime()
	k.SetPoolDraw(ctx, *draw)
	pool.LastDrawState = draw.State
	k.updatePool(ctx, pool)

	if pool.State == types.PoolState_Closing {
		// Continue closing procedure
		// voluntary ignore errors
		if err := k.ClosePool(ctx, pool.GetPoolId()); err != nil {
			k.Logger(ctx).With("ctx", "draw_completed", "pool_id", pool.GetPoolId()).Error("Silently failed to continue close pool procedure: %v", err)
		}
	}
	return draw, err
}

// ComputeDepositsTWB takes deposits and computes the weight based on their deposit time and the draw duration
// It essentially computes the Time Weighted Balance of each deposit for the DrawPrizes phase
func (k Keeper) ComputeDepositsTWB(ctx sdk.Context, depositStartAt time.Time, drawAt time.Time, deposits []types.Deposit) []DepositTWB {
	params := k.GetParams(ctx)

	totalElapsed := drawAt.Unix() - depositStartAt.Unix()
	if totalElapsed < 0 {
		totalElapsed = 0
	}

	var depositsTWB []DepositTWB
	for _, d := range deposits {
		twb := d.Amount.Amount
		if d.State != types.DepositState_Success {
			// Only take into account completed deposits
			continue
		} else if d.IsSponsor {
			// Sponsors waive their drawing chances
			continue
		}
		if d.CreatedAt.After(depositStartAt) {
			// Apply Time Weight for deposits within the draw deposit delta
			elapsed := drawAt.Unix() - d.CreatedAt.Unix()
			if elapsed < int64(params.MinDepositDrawDelta.Seconds()) {
				// Consider deposits which do not abide to the min deposit to draw delta to be 0
				elapsed = 0
			}
			twb = sdkmath.LegacyNewDec(int64(elapsed)).QuoInt64(int64(totalElapsed)).MulInt(d.Amount.Amount).RoundInt()
		}
		dtwb := DepositTWB{
			Address: d.WinnerAddress,
			Amount:  twb,
		}
		depositsTWB = append(depositsTWB, dtwb)
	}

	return depositsTWB
}

// RunDrawPrizes computes available prizes and draws the prizes and their potential winners based on the specified prize strategy
// this method does not store nor send anything, it only computes the DrawResult
func (k Keeper) RunDrawPrizes(ctx sdk.Context, prizePool sdk.Coin, prizeStrat types.PrizeStrategy, deposits []DepositTWB, randSeed int64) (result DrawResult, err error) {
	if prizeStrat.ContainsUniqueBatch() {
		return k.RunDrawPrizesWithUniques(ctx, prizePool, prizeStrat, deposits, randSeed)
	}
	result.TotalWinAmount = sdk.ZeroInt()

	// Compute all prizes probs
	prizes, _, _, err := prizeStrat.ComputePrizesProbs(prizePool)
	if err != nil {
		return result, err
	}

	// Create deposits buffer and mapping
	// - drawBuffer is a representation of the deposits as if we put them on a line with the distance between them being their deposited amount
	// - bufferToDeposit is the link between the deposit position on the line and the actual deposit
	// Visual representation:
	// - drawBuffer = deposits weighted position A(2), B(5), C(3): ..A.....B...C,,,,,
	// - bufferToDeposit = {2: A, 7: B, 12: C}
	// - note that the positions marked by commas (,) instead of dots (.) are outside of the depositors owned area (no winner zone)
	drawBuffer := []sdkmath.Int{}
	bufferToDeposit := map[sdkmath.Int]DepositTWB{}
	totalDeposits := sdk.ZeroInt()
	for _, d := range deposits {
		if d.Amount.LTE(sdk.ZeroInt()) {
			// Ignore 0 deposits just in case since it would break the draw logic
			// can happen due to rounding approximation (TWB)
			continue
		}
		totalDeposits = totalDeposits.Add(d.Amount)
		drawBuffer = append(drawBuffer, totalDeposits)
		bufferToDeposit[totalDeposits] = d
	}

	// Initialise rand source
	rnd := rand.New(rand.NewSource(randSeed))

	// Compute each prize draw and sort by drawValue ascending in order to only iterate forward on the deposits buffer
	// We maintain the draw order here since we do not want to mess with the individual probabilities of prizes
	var draws []draw
	for i := range prizes {
		draws = append(draws, draw{
			DrawValue: sdk.NewDec(rnd.Int63()).QuoInt64(math.MaxInt64),
			PrizeIdx:  i,
		})
	}
	sort.SliceStable(draws, func(i, j int) bool {
		return draws[i].DrawValue.LT(draws[j].DrawValue)
	})

	// Compute each prize winner based on the price draw and the deposit position in the buffer
	// - No winner if the drawValue >= drawProbability
	// Visual representation (no winner case):
	// - drawBuffer  : ..A.....B...C,,,,,,,,,,
	// - drawPosition: .............,,,,,,x,,,
	// - no winner since x is outside the buffer owned by depositors
	//
	// Otherwise the winner is the one which owns the position in the buffer
	// Visual representation (winner case):
	// - drawBuffer  : ..A.....B...C,,,,,,,,,,
	// - drawPosition: .....x.......,,,,,,,,,,
	// - winner is depositor B since they own the range from A) to B]
	i := 0
	result.PrizeDraws = make([]PrizeDraw, len(prizes))
	for _, draw := range draws {
		prize := prizes[draw.PrizeIdx]
		nowinner := false
		winner := false
		if totalDeposits.GT(sdk.ZeroInt()) && draw.DrawValue.LT(prize.DrawProbability) {
			// Prize draw has a winner (inside the buffer owned by depositors)
			// normalize draw position to make it a portion of the depositors owned buffer and ignore the potential extra unassigned buffer part
			drawPosition := draw.DrawValue.Quo(prize.DrawProbability).MulInt(totalDeposits).RoundInt()
			for i < len(drawBuffer) {
				// keep iterating in the buffer
				// winner is the one owning the current portion of the buffer
				if drawPosition.LTE(drawBuffer[i]) {
					dep := bufferToDeposit[drawBuffer[i]]
					result.PrizeDraws[draw.PrizeIdx] = PrizeDraw{
						Amount: prize.Amount,
						Winner: &dep,
					}
					result.TotalWinAmount = result.TotalWinAmount.Add(prize.Amount)
					result.TotalWinCount++
					winner = true
					break
				} else {
					i++
				}
			}
		} else {
			// Prize draw has no winner
			result.PrizeDraws[draw.PrizeIdx] = PrizeDraw{
				Amount: prize.Amount,
				Winner: nil,
			}
			nowinner = true
		}
		if !winner && !nowinner {
			// This should never happen except in case of algorithm failure
			return result, fmt.Errorf("failed to find an outcome for prize draw")
		}
	}

	return result, err
}

// RunDrawPrizesWithUniques computes available prizes and draws the prizes and their potential winners based on the specified prize strategy
// this method does not store nor send anything, it only computes the DrawResult
// this is a variant of RunDrawPrizes which takes into account unique prize batches (i.e: the kind of prize which prevent a winner to win other prizes)
func (k Keeper) RunDrawPrizesWithUniques(ctx sdk.Context, prizePool sdk.Coin, prizeStrat types.PrizeStrategy, deposits []DepositTWB, randSeed int64) (result DrawResult, err error) {
	result.TotalWinAmount = sdk.ZeroInt()

	// Compute all prizes probs
	prizes, _, _, err := prizeStrat.ComputePrizesProbs(prizePool)
	if err != nil {
		return result, err
	}

	// Create deposits buffer and mapping
	// - drawBuffer is a representation of the deposits as if we put them on a line with the distance between them being their deposited amount
	// - bufferToDeposit is the link between the deposit position on the line and the actual deposit
	// Visual representation:
	// - drawBuffer = deposits weighted position A(2), B(5), C(3): ..A.....B...C,,,,,
	// - bufferToDeposit = {2: A, 7: B, 12: C}
	// - note that the positions marked by commas (,) instead of dots (.) are outside of the depositors owned area (no winner zone)
	var drawBuffer []sdkmath.Int
	var bufferToDeposit map[sdkmath.Int]DepositTWB
	var totalDeposits sdkmath.Int
	excludedWinnerAddresses := map[string]bool{}

	refreshBuffer := func() {
		totalDeposits = sdk.ZeroInt()
		drawBuffer = []sdkmath.Int{}
		bufferToDeposit = map[sdkmath.Int]DepositTWB{}
		for _, d := range deposits {
			if d.Amount.LTE(sdk.ZeroInt()) {
				// Ignore 0 deposits just in case since it would break the draw logic
				// can happen due to rounding approximation (TWB)
				continue
			}
			if _, excluded := excludedWinnerAddresses[d.Address]; excluded {
				// Ingore excluded winner addresses
				// can happen if the address already won a prize marked as unique
				continue
			}
			totalDeposits = totalDeposits.Add(d.Amount)
			drawBuffer = append(drawBuffer, totalDeposits)
			bufferToDeposit[totalDeposits] = d
		}
	}
	refreshBuffer()

	// Initialise rand source
	rnd := rand.New(rand.NewSource(randSeed))

	// Compute each prize winner based on the price draw and the deposit position in the buffer
	// - No winner if the drawValue >= drawProbability
	// Visual representation (no winner case):
	// - drawBuffer  : ..A.....B...C,,,,,,,,,,
	// - drawPosition: .............,,,,,,x,,,
	// - no winner since x is outside the buffer owned by depositors
	//
	// Otherwise the winner is the one which owns the position in the buffer
	// Visual representation (winner case):
	// - drawBuffer  : ..A.....B...C,,,,,,,,,,
	// - drawPosition: .....x.......,,,,,,,,,,
	// - winner is depositor B since they own the range from A) to B]
	result.PrizeDraws = make([]PrizeDraw, len(prizes))
	for pIdx, prize := range prizes {
		nowinner := false
		winner := false
		drawValue := sdk.NewDec(rnd.Int63()).QuoInt64(math.MaxInt64)
		if totalDeposits.GT(sdk.ZeroInt()) && drawValue.LT(prize.DrawProbability) {
			// Prize draw has a winner (inside the buffer owned by depositors)
			// normalize draw position to make it a portion of the depositors owned buffer and ignore the potential extra unassigned buffer part
			drawPosition := drawValue.Quo(prize.DrawProbability).MulInt(totalDeposits).RoundInt()
			for i := 0; i < len(drawBuffer); i++ {
				// keep iterating in the buffer
				// winner is the one owning the current portion of the buffer
				if drawPosition.LTE(drawBuffer[i]) {
					dep := bufferToDeposit[drawBuffer[i]]
					result.PrizeDraws[pIdx] = PrizeDraw{
						Amount: prize.Amount,
						Winner: &dep,
					}
					result.TotalWinAmount = result.TotalWinAmount.Add(prize.Amount)
					result.TotalWinCount++
					winner = true
					if prize.IsUnique {
						// Winner gets excluded from winning more prizes
						// Refresh buffer to remove the winner address from the buffer
						excludedWinnerAddresses[dep.Address] = true
						refreshBuffer()
					}
					break
				}
			}
		} else {
			// Prize draw has no winner
			result.PrizeDraws[pIdx] = PrizeDraw{
				Amount: prize.Amount,
				Winner: nil,
			}
			nowinner = true
		}
		if !winner && !nowinner {
			// This should never happen except in case of algorithm failure
			return result, fmt.Errorf("failed to find an outcome for prize draw")
		}
	}

	return result, err
}

// DistributePrizes distributes the prizes if they have a winner
func (k Keeper) DistributePrizes(ctx sdk.Context, fc *feeManager, dRes DrawResult, draw types.Draw) error {
	var prizeRefs []types.PrizeRef
	for _, pd := range dRes.PrizeDraws {
		if pd.Winner != nil {
			winnerAddress, err := sdk.AccAddressFromBech32(pd.Winner.Address)
			if err != nil {
				return types.ErrInvalidWinnerAddress
			}

			nextPrizeId := k.GetNextPrizeIdAndIncrement(ctx)

			coin := sdk.NewCoin(draw.PrizePool.Denom, sdk.NewInt(pd.Amount.Int64()))
			prize := types.Prize{
				PoolId:          draw.PoolId,
				DrawId:          draw.DrawId,
				PrizeId:         nextPrizeId,
				State:           types.PrizeState_Pending,
				WinnerAddress:   winnerAddress.String(),
				Amount:          coin,
				CreatedAtHeight: ctx.BlockHeight(),
				UpdatedAtHeight: ctx.BlockHeight(),
				ExpiresAt:       ctx.BlockTime().Add(k.GetParams(ctx).PrizeExpirationDelta),
				CreatedAt:       ctx.BlockTime(),
				UpdatedAt:       ctx.BlockTime(),
			}

			pz := types.PrizeRef{
				Amount:        pd.Amount,
				WinnerAddress: pd.Winner.Address,
				PrizeId:       prize.PrizeId,
			}
			prizeRefs = append(prizeRefs, pz)

			fc.CollectPrizeFees(ctx, &prize)
			k.AddPrize(ctx, prize)

			ctx.EventManager().EmitEvents(sdk.Events{
				sdk.NewEvent(
					sdk.EventTypeMessage,
					sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
				),
				sdk.NewEvent(
					types.EventTypeNewPrize,
					sdk.NewAttribute(types.AttributeKeyPoolID, strconv.FormatUint(prize.PoolId, 10)),
					sdk.NewAttribute(types.AttributeKeyDrawID, strconv.FormatUint(prize.DrawId, 10)),
					sdk.NewAttribute(types.AttributeKeyPrizeID, strconv.FormatUint(prize.PrizeId, 10)),
					sdk.NewAttribute(types.AttributeKeyWinner, prize.WinnerAddress),
					sdk.NewAttribute(sdk.AttributeKeyAmount, prize.Amount.String()),
				),
			})
		}
	}
	// Only update if we have a winner
	if len(prizeRefs) > 0 {
		draw.PrizesRefs = prizeRefs
		draw.UpdatedAt = ctx.BlockTime()
		draw.UpdatedAtHeight = ctx.BlockHeight()
		// Update PoolDraw state with correct PrizesRefs
		k.SetPoolDraw(ctx, draw)
	}

	return nil
}

// HasPoolDraw Returns a boolean that indicates if the given poolID and drawID combination exists in the KVStore or not
func (k Keeper) HasPoolDraw(ctx sdk.Context, poolID uint64, drawID uint64) bool {
	store := ctx.KVStore(k.storeKey)
	return store.Has(types.GetPoolDrawIDKey(poolID, drawID))
}

// SetPoolDraw Sets a draw result in the KVStore for a given poolID and drawID
func (k Keeper) SetPoolDraw(ctx sdk.Context, draw types.Draw) {
	store := ctx.KVStore(k.storeKey)
	encodedDraw := k.cdc.MustMarshal(&draw)
	store.Set(types.GetPoolDrawIDKey(draw.GetPoolId(), draw.GetDrawId()), encodedDraw)
}

// GetPoolDraw Returns a draw instance for the given poolID and drawID combination
func (k Keeper) GetPoolDraw(ctx sdk.Context, poolID uint64, drawID uint64) (types.Draw, error) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.GetPoolDrawIDKey(poolID, drawID))
	if bz == nil {
		return types.Draw{}, errorsmod.Wrapf(types.ErrPoolDrawNotFound, "%d/%d", poolID, drawID)
	}

	var draw types.Draw
	if err := k.cdc.Unmarshal(bz, &draw); err != nil {
		return types.Draw{}, err
	}

	return draw, nil
}

// PoolDrawsIterator Return a ready to use iterator for a pool draws store
func (k Keeper) PoolDrawsIterator(ctx sdk.Context, poolID uint64) sdk.Iterator {
	kvStore := ctx.KVStore(k.storeKey)
	return sdk.KVStorePrefixIterator(kvStore, types.GetPoolDrawsKey(poolID))
}

// IteratePoolDraws Iterates over a pool draws store, and for each entry call the callback
func (k Keeper) IteratePoolDraws(ctx sdk.Context, poolID uint64, cb func(draw types.Draw) bool) {
	iterator := k.PoolDrawsIterator(ctx, poolID)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var draw types.Draw
		k.cdc.MustUnmarshal(iterator.Value(), &draw)

		if cb(draw) {
			break
		}
	}
}

// ListPoolDraws return the full pool draws list
// expensive operation that should only be used by Genesis like features and unittests
func (k Keeper) ListPoolDraws(ctx sdk.Context, poolID uint64) (draws []types.Draw) {
	k.IteratePoolDraws(ctx, poolID, func(draw types.Draw) bool {
		draws = append(draws, draw)
		return false
	})
	return
}

// DrawsIterator Return a ready to use iterator for the draws store (all draws from all pools)
func (k Keeper) DrawsIterator(ctx sdk.Context) sdk.Iterator {
	kvStore := ctx.KVStore(k.storeKey)
	return sdk.KVStorePrefixIterator(kvStore, types.DrawPrefix)
}

// IterateDraws Iterate over the draws store (all draws from all pools), and for each entry call the callback
func (k Keeper) IterateDraws(ctx sdk.Context, cb func(draw types.Draw) bool) {
	iterator := k.DrawsIterator(ctx)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var draw types.Draw
		k.cdc.MustUnmarshal(iterator.Value(), &draw)

		if cb(draw) {
			break
		}
	}
}

// ListDraws return the full draws list (all draws from all pools)
// expensive operation that should only be used by Genesis like features
func (k Keeper) ListDraws(ctx sdk.Context) (draws []types.Draw) {
	k.IterateDraws(ctx, func(draw types.Draw) bool {
		draws = append(draws, draw)
		return false
	})
	return
}
