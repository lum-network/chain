package types_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	millionstypes "github.com/lum-network/chain/x/dfract/types"
)

type ProposalTestSuite struct {
	suite.Suite
}

func TestProposalTestSuite(t *testing.T) {
	suite.Run(t, new(ProposalTestSuite))
}

func (suite *ProposalTestSuite) TestKeysTypes() {
	suite.Require().Equal("dfract", (&millionstypes.ProposalUpdateParams{}).ProposalRoute())
	suite.Require().Equal("UpdateParamsProposal", (&millionstypes.ProposalUpdateParams{}).ProposalType())
}
