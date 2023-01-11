package types

const (
	EventTypeDeposit   = "deposit"
	EventTypeWithdraw  = "withdraw"
	EventTypeMint      = "mint"
	EventTypeBond      = "bond"
	EventTypeUnbonding = "unbonding"
	EventTypeUnbond    = "unbond"

	AttributeKeyDepositor          = "depositor"
	AttributeKeyAmount             = "amount"
	AttributeKeyMinted             = "minted"
	AttributeKeyMicroMintRate      = "micro_mint_rate"
	AttributeKeyMintBlock          = "mint_block"
	AttributeKeyBondDelegator      = "staking_bond_delegator"
	AttributeKeyUnbondingDelegator = "staking_unbonding_delegator"
)
