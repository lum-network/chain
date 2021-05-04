package cli

import (
	flag "github.com/spf13/pflag"
)

const (
	FlagReward = "reward"
	FlagReview = "review"
)

func flagSetBeamMetadata() *flag.FlagSet {
	fs := flag.NewFlagSet("", flag.ContinueOnError)

	fs.String(FlagReward, "", "Beam reward metadata")
	fs.String(FlagReview, "", "Beam review metadata")
	return fs
}
