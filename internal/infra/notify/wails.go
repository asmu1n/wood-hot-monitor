package notify

import (
	"github.com/wailsapp/wails/v3/pkg/application"
)

type wailsWebViewNotifier struct {
	app *application.App
}

func (w *wailsWebViewNotifier) Emit(eventName string, data any) {
	if w.app != nil {
		w.app.Event.EmitEvent(&application.CustomEvent{
			Name: eventName,
			Data: data,
		})
	}
}

func NewWailsNotifier(app *application.App) *wailsWebViewNotifier {
	return &wailsWebViewNotifier{app: app}
}
