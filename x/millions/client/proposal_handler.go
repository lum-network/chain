package client

import (
	govclient "github.com/cosmos/cosmos-sdk/x/gov/client"

	"github.com/lum-network/chain/x/millions/client/cli"
)

var (
	RegisterPoolProposalHandler = govclient.NewProposalHandler(cli.CmdProposalRegisterPool)
	UpdatePoolProposalHandler   = govclient.NewProposalHandler(cli.CmdProposalUpdatePool)
	UpdateParamsProposalHandler = govclient.NewProposalHandler(cli.CmdProposalUpdateParams)
)
