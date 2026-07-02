package email

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

const resendAPIURL = "https://api.resend.com/emails"

type Service struct {
	apiKey    string
	fromEmail string
}

func NewService(apiKey, fromEmail string) *Service {
	return &Service{apiKey: apiKey, fromEmail: fromEmail}
}

type sendRequest struct {
	From    string   `json:"from"`
	To      []string `json:"to"`
	Subject string   `json:"subject"`
	HTML    string   `json:"html"`
}

// SendHotspotAlert 发送热点告警邮件（仅 high/urgent 级别触发）
func (s *Service) SendHotspotAlert(toEmail, title, source, importance, summary, url string) error {
	if s.apiKey == "" || toEmail == "" {
		return nil
	}

	subject := fmt.Sprintf("[%s] 热点告警: %s", importance, title)
	html := fmt.Sprintf(`
		<h2>🔥 热点告警</h2>
		<p><strong>标题:</strong> %s</p>
		<p><strong>来源:</strong> %s</p>
		<p><strong>级别:</strong> %s</p>
		<p><strong>摘要:</strong> %s</p>
		<p><a href="%s">查看原文</a></p>
	`, title, source, importance, summary, url)

	body := sendRequest{
		From:    s.fromEmail,
		To:      []string{toEmail},
		Subject: subject,
		HTML:    html,
	}

	payload, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("marshal email body: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, resendAPIURL, bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("create email request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+s.apiKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("send email: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("resend api error: status %d", resp.StatusCode)
	}
	return nil
}
