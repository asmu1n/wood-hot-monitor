package notify

import (
	"log"

	"wood-hot-monitor/internal/config"
)

// Dispatcher routes hotspot alerts according to config:
//   - always emits in-app events for UI
//   - OS notify when osNotifyEnabled and importance meets threshold
//   - email for high/urgent when email is configured
type Dispatcher struct {
	events Notifier
	os     OSNotifier
}

func NewDispatcher(events Notifier, os OSNotifier) *Dispatcher {
	if os == nil {
		os = NoopOSNotifier{}
	}
	return &Dispatcher{events: events, os: os}
}

func (d *Dispatcher) Emit(eventName string, data any) {
	if d.events != nil {
		d.events.Emit(eventName, data)
	}
}

func (d *Dispatcher) OnHotspotNew(cfg *config.AppConfig, alert HotspotAlert) {
	d.Emit(EventHotspotNew, map[string]string{
		"id":     alert.ID,
		"title":  alert.Title,
		"source": alert.Source,
	})

	if cfg != nil && cfg.OSNotifyEnabled && MeetsMinImportance(alert.Importance, cfg.OSNotifyMinImportance) {
		title := "新热点"
		if alert.Source != "" {
			title = "新热点 · " + alert.Source
		}
		body := alert.Title
		if body == "" {
			body = alert.Summary
		}
		id := "hotspot-" + alert.ID
		if err := d.os.Notify(id, title, body, string(alert.Importance), map[string]any{
			"id":     alert.ID,
			"source": alert.Source,
			"url":    alert.URL,
		}); err != nil {
			log.Printf("notify: os notification failed: %v", err)
		}
	}

	if cfg != nil && (alert.Importance == "high" || alert.Importance == "urgent") {
		SendEmailAlert(cfg, EmailAlert{
			Title:      alert.Title,
			Source:     alert.Source,
			URL:        alert.URL,
			Importance: alert.Importance,
			Summary:    alert.Summary,
		})
	}
}
