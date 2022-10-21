package client

import (
	govclient "github.com/cosmos/cosmos-sdk/x/gov/client"
	"github.com/lum-network/chain/x/dfract/client/cli"
)

var ProposalHandler = govclient.NewProposalHandler(cli.NewSubmitWithdrawAndMintProposal)
