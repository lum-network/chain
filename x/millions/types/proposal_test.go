package types_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	millionstypes "github.com/lum-network/chain/x/millions/types"
)

type ProposalTestSuite struct {
	suite.Suite
}

func TestProposalTestSuite(t *testing.T) {
	suite.Run(t, new(ProposalTestSuite))
}

func (suite *ProposalTestSuite) TestKeysTypes() {
	suite.Require().Equal("millions", (&millionstypes.ProposalRegisterPool{}).ProposalRoute())
	suite.Require().Equal("RegisterPool", (&millionstypes.ProposalRegisterPool{}).ProposalType())
	suite.Require().Equal("millions", (&millionstypes.ProposalUpdatePool{}).ProposalRoute())
	suite.Require().Equal("UpdatePool", (&millionstypes.ProposalUpdatePool{}).ProposalType())
	suite.Require().Equal("millions", (&millionstypes.ProposalUpdatePool{}).ProposalRoute())
	suite.Require().Equal("UpdateParams", (&millionstypes.ProposalUpdateParams{}).ProposalType())
}
