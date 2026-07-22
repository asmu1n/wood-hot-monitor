package notify

import (
	"fmt"
	"log"
	"sync"
	"wood-hot-monitor/internal/port"

	"github.com/wailsapp/wails/v3/pkg/services/notifications"
)

type wailsOSNotifier struct {
	ns *notifications.NotificationService

	authOnce sync.Once
	authOK   bool
	authErr  error
}

func NewOSNotifier(ns *notifications.NotificationService) port.OSNotifier {
	return &wailsOSNotifier{ns: ns}
}

// Notify sends a basic system notification. Requests authorization on first use (macOS).
func (n *wailsOSNotifier) Notify(id, title, body, subtitle string, data map[string]any) error {
	if n == nil || n.ns == nil {
		return fmt.Errorf("os notifier not initialized")
	}
	if id == "" {
		return fmt.Errorf("notification id is required")
	}
	if title == "" {
		return fmt.Errorf("notification title is required")
	}

	if err := n.ensureAuthorized(); err != nil {
		return err
	}

	opts := notifications.NotificationOptions{
		ID:       id,
		Title:    title,
		Body:     body,
		Subtitle: subtitle,
		Data:     data,
		ThreadID: "hotspots",
	}
	return n.ns.SendNotification(opts)
}

func (n *wailsOSNotifier) ensureAuthorized() error {
	n.authOnce.Do(func() {
		ok, err := n.ns.CheckNotificationAuthorization()
		if err != nil {
			n.authErr = err
			return
		}
		if ok {
			n.authOK = true
			return
		}
		ok, err = n.ns.RequestNotificationAuthorization()
		if err != nil {
			n.authErr = err
			return
		}
		n.authOK = ok
		if !ok {
			n.authErr = fmt.Errorf("os notification authorization denied")
		}
	})
	if !n.authOK {
		if n.authErr != nil {
			return n.authErr
		}
		return fmt.Errorf("os notification not authorized")
	}
	return nil
}

// NoopOSNotifier is used when the OS notification service is unavailable.
type NoopOSNotifier struct{}

func (NoopOSNotifier) Notify(id, title, body, subtitle string, data map[string]any) error {
	log.Printf("notify: os notifications disabled (noop): %s — %s", title, body)
	return nil
}
