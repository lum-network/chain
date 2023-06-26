package millions

import (
	"fmt"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"
	ibctransfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
	channeltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"
	porttypes "github.com/cosmos/ibc-go/v7/modules/core/05-port/types"
	ibcexported "github.com/cosmos/ibc-go/v7/modules/core/exported"

	"github.com/lum-network/chain/x/icacallbacks"
	icacallbacktypes "github.com/lum-network/chain/x/icacallbacks/types"
	"github.com/lum-network/chain/x/millions/keeper"
	"github.com/lum-network/chain/x/millions/types"
)

// Map the interface
var _ porttypes.IBCModule = IBCModule{}

// IBCModule Declare our IBCModule structure
type IBCModule struct {
	keeper keeper.Keeper
}

// NewIBCModule Initialize a new IBCModule instance
func NewIBCModule(k keeper.Keeper) IBCModule {
	return IBCModule{
		keeper: k,
	}
}

// OnChanOpenInit implements the IBCModule interface
func (im IBCModule) OnChanOpenInit(ctx sdk.Context, order channeltypes.Order, connectionHops []string, portID string, channelID string, channelCap *capabilitytypes.Capability, counterparty channeltypes.Counterparty, version string) (string, error) {
	im.keeper.Logger(ctx).Debug(fmt.Sprintf("OnChanOpenInit: portID %s, channelID %s", portID, channelID))
	return version, nil
}

// OnChanOpenTry implements the IBCModule interface
func (im IBCModule) OnChanOpenTry(ctx sdk.Context, order channeltypes.Order, connectionHops []string, portID, channelID string, chanCap *capabilitytypes.Capability, counterparty channeltypes.Counterparty, counterpartyVersion string) (string, error) {
	return "", nil
}

// OnChanOpenAck implements the IBCModule interface
func (im IBCModule) OnChanOpenAck(ctx sdk.Context, portID, channelID string, counterpartyChannelID string, counterpartyVersion string) error {
	im.keeper.Logger(ctx).Debug(fmt.Sprintf("OnChanOpenAck: portID %s, channelID %s, counterpartyChannelID %s, counterpartyVersion %s", portID, channelID, counterpartyChannelID, counterpartyVersion))

	// Grab the controller connection ID for the port
	controllerConnectionId, err := im.keeper.GetConnectionID(ctx, portID)
	if err != nil {
		ctx.Logger().Error(fmt.Sprintf("Unable to get connection for port: %s", portID))
		return err
	}

	// Extract the interchain account from the connection ID & port
	interchainAccountAddress, found := im.keeper.ICAControllerKeeper.GetInterchainAccountAddress(ctx, controllerConnectionId, portID)
	if !found {
		ctx.Logger().Error("Failed to find the interchain account")
		return nil
	}

	// Get the pool for the matching controller port
	pool, found := im.keeper.GetPoolForControllerPortID(ctx, portID)
	if !found {
		ctx.Logger().Error(fmt.Sprintf("Failed to get pool for controller port id %s", portID))
		return nil
	}

	// Handle which of the ICA addresses has been initialized
	depositAddressPortID := pool.GetIcaDepositPortId()
	prizePoolAddressPortID := pool.GetIcaPrizepoolPortId()

	var accountType string
	if portID == depositAddressPortID {
		accountType = types.ICATypeDeposit
	} else if portID == prizePoolAddressPortID {
		accountType = types.ICATypePrizePool
	}

	// Call our internal callback
	if _, err := im.keeper.OnSetupPoolICACompleted(ctx, pool.PoolId, accountType, interchainAccountAddress); err != nil {
		return err
	}
	im.keeper.Logger(ctx).Info(fmt.Sprintf("OnChanOpenAck: Found pool %d, set interchain account type %s address %s", pool.GetPoolId(), accountType, interchainAccountAddress))
	return nil
}

// OnChanOpenConfirm implements the IBCModule interface
func (im IBCModule) OnChanOpenConfirm(ctx sdk.Context, portID, channelID string) error {
	return nil
}

// OnChanCloseInit implements the IBCModule interface
func (im IBCModule) OnChanCloseInit(ctx sdk.Context, portID, channelID string) error {
	return nil
}

// OnChanCloseConfirm implements the IBCModule interface
func (im IBCModule) OnChanCloseConfirm(ctx sdk.Context, portID, channelID string) error {
	return nil
}

// OnRecvPacket implements the IBCModule interface. A successful acknowledgement
// is returned if the packet data is successfully decoded and the received application logic returns without error.
func (im IBCModule) OnRecvPacket(ctx sdk.Context, packet channeltypes.Packet, relayer sdk.AccAddress) ibcexported.Acknowledgement {
	panic("UNIMPLEMENTED")
}

// OnAcknowledgementPacket implements the IBCModule interface
func (im IBCModule) OnAcknowledgementPacket(ctx sdk.Context, modulePacket channeltypes.Packet, acknowledgement []byte, relayer sdk.AccAddress) error {
	im.keeper.Logger(ctx).Debug(fmt.Sprintf("OnAcknowledgementPacket (millions) - packet: %+v, relayer: %v", modulePacket, relayer))

	// Unpack the response
	ackResponse, err := icacallbacks.UnpackAcknowledgementResponse(ctx, im.keeper.Logger(ctx), acknowledgement, true)
	if err != nil {
		errMsg := fmt.Sprintf("Unable to unpack message data from acknowledgement, Sequence %d, from %s %s, to %s %s: %s", modulePacket.Sequence, modulePacket.SourceChannel, modulePacket.SourcePort, modulePacket.DestinationChannel, modulePacket.DestinationPort, err.Error())
		im.keeper.Logger(ctx).Error(errMsg)
		return errorsmod.Wrapf(icacallbacktypes.ErrInvalidAcknowledgement, errMsg)
	}

	// Print in the logs for easier debugging
	ackInfo := fmt.Sprintf("sequence #%d, from %s %s, to %s %s", modulePacket.Sequence, modulePacket.SourceChannel, modulePacket.SourcePort, modulePacket.DestinationChannel, modulePacket.DestinationPort)
	im.keeper.Logger(ctx).Debug(fmt.Sprintf("Acknowledgement was successfully unmarshalled: ackInfo: %s", ackInfo))

	// Emit the events
	eventType := "ack"
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			eventType,
			sdk.NewAttribute(sdk.AttributeKeyModule, ibctransfertypes.ModuleName),
			sdk.NewAttribute(ibctransfertypes.AttributeKeyAck, ackInfo),
		),
	)

	// Notify the registered callbacks
	err = im.keeper.ICACallbacksKeeper.CallRegisteredICACallback(ctx, modulePacket, ackResponse)
	if err != nil {
		errMsg := fmt.Sprintf("Unable to call registered callback from OnAcknowledgePacket | Sequence %d, from %s %s, to %s %s => %s", modulePacket.Sequence, modulePacket.SourceChannel, modulePacket.SourcePort, modulePacket.DestinationChannel, modulePacket.DestinationPort, err.Error())
		im.keeper.Logger(ctx).Error(errMsg)
		return errorsmod.Wrapf(icacallbacktypes.ErrCallbackFailed, errMsg)
	}
	return nil
}

// OnTimeoutPacket implements the IBCModule interface
func (im IBCModule) OnTimeoutPacket(ctx sdk.Context, modulePacket channeltypes.Packet, relayer sdk.AccAddress) error {
	im.keeper.Logger(ctx).Debug(fmt.Sprintf("OnTimeoutPacket: packet %v, relayer %v", modulePacket, relayer))

	// Construct the timeout response
	ackResponse := icacallbacktypes.AcknowledgementResponse{Status: icacallbacktypes.AckResponseStatus_TIMEOUT}

	// Notify the callbacks
	err := im.keeper.ICACallbacksKeeper.CallRegisteredICACallback(ctx, modulePacket, &ackResponse)
	if err != nil {
		return err
	}
	return nil
}

func (im IBCModule) NegotiateAppVersion(ctx sdk.Context, order channeltypes.Order, connectionID string, portID string, counterparty channeltypes.Counterparty, proposedVersion string) (version string, err error) {
	return proposedVersion, nil
}
