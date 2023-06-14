package types

import (
	"fmt"
	time "time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (ps DrawSchedule) ValidateBasic(params Params) error {
	if ps.DrawDelta < params.MinDrawScheduleDelta {
		return fmt.Errorf("draw delta cannot be lower than %s", params.MinDrawScheduleDelta.String())
	}
	if ps.DrawDelta > params.MaxDrawScheduleDelta {
		return fmt.Errorf("draw delta cannot be higher than %s", params.MaxDrawScheduleDelta.String())
	}
	return nil
}

// ValidateNew drawSchedule validation at pool creation time.
func (ps DrawSchedule) ValidateNew(ctx sdk.Context, params Params) error {
	if ps.DrawDelta < params.MinDrawScheduleDelta {
		return fmt.Errorf("draw delta cannot be lower than %s", params.MinDrawScheduleDelta.String())
	}
	if ps.DrawDelta > params.MaxDrawScheduleDelta {
		return fmt.Errorf("draw delta cannot be higher than %s", params.MaxDrawScheduleDelta.String())
	}
	if ps.InitialDrawAt.Before(ctx.BlockTime().Add(ps.DrawDelta)) {
		return fmt.Errorf("initial draw must be in the future and at least after draw delta")
	}
	return nil
}

// sanitizeTime returns the time rounded to the minute.
func (ps DrawSchedule) sanitizeTime(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), 0, 0, t.Location())
}

// Sanitized returns the draw schedule with time rounded to the minute.
func (ps DrawSchedule) Sanitized() DrawSchedule {
	return DrawSchedule{
		InitialDrawAt: ps.sanitizeTime(ps.GetInitialDrawAt()),
		DrawDelta:     ps.GetDrawDelta(),
	}
}

// ShouldDraw returns whether or not the current block time is past the next draw time (= time to launch draw).
func (ps DrawSchedule) ShouldDraw(ctx sdk.Context, lastDrawAt *time.Time) bool {
	if lastDrawAt == nil {
		return !ctx.BlockTime().Before(ps.InitialDrawAt)
	}
	// Simulate lda as if it did run at the appropriate minute to prevent small time drifts
	saneLda := time.Date(
		lastDrawAt.Year(),
		lastDrawAt.Month(),
		lastDrawAt.Day(),
		lastDrawAt.Hour(),
		ps.InitialDrawAt.Minute(),
		0,
		0,
		lastDrawAt.Location(),
	)
	return !ctx.BlockTime().Before(saneLda.Add(ps.DrawDelta))
}
