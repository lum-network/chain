package types

// DefaultParams return the default dfract module params
func DefaultParams() Params {
	return Params{
		DepositDenom:     "ibc/05554A9BFDD28894D7F18F4C707AA0930D778751A437A9FE1F4684A3E1199728", // USDC ibc denom from Osmosis to Lum Network mainnet
		MintDenom:        "udfr",
		MinDepositAmount: 1000000,
	}
}
