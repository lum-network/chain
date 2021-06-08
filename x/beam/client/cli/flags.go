package cli

import (
	flag "github.com/spf13/pflag"
)

const (
	FlagOwner  = "owner"
	FlagData   = "data"
	FlagStatus = "status"
	FlagHideContent = "hide-content"
	FlagCancelReason = "cancel-reason"
)

func flagSetOwner() *flag.FlagSet {
	fs := flag.NewFlagSet("", flag.ContinueOnError)
	fs.String(FlagOwner, "", "Beam destination owner")
	return fs
}

func flagSetBeamMetadata() *flag.FlagSet {
	fs := flag.NewFlagSet("", flag.ContinueOnError)

	fs.String(FlagData, "", "Beam metadata")
	fs.Int32(FlagStatus, 0, "Beam new status")
	fs.Bool(FlagHideContent, false, "Beam hide content")
	fs.String(FlagCancelReason, "", "Beam cancel reason")
	return fs
}
