package notify

import (
	"wood-hot-monitor/internal/config"
	"wood-hot-monitor/pkg/types"
)

// Notifier is the in-app event bus (Wails Event → frontend UI).
type Notifier interface {
	Emit(eventName string, data any)
}

// HotspotAlert is a channel-agnostic payload for a newly discovered hotspot.
type HotspotAlert struct {
	ID         string
	Title      string
	Source     string
	URL        string
	Importance types.Importance
	Summary    string
}

// Alerter fans out business alerts: in-app events, OS notifications, email, etc.
// Checker depends on this interface; it does not call OS APIs directly.
type Alerter interface {
	Notifier
	OnHotspotNew(cfg *config.AppConfig, alert HotspotAlert)
}

const (
	EventCheckerStarted   = "checker:started"
	EventCheckerCompleted = "checker:completed"
	EventCheckerError     = "checker:error"
	EventHotspotNew       = "hotspot:new"
)

// importanceRank maps importance labels to comparable ranks.
// Unknown / empty values rank as 0.
func importanceRank(level types.Importance) int {
	switch level {
	case types.ImportanceLow:
		return 1
	case types.ImportanceMedium:
		return 2
	case types.ImportanceHigh:
		return 3
	case types.ImportanceUrgent:
		return 4
	default:
		return 0
	}
}

// MeetsMinImportance reports whether importance is at least minLevel.
// Empty minLevel defaults to "high".
func MeetsMinImportance(importance, minLevel types.Importance) bool {
	if minLevel == "" {
		minLevel = types.ImportanceHigh
	}
	return importanceRank(importance) >= importanceRank(minLevel)
}
