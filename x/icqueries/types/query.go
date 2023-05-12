package types

import (
	errorsmod "cosmossdk.io/errors"
	"strings"
)

func (query *Query) ValidateBasic() error {
	if len(query.GetId()) <= 0 {
		return errorsmod.Wrapf(ErrInvalidQuery, "query id is required")
	}

	if len(query.GetConnectionId()) <= 0 {
		return errorsmod.Wrapf(ErrInvalidQuery, "connection id is required")
	}

	if !strings.HasPrefix(query.GetConnectionId(), "connection") {
		return errorsmod.Wrapf(ErrInvalidQuery, "connection id must start with connection")
	}

	if len(query.GetChainId()) <= 0 {
		return errorsmod.Wrapf(ErrInvalidQuery, "chain id is required")
	}

	if len(query.GetQueryType()) <= 0 {
		return errorsmod.Wrapf(ErrInvalidQuery, "query type is required")
	}

	if len(query.GetRequest()) <= 0 {
		return errorsmod.Wrapf(ErrInvalidQuery, "query request is required")
	}

	if len(query.GetCallbackId()) <= 0 {
		return errorsmod.Wrapf(ErrInvalidQuery, "callback id is required")
	}
	return nil
}
