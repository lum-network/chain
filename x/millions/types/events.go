package types

const (
	EventTypeRegisterPool         = "pool_new"
	EventTypeUpdatePool           = "pool_update"
	EventTypeNewDraw              = "draw_new"
	EventTypeDrawSuccess          = "draw_success"
	EventTypeDeposit              = "deposit"
	EventTypeDepositRetry         = "deposit_retry"
	EventTypeNewPrize             = "prize_new"
	EventTypeClaimPrize           = "prize_claim"
	EventTypeClawbackPrize        = "prize_clawback"
	EventTypeWithdrawDeposit      = "withdraw_deposit"
	EventTypeWithdrawDepositRetry = "withdraw_deposit_retry"

	AttributeKeyPoolID         = "pool_id"
	AttributeKeyDrawID         = "draw_id"
	AttributeKeyPrizeID        = "prize_id"
	AttributeKeyDepositID      = "deposit_id"
	AttributeKeyWithdrawalID   = "withdrawal_id"
	AttributeKeyState          = "state"
	AttributeKeyDepositor      = "depositor"
	AttributeKeyWinner         = "winner"
	AttributeKeyRecipient      = "recipient"
	AttributeKeyPrizePool      = "prize_pool"
	AttributeKeyTotalWinners   = "total_winners"
	AttributeKeyTotalWinAmount = "total_win_amount"
)
