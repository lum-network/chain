package cli

import (
	flag "github.com/spf13/pflag"
)

const (
	FlagReward = "reward"
	FlagReview = "review"
	FlagStatus = "status"
)

func flagSetBeamMetadata() *flag.FlagSet {
	fs := flag.NewFlagSet("", flag.ContinueOnError)

	fs.String(FlagReward, "", "Beam reward metadata")
	fs.String(FlagReview, "", "Beam review metadata")
	fs.Int32(FlagStatus, 0, "Beam new status")
	return fs
}
