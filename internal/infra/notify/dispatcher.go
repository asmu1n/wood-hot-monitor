package notify

import (
	"log"

	"wood-hot-monitor/internal/config"
	"wood-hot-monitor/internal/port"
	"wood-hot-monitor/pkg/types"
)

// dispatcher routes hotspot alerts according to config:
//   - always emits in-app events for UI
//   - OS notify when osNotifyEnabled and importance meets threshold
//   - email for high/urgent when email is configured
type dispatcher struct {
	wv port.WebViewNotifier
	os port.OSNotifier
}

func NewDispatcher(wv port.WebViewNotifier, os port.OSNotifier) port.Alerter {
	if os == nil {
		os = NoopOSNotifier{}
	}
	return &dispatcher{wv: wv, os: os}
}

func (d *dispatcher) Emit(eventName string, data any) {
	if d.wv != nil {
		d.wv.Emit(eventName, data)
	}
}

func (d *dispatcher) OnHotspotNew(cfg config.NotifyConfig, alert port.HotspotAlert) {
	d.Emit(port.EventHotspotNew, map[string]string{
		"id":     alert.ID,
		"title":  alert.Title,
		"source": alert.Source,
	})

	if cfg.OSNotifyEnabled && meetsMinImportance(alert.Importance, cfg.OSNotifyMinImportance) {
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

	if alert.Importance == types.ImportanceHigh || alert.Importance == types.ImportanceUrgent {
		SendEmailAlert(cfg, EmailAlert{
			Title:      alert.Title,
			Source:     alert.Source,
			URL:        alert.URL,
			Importance: alert.Importance,
			Summary:    alert.Summary,
		})
	}
}

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

// meetsMinImportance reports whether importance is at least minLevel.
// Empty minLevel defaults to "high".
func meetsMinImportance(importance, minLevel types.Importance) bool {
	if minLevel == "" {
		minLevel = types.ImportanceHigh
	}
	return importanceRank(importance) >= importanceRank(minLevel)
}
