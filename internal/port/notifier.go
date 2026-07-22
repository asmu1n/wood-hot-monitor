package port

import (
	"wood-hot-monitor/internal/config"
	"wood-hot-monitor/pkg/types"
)

const (
	EventCheckerStarted   = "checker:started"
	EventCheckerCompleted = "checker:completed"
	EventCheckerError     = "checker:error"
	EventHotspotNew       = "hotspot:new"
)

type OSNotifier interface {
	Notify(id, title, body, subtitle string, data map[string]any) error
}

type WebViewNotifier interface {
	Emit(eventName string, data any)
}

type Notifier interface {
	WebViewNotifier
	OnHotspotNew(cfg config.NotifyConfig, alert HotspotAlert)
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
