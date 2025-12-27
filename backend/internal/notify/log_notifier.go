package notify

import (
	"cert-manager-backend/internal/model"
	"context"
	"fmt"
	"log"
	"strings"
	"time"
)

type LogNotifier struct{}

func NewLogNotifier() *LogNotifier {
	return &LogNotifier{}
}

// Notify implements the service.Notifier interface.
// REFACTOR: Renamed from 'SendExpiryAlert' to 'Notify'.
// It now accepts the RAW list of certs and the USER MAP (Lookaside pattern).
func (n *LogNotifier) Notify(ctx context.Context, certs []model.CertResponse, users map[string]model.User) error {
	if len(certs) == 0 {
		return nil
	}

	// 1. Generate the Subject (Global summary)
	subject := fmt.Sprintf("Action Required: %d Certificates Expiring Soon", len(certs))

	// 2. Render the Body (The "View" Logic)
	body := n.renderBody(certs, users)

	// 3. "Send" it
	log.Println("---------------------------------------------------")
	log.Printf("🔔 [LogNotifier] ALERT SYSTEM")
	log.Printf("Subject: %s", subject)
	log.Printf("Body:\n%s", body)
	log.Println("---------------------------------------------------")

	return nil
}

// renderBody contains the presentation logic specific to Logs (Text).
// Accepts 'users' map to resolve OwnerID -> Email/OrgName.
func (n *LogNotifier) renderBody(certs []model.CertResponse, users map[string]model.User) string {
	var sb strings.Builder
	sb.WriteString("The following certificates are expiring soon:\n\n")

	for _, cert := range certs {
		// Helper to format DN (Distinguished Name) - Preserved from your original code
		formatDN := func(dn model.DN) string {
			var parts []string
			if dn.CN != "" {
				parts = append(parts, "CN="+dn.CN)
			}
			if dn.Org != "" {
				parts = append(parts, "Org="+dn.Org)
			}
			if dn.OU != "" {
				parts = append(parts, "OU="+dn.OU)
			}
			if len(parts) == 0 {
				return "(Unknown)"
			}
			return strings.Join(parts, ", ")
		}

		// Resolve Owner Information using the Lookaside Map
		user, exists := users[cert.OwnerID]
		ownerInfo := "[Unknown Owner]"
		if exists {
			ownerInfo = fmt.Sprintf("[%s | %s]", user.Email, user.OrgName)
		}

		// Calculate precise days left for the log
		daysLeft := int(time.Until(cert.ValidUntil).Hours() / 24)

		subjectStr := formatDN(cert.Subject)
		issuerStr := formatDN(cert.Issuer)

		sb.WriteString(fmt.Sprintf("🔴 [Host: %s] %s\n", cert.AgentHostname, ownerInfo))
		sb.WriteString(fmt.Sprintf("   Subject: %s\n", subjectStr))
		sb.WriteString(fmt.Sprintf("   Issuer:  %s\n", issuerStr))
		sb.WriteString(fmt.Sprintf("   Expires: %s (%s - %d days left)\n", cert.ValidUntil.Format("2006-01-02"), cert.Status, daysLeft))
		sb.WriteString("\n")
	}

	sb.WriteString("Please update these certificates to avoid service interruption.\n")
	return sb.String()
}
