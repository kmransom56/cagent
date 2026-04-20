package certserver

import (
	"time"

	"golang.org/x/crypto/ssh"
)

// CertificateRequest describes a request to generate an SSH user certificate.
type CertificateRequest struct {
	ServerName      string `json:"serverName"`
	ServerIP        string `json:"serverIp"`
	PublicKey       string `json:"publicKey"`
	CertificateType string `json:"certificateType"` // "user" or "host"
	ValidityDays    int    `json:"validityDays"`
	OutputFormat    string `json:"outputFormat"` // reserved for future TLS certs, unused here
	KeySize         int    `json:"keySize"`      // reserved for leaf key generation if needed
}

// CertificateResponse is the API response for a generated certificate.
type CertificateResponse struct {
	CertificateLine string `json:"certificateLine"`
	NotBefore       string `json:"notBefore"`
	NotAfter        string `json:"notAfter"`
	Serial          string `json:"serial"`
	CAKeyLine       string `json:"caKeyLine,omitempty"`
}

// CAStore abstracts a source of CA material capable of signing SSH certificates.
type CAStore interface {
	// SignSSH signs a leaf public key for a given user/principal and returns
	// the OpenSSH certificate line (ready to put in authorized_keys) and
	// the CA public key line to configure TrustedUserCAKeys.
	SignSSH(leafPub ssh.PublicKey, username string, validityDays int) (certLine string, caPub ssh.PublicKey, notBefore time.Time, notAfter time.Time, serial string, caLine string, err error)
}
