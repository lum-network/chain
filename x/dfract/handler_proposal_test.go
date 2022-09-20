package dfract_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/lum-network/chain/app"
	apptesting "github.com/lum-network/chain/app/testing"
	"github.com/lum-network/chain/x/dfract"
	"github.com/lum-network/chain/x/dfract/types"
	"github.com/stretchr/testify/suite"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	"testing"
)

type HandlerTestSuite struct {
	suite.Suite

	app   *app.App
	ctx   sdk.Context
	addrs []sdk.AccAddress

	handler govtypes.Handler
}

func (suite *HandlerTestSuite) SetupTest() {
	app := app.SetupForTesting(false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})

	suite.app = app
	suite.ctx = ctx
	suite.handler = dfract.NewDFractProposalHandler(*app.DFractKeeper)
	suite.addrs = apptesting.AddTestAddrsWithDenom(app, ctx, 2, sdk.NewInt(300000000), "ulum")
}

func generateSpendAndAdjustProposal(spendDestination string, mintRate int64) *types.SpendAndAdjustProposal {
	return types.NewSpendAndAdjustProposal("title", "description", spendDestination, mintRate)
}

func (suite *HandlerTestSuite) TestSpendAndAdjustProposal() {
	cases := []struct {
		name        string
		proposal    *types.SpendAndAdjustProposal
		expectError bool
	}{
		{
			"invalid mint rate",
			generateSpendAndAdjustProposal(suite.addrs[0].String(), -1),
			true,
		},
		{
			"invalid address",
			generateSpendAndAdjustProposal("test", 0),
			true,
		},
		/*{
			"valid proposal",
			generateSpendAndAdjustProposal(suite.addrs[1].String(), 0),
			false,
		},*/
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
