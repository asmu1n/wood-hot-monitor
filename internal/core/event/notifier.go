package event

type Notifier interface {
	Emit(eventName string, data any)
}

const (
	EventCheckerStarted   = "checker:started"
	EventCheckerCompleted = "checker:completed"
	EventCheckerError     = "checker:error"
	EventHotspotNew       = "hotspot:new"
)
