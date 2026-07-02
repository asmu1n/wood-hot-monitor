package event

import "github.com/wailsapp/wails/v3/pkg/application"

type Notifier interface {
	Emit(eventName string, data any)
}

const (
	EventCheckerStarted   = "checker:started"
	EventCheckerCompleted = "checker:completed"
	EventCheckerError     = "checker:error"
	EventHotspotNew       = "hotspot:new"
)

type WailsNotifier struct {
	app *application.App
}

// 实现 events.Notifier 接口
func (w *WailsNotifier) Emit(eventName string, data any) {
	if w.app != nil {
		w.app.Event.EmitEvent(&application.CustomEvent{
			Name: eventName,
			Data: data,
		})
	}
}

func NewWailsNotifier(app *application.App) *WailsNotifier {
	return &WailsNotifier{
		app: app,
	}
}
