package scanner

import (
	"cert-manager-backend/internal/model"
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"log"
	"net"
	"time"
)

// TLSScanner is the concrete implementation using crypto/tls
type TLSScanner struct {
	Dialer *net.Dialer
}

// NewTLSScanner creates a scanner with a default timeout
func NewTLSScanner(timeout time.Duration) *TLSScanner {
	return &TLSScanner{
		Dialer: &net.Dialer{
			Timeout: timeout,
		},
	}
}

// Singleton for System Roots
var systemRoots *x509.CertPool

func init() {
	var err error
	// Load the system roots once at startup
	systemRoots, err = x509.SystemCertPool()
	if err != nil {
		log.Printf("⚠️ Scanner Warning: Failed to load system root certificates: %v", err)
		// Fallback to empty pool (verification will likely fail, but we don't crash)
		systemRoots = x509.NewCertPool()
	}
}

// Scan performs the TLS handshake and extracts the certificate chain.
func (s *TLSScanner) Scan(ctx context.Context, target string) ([]model.Certificate, error) {

	// plit Host and Port to get the SNI ServerName
	host, _, err := net.SplitHostPort(target)
	if err != nil {
		// Fallback: If split fails (e.g. no port), assume target is just the host
		host = target
		target = net.JoinHostPort(host, "443")
	}

	// 1. Prepare Config (InsecureSkipVerify: true to capture expired/self-signed certs)
	// CRITICAL: ServerName must be set for SNI to work on virtual hosts
	cfg := &tls.Config{
		ServerName:         host,
		InsecureSkipVerify: true,
	}

	// 2. Dial (with Context)
	conn, err := s.Dialer.DialContext(ctx, "tcp", target)
	if err != nil {
		return nil, fmt.Errorf("dial failed: %w", err)
	}
	defer conn.Close()

	// 3. Upgrade to TLS
	tlsConn := tls.Client(conn, cfg)

	// Force Handshake with deadline
	tlsConn.SetDeadline(time.Now().Add(s.Dialer.Timeout))
	if err := tlsConn.Handshake(); err != nil {
		return nil, fmt.Errorf("tls handshake failed: %w", err)
	}

	// 4. Extract Certificates
	state := tlsConn.ConnectionState()
	if len(state.PeerCertificates) == 0 {
		return nil, fmt.Errorf("no certificates found")
	}

	// 5. Convert to Domain Model (Leaf only for now)
	leaf := state.PeerCertificates[0]

	// Prepare intermediates for trust verification
	intermediates := x509.NewCertPool()
	for _, c := range state.PeerCertificates[1:] {
		intermediates.AddCert(c)
	}

	isTrusted, trustErr := verifyTrust(leaf, intermediates)

	cert := model.Certificate{
		SourceUID:  target,
		SourceType: "CLOUD",
		Serial:     leaf.SerialNumber.String(),
		Subject: model.DN{
			CN:  leaf.Subject.CommonName,
			Org: join(leaf.Subject.Organization),
			OU:  join(leaf.Subject.OrganizationalUnit),
		},
		Issuer: model.DN{
			CN:  leaf.Issuer.CommonName,
			Org: join(leaf.Issuer.Organization),
			OU:  join(leaf.Issuer.OrganizationalUnit),
		},
		SignatureAlgo: leaf.SignatureAlgorithm.String(),
		ValidFrom:     leaf.NotBefore,
		ValidUntil:    leaf.NotAfter,
		DNSNames:      leaf.DNSNames,
		IsTrusted:     isTrusted,
		TrustError:    trustErr,
	}

	return []model.Certificate{cert}, nil
}

// verifyTrust checks if the cert is trusted by the System Root CAs
func verifyTrust(leaf *x509.Certificate, intermediates *x509.CertPool) (bool, string) {
	opts := x509.VerifyOptions{
		Intermediates: intermediates,
		CurrentTime:   time.Now(),
		Roots:         systemRoots, // Explicitly use our cached pool
	}

	if _, err := leaf.Verify(opts); err != nil {
		return false, err.Error()
	}
	return true, ""
}

func join(s []string) string {
	if len(s) == 0 {
		return ""
	}
	return s[0]
}
