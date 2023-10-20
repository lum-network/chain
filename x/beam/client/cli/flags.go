package cli

import (
	flag "github.com/spf13/pflag"
)

const (
	FlagOwner               = "owner"
	FlagData                = "data"
	FlagStatus              = "status"
	FlagAmount              = "amount"
	FlagHideContent         = "hide-content"
	FlagCancelReason        = "cancel-reason"
	FlagClosesAtBlock       = "closes-at-block"
	FlagClaimExpiresAtBlock = "claim-expires-at-block"
)

func flagSetOwner() *flag.FlagSet {
	fs := flag.NewFlagSet("", flag.ContinueOnError)
	fs.String(FlagOwner, "", "Beam destination owner")
	return fs
}

func flagSetBeamMetadata() *flag.FlagSet {
	fs := flag.NewFlagSet("", flag.ContinueOnError)

	fs.String(FlagAmount, "", "Beam amount")
	fs.String(FlagData, "", "Beam metadata")
	fs.Int32(FlagStatus, 1, "Beam new status")
	fs.Bool(FlagHideContent, false, "Beam hide content")
	fs.String(FlagCancelReason, "", "Beam cancel reason")
	fs.Int32(FlagClosesAtBlock, 0, "Beam closes at block")
	fs.Int32(FlagClaimExpiresAtBlock, 0, "Beam claim expires at block")
	return fs
}
