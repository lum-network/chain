package dfract_test

import (
	"testing"

	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	"github.com/stretchr/testify/suite"

	gogotypes "github.com/cosmos/gogoproto/types"

	"github.com/lum-network/chain/app"
	apptesting "github.com/lum-network/chain/app/testing"
	"github.com/lum-network/chain/x/dfract"
	dfracttypes "github.com/lum-network/chain/x/dfract/types"
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

func (suite *HandlerTestSuite) TestProposal_UpdateParams() {
	var emptyDenpositDenoms []string
	validDepositDenoms := []string{dfracttypes.DefaultDenom, "udfr"}
	invalidDepositAmount := sdk.NewInt(500000)
	newMinDeposit := sdk.NewInt(2000000)

	cases := []struct {
		name            string
		proposal        govtypes.Content
		expectPreError  bool
		expectPostError bool
	}{
		{
			"Partial update nil deposit enablement should be fine",
			dfracttypes.NewUpdateParamsProposal("Test", "Test", "lum1qx2dts3tglxcu0jh47k7ghstsn4nactukljgyj", nil, emptyDenpositDenoms, nil),
			false,
			false,
		},
		{
			"Partial update nil deposit enablement should be fine",
			dfracttypes.NewUpdateParamsProposal("Test", "Test", "lum1qx2dts3tglxcu0jh47k7ghstsn4nactukljgyj", nil, emptyDenpositDenoms, nil),
			false,
			false,
		},
		{
			"Partial update with empty management address should be fine",
			dfracttypes.NewUpdateParamsProposal("Test", "Test", "", &gogotypes.BoolValue{Value: false}, emptyDenpositDenoms, nil),
			false,
			false,
		},
		{
			"Partial update with invalid address should not be fine",
			dfracttypes.NewUpdateParamsProposal("Test", "Test", "lum1qx", nil, emptyDenpositDenoms, nil),
			true,
			true,
		},
		{
			"Partial update with valid deposit denoms should be fine",
			dfracttypes.NewUpdateParamsProposal("Test", "Test", "", nil, validDepositDenoms, nil),
			false,
			false,
		},
		{
			"Partial update with invalid min deposit amount should not be fine",
			dfracttypes.NewUpdateParamsProposal("Test", "Test", "", nil, validDepositDenoms, &invalidDepositAmount),
			true,
			true,
		},
		{
			"Full update should be fine",
			dfracttypes.NewUpdateParamsProposal("Test", "Test", "lum1qx2dts3tglxcu0jh47k7ghstsn4nactukljgyj", &gogotypes.BoolValue{Value: false}, validDepositDenoms, &newMinDeposit),
			false,
			false,
		},
	}

	for _, tc := range cases {
		tc := tc
		suite.Run(tc.name, func() {
			preError := tc.proposal.ValidateBasic()
			if tc.expectPreError {
				suite.Require().Error(preError)
			} else {
				suite.Require().NoError(preError)
			}
			err := suite.handler(suite.ctx, tc.proposal)
			if tc.expectPostError {
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
