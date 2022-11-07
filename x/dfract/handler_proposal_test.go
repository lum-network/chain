package dfract_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypesv1beta1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	"github.com/lum-network/chain/app"
	apptesting "github.com/lum-network/chain/app/testing"
	"github.com/lum-network/chain/x/dfract"
	"github.com/lum-network/chain/x/dfract/types"
	"github.com/stretchr/testify/suite"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
)

type HandlerTestSuite struct {
	suite.Suite

	app   *app.App
	ctx   sdk.Context
	addrs []sdk.AccAddress

	handler govtypesv1beta1.Handler
}

func (suite *HandlerTestSuite) SetupTest() {
	app := app.SetupForTesting(false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})

	suite.app = app
	suite.ctx = ctx
	suite.handler = dfract.NewDFractProposalHandler(*app.DFractKeeper)
	suite.addrs = apptesting.AddTestAddrsWithDenom(app, ctx, 2, sdk.NewInt(300000000), "ulum")
}

func generateWithdrawAndMintProposal(withdrawalAddress string, microMintRate int64) *types.WithdrawAndMintProposal {
	return types.NewWithdrawAndMintProposal("title", "description", withdrawalAddress, microMintRate)
}

func (suite *HandlerTestSuite) TestWithdrawAndMintProposal() {
	cases := []struct {
		name        string
		proposal    *types.WithdrawAndMintProposal
		expectError bool
	}{
		{
			"micro mint rate cannot be negative",
			generateWithdrawAndMintProposal(suite.addrs[0].String(), -1),
			true,
		},
		{
			"invalid address",
			generateWithdrawAndMintProposal("test", 0),
			true,
		},
	}

	for _, tc := range cases {
		tc := tc
		suite.Run(tc.name, func() {
			err := suite.handler(suite.ctx, tc.proposal)
			if tc.expectError {
				suite.Require().Error(err)
			} else {
				suite.Require().NoError(err)
			}
		})
	}
}

func TestHandlerSuite(t *testing.T) {
	suite.Run(t, new(HandlerTestSuite))
}
