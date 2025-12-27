package notify

import (
	"cert-manager-backend/internal/config"
	"cert-manager-backend/internal/model"
	"cert-manager-backend/internal/service"
	"context"
	"fmt"
	"log"
	"net/smtp"
	"strings"
	"time"
)

type EmailNotifier struct {
	cfg         config.SMTPConfig
	frontendURL string // NEW field
	historySvc  service.HistoryService
}

// Updated Constructor to accept frontendURL
func NewEmailNotifier(cfg config.SMTPConfig, frontendURL string, historySvc service.HistoryService) *EmailNotifier {
	return &EmailNotifier{
		cfg:         cfg,
		frontendURL: frontendURL,
		historySvc:  historySvc,
	}
}

// --- PART 1: Notifier Interface (Alerts) ---

func (e *EmailNotifier) Notify(ctx context.Context, certs []model.CertResponse, users map[string]model.User) error {
	if len(certs) == 0 {
		return nil
	}

	// 1. Filter History (Deduplication)
	toSend, err := e.historySvc.FilterByCertID(ctx, certs, string(model.AlertTypeEmail), model.FrequencyDaily)
	if err != nil {
		log.Printf("⚠️ [EmailNotifier] Failed to check history, defaulting to sending all: %v", err)
		toSend = certs
	}

	if len(toSend) == 0 {
		return nil
	}

	// 2. Group by Owner
	buckets := make(map[string][]model.CertResponse)
	for _, cert := range toSend {
		buckets[cert.OwnerID] = append(buckets[cert.OwnerID], cert)
	}

	// 3. Send
	var sentCerts []model.CertResponse
	for ownerID, userCerts := range buckets {
		user, exists := users[ownerID]
		if !exists || user.Email == "" || !user.EmailEnabled {
			continue
		}

		// Build content
		subject := fmt.Sprintf("Action Required: %d Certificates Expiring Soon", len(userCerts))
		body := e.buildAlertHTML(user, userCerts)

		// Send via Brevo
		if err := e.sendSMTP(user.Email, subject, body); err != nil {
			log.Printf("❌ [EmailNotifier] Failed to send alert to %s: %v", user.Email, err)
		} else {
			log.Printf("✅ [EmailNotifier] Sent alert to %s", user.Email)
			sentCerts = append(sentCerts, userCerts...)
		}
	}

	// 4. Record History
	if len(sentCerts) > 0 {
		e.historySvc.RecordSent(ctx, sentCerts, string(model.AlertTypeEmail))
	}

	return nil
}

// --- PART 2: EmailService Interface (Auth) ---

func (e *EmailNotifier) SendVerificationEmail(toEmail, token string) error {
	// Use dynamic URL
	link := fmt.Sprintf("%s/verify-email?token=%s", e.frontendURL, token)
	subject := "Verify your CertMonitor Account"

	// Simple HTML Template
	body := fmt.Sprintf(`
		<html>
		<body style="font-family: Arial, sans-serif; padding: 20px;">
			<h2>Welcome to CertMonitor!</h2>
			<p>Please verify your email address to activate your account.</p>
			<p>
				<a href="%s" style="background-color: #2563EB; color: white; padding: 10px 20px; text-decoration: none; border-radius: 5px;">Verify Email</a>
			</p>
			<p style="font-size: 12px; color: #666;">Or copy this link: %s</p>
			<p>This link expires in 24 hours.</p>
		</body>
		</html>
	`, link, link)

	log.Printf("📧 Sending Verification Email to %s", toEmail)
	return e.sendSMTP(toEmail, subject, body)
}

func (e *EmailNotifier) SendPasswordResetEmail(toEmail, token string) error {
	// Use dynamic URL
	link := fmt.Sprintf("%s/reset-password?token=%s", e.frontendURL, token)
	subject := "Reset your Password"

	body := fmt.Sprintf(`
		<html>
		<body style="font-family: Arial, sans-serif; padding: 20px;">
			<h2>Password Reset Request</h2>
			<p>We received a request to reset your password. If this wasn't you, please ignore this email.</p>
			<p>
				<a href="%s" style="background-color: #DC2626; color: white; padding: 10px 20px; text-decoration: none; border-radius: 5px;">Reset Password</a>
			</p>
			<p style="font-size: 12px; color: #666;">Or copy this link: %s</p>
			<p>This link expires in 1 hour.</p>
		</body>
		</html>
	`, link, link)

	log.Printf("📧 Sending Password Reset Email to %s", toEmail)
	return e.sendSMTP(toEmail, subject, body)
}

// --- Internal Helpers ---

// sendSMTP is the generic function that actually talks to Brevo
func (e *EmailNotifier) sendSMTP(to, subject, body string) error {
	// 1. Setup Auth
	auth := smtp.PlainAuth("", e.cfg.Username, e.cfg.Password, e.cfg.Host)
	addr := fmt.Sprintf("%s:%d", e.cfg.Host, e.cfg.Port)

	// 2. Build Headers (MIME)
	headers := make(map[string]string)
	headers["From"] = e.cfg.SenderAddr
	headers["To"] = to
	headers["Subject"] = subject
	headers["MIME-Version"] = "1.0"
	headers["Content-Type"] = "text/html; charset=\"UTF-8\""

	headerStr := ""
	for k, v := range headers {
		headerStr += fmt.Sprintf("%s: %s\r\n", k, v)
	}
	message := []byte(headerStr + "\r\n" + body)

	// 3. Send
	return smtp.SendMail(addr, auth, e.cfg.SenderAddr, []string{to}, message)
}

// buildAlertHTML generates the table for certificate alerts (Legacy logic preserved)
func (e *EmailNotifier) buildAlertHTML(user model.User, certs []model.CertResponse) string {
	var sb strings.Builder
	sb.WriteString("<html><body style='font-family: Arial, sans-serif; color: #333;'>")
	sb.WriteString(fmt.Sprintf("<h3>Hello %s,</h3>", user.OrgName))
	sb.WriteString(fmt.Sprintf("<p>The following <strong>%d certificates</strong> are expiring soon:</p>", len(certs)))

	sb.WriteString("<table border='1' cellpadding='10' cellspacing='0' style='border-collapse: collapse; width: 100%; border-color: #ddd;'>")
	sb.WriteString("<tr style='background-color: #f8f9fa; text-align: left;'><th>Host</th><th>Certificate</th><th>Expires</th></tr>")

	for _, cert := range certs {
		daysLeft := int(time.Until(cert.ValidUntil).Hours() / 24)
		color := "#28a745"
		if daysLeft < 7 {
			color = "#dc3545"
		} else if daysLeft < 30 {
			color = "#ffc107"
		}

		sb.WriteString("<tr>")
		sb.WriteString(fmt.Sprintf("<td>%s<br/><small>%s</small></td>", cert.AgentHostname, cert.SourceUID))
		sb.WriteString(fmt.Sprintf("<td>CN=%s<br/><small>Issuer: %s</small></td>", cert.Subject.CN, cert.Issuer.CN))
		sb.WriteString(fmt.Sprintf("<td><b style='color:%s'>%s</b><br/><small>%d days left</small></td>", color, cert.ValidUntil.Format("2006-01-02"), daysLeft))
		sb.WriteString("</tr>")
	}
	sb.WriteString("</table></body></html>")
	return sb.String()
}
