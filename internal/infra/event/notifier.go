package event

import (
	corevent "wood-hot-monitor/internal/core/event"

	"github.com/wailsapp/wails/v3/pkg/application"
)

type WailsNotifier struct {
	app *application.App
}

func (w *WailsNotifier) Emit(eventName string, data any) {
	if w.app != nil {
		w.app.Event.EmitEvent(&application.CustomEvent{
			Name: eventName,
			Data: data,
		})
	}
}

func NewWailsNotifier(app *application.App) corevent.Notifier {
	return &WailsNotifier{app: app}
}
