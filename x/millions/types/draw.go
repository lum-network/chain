package types

import (
	errorsmod "cosmossdk.io/errors"
)

// ValidateDrawRetryBasic validates the incoming message for a draw retry.
func (msg *MsgDrawRetry) ValidateDrawRetryBasic() error {
	if msg.PoolId == UnknownID {
		return errorsmod.Wrapf(ErrInvalidID, "pool ID")
	}
	if msg.DrawId == UnknownID {
		return errorsmod.Wrapf(ErrInvalidID, "draw ID")
	}

	return nil
}
