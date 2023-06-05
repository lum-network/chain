package millions

import (
	"fmt"

	porttypes "github.com/cosmos/ibc-go/v7/modules/core/05-port/types"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"
	ibctransfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
	channeltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"
	ibcexported "github.com/cosmos/ibc-go/v7/modules/core/exported"

	"github.com/lum-network/chain/x/icacallbacks"
	icacallbacktypes "github.com/lum-network/chain/x/icacallbacks/types"
	"github.com/lum-network/chain/x/millions/keeper"
)

// IBCMiddleware Declare our IBCMiddleware structure
type IBCMiddleware struct {
	keeper keeper.Keeper
	app    porttypes.IBCModule
}

// NewIBCMiddleware Initialize a new IBCModule instance
func NewIBCMiddleware(k keeper.Keeper, app porttypes.IBCModule) IBCMiddleware {
	return IBCMiddleware{
		keeper: k,
		app:    app,
	}
}

// OnChanOpenInit implements the IBCModule interface
func (im IBCMiddleware) OnChanOpenInit(ctx sdk.Context, order channeltypes.Order, connectionHops []string, portID string, channelID string, channelCap *capabilitytypes.Capability, counterparty channeltypes.Counterparty, version string) (string, error) {
	return im.app.OnChanOpenInit(
		ctx,
		order,
		connectionHops,
		portID,
		channelID,
		channelCap,
		counterparty,
		version,
	)
}

// OnChanOpenTry implements the IBCModule interface
func (im IBCMiddleware) OnChanOpenTry(ctx sdk.Context, order channeltypes.Order, connectionHops []string, portID, channelID string, chanCap *capabilitytypes.Capability, counterparty channeltypes.Counterparty, counterpartyVersion string) (string, error) {
	return im.app.OnChanOpenTry(
		ctx,
		order,
		connectionHops,
		portID,
		channelID,
		chanCap,
		counterparty,
		counterpartyVersion,
	)
}

// OnChanOpenAck implements the IBCModule interface
func (im IBCMiddleware) OnChanOpenAck(ctx sdk.Context, portID, channelID string, counterpartyChannelID string, counterpartyVersion string) error {
	return im.app.OnChanOpenAck(ctx, portID, channelID, counterpartyChannelID, counterpartyVersion)
}

// OnChanOpenConfirm implements the IBCModule interface
func (im IBCMiddleware) OnChanOpenConfirm(ctx sdk.Context, portID, channelID string) error {
	return im.app.OnChanOpenConfirm(ctx, portID, channelID)
}

// OnChanCloseInit implements the IBCModule interface
func (im IBCMiddleware) OnChanCloseInit(ctx sdk.Context, portID, channelID string) error {
	return im.app.OnChanCloseInit(ctx, portID, channelID)
}

// OnChanCloseConfirm implements the IBCModule interface
func (im IBCMiddleware) OnChanCloseConfirm(ctx sdk.Context, portID, channelID string) error {
	return im.app.OnChanCloseConfirm(ctx, portID, channelID)
}

// OnRecvPacket implements the IBCModule interface. A successful acknowledgement
// is returned if the packet data is successfully decoded and the receive application
// logic returns without error.
func (im IBCMiddleware) OnRecvPacket(ctx sdk.Context, packet channeltypes.Packet, relayer sdk.AccAddress) ibcexported.Acknowledgement {
	return im.app.OnRecvPacket(ctx, packet, relayer)
}

// OnAcknowledgementPacket implements the IBCModule interface
func (im IBCMiddleware) OnAcknowledgementPacket(ctx sdk.Context, modulePacket channeltypes.Packet, acknowledgement []byte, relayer sdk.AccAddress) error {
	im.keeper.Logger(ctx).Debug(fmt.Sprintf("OnAcknowledgementPacket (millions) - packet: %+v, relayer: %v", modulePacket, relayer))

	// Unpack the response
	ackResponse, err := icacallbacks.UnpackAcknowledgementResponse(ctx, im.keeper.Logger(ctx), acknowledgement, false)
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

	return im.app.OnAcknowledgementPacket(ctx, modulePacket, acknowledgement, relayer)
}

// OnTimeoutPacket implements the IBCModule interface
func (im IBCMiddleware) OnTimeoutPacket(ctx sdk.Context, modulePacket channeltypes.Packet, relayer sdk.AccAddress) error {
	im.keeper.Logger(ctx).Debug(fmt.Sprintf("OnTimeoutPacket: packet %v, relayer %v", modulePacket, relayer))

	// Construct the timeout response
	ackResponse := icacallbacktypes.AcknowledgementResponse{Status: icacallbacktypes.AckResponseStatus_TIMEOUT}

	// Notify the callbacks
	err := im.keeper.ICACallbacksKeeper.CallRegisteredICACallback(ctx, modulePacket, &ackResponse)
	if err != nil {
		return err
	}
	return im.app.OnTimeoutPacket(ctx, modulePacket, relayer)
}

func (im IBCMiddleware) NegotiateAppVersion(ctx sdk.Context, order channeltypes.Order, connectionID string, portID string, counterparty channeltypes.Counterparty, proposedVersion string) (version string, err error) {
	return proposedVersion, nil
}
