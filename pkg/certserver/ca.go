package certserver

import (
    "crypto/rand"
    "encoding/pem"
    "fmt"
    "time"

    "golang.org/x/crypto/ssh"
    vaultapi "github.com/hashicorp/vault/api"
    "errors"
    "crypto/x509"
)

// CertificateRequest describes a request to generate an SSH user certificate.
type CertificateRequest struct {
    ServerName      string `json:"serverName"`
    ServerIP        string `json:"serverIp"`
    PublicKey       string `json:"publicKey"`
    CertificateType string `json:"certificateType"` // "user" or "host"
    ValidityDays    int    `json:"validityDays"`
    OutputFormat    string `json:"outputFormat"`  // reserved for future TLS certs, unused here
    KeySize         int    `json:"keySize"`       // reserved for leaf key generation if needed
}

// CertificateResponse is the API response for a generated certificate.
type CertificateResponse struct {
    CertificateLine    string `json:"certificateLine"`
    NotBefore          string `json:"notBefore"`
    NotAfter           string `json:"notAfter"`
    Serial             string `json:"serial"`
    CAKeyLine          string `json:"caKeyLine,omitempty"`
}

// CAStore abstracts a source of CA material capable of signing SSH certificates.
type CAStore interface {
    // SignSSH signs a leaf public key for a given user/principal and returns
    // the OpenSSH certificate line (ready to put in authorized_keys) and
    // the CA public key line to configure TrustedUserCAKeys.
    SignSSH(leafPub ssh.PublicKey, username string, validityDays int) (certLine string, caPub ssh.PublicKey, notBefore time.Time, notAfter time.Time, serial string, caLine string, err error)
}

// VaultCAStore implements CAStore using a Vault-backed CA private key.
// It fetches the CA private key PEM from Vault and signs SSH certificates.
type VaultCAStore struct {
    Client     *vaultapi.Client
    SecretPath string
}

// SignSSH implements CA signing using Vault-stored CA key.
func (v *VaultCAStore) SignSSH(leafPub ssh.PublicKey, username string, validityDays int) (string, ssh.PublicKey, time.Time, time.Time, string, string, error) {
    // Fetch CA key PEM from Vault
    if v.Client == nil {
        return "", nil, time.Time{}, time.Time{}, "", "", errors.New("vault client not configured")
    }
    secret, err := v.Client.Logical().Read(v.SecretPath)
    if err != nil {
        return "", nil, time.Time{}, time.Time{}, "", "", fmt.Errorf("vault read: %w", err)
    }
    if secret == nil {
        return "", nil, time.Time{}, time.Time{}, "", "", errors.New("vault secret not found")
    }

    // Support both KV v1 style and KV v2 style shapes
    var pemBytes []byte
    // attempt common shapes
    if data, ok := secret.Data["data"].(map[string]interface{}); ok {
        if s, ok := data["ca_private_key_pem"].(string); ok {
            pemBytes = []byte(s)
        }
    }
    if pemBytes == nil {
        if s, ok := secret.Data["ca_private_key_pem"].(string); ok {
            pemBytes = []byte(s)
        }
    }
    if len(pemBytes) == 0 {
        return "", nil, time.Time{}, time.Time{}, "", "", errors.New("ca private key not found in Vault secret")
    }

    // Build SSH signer from PEM
    signer, err := parsePEMToSigner(pemBytes)
    if err != nil {
        return "", nil, time.Time{}, time.Time{}, "", "", fmt.Errorf("parse CA signer: %w", err)
    }

    // Build leaf cert
    caPub := signer.PublicKey()
    notBefore := time.Now().Add(-1 * time.Minute)
    notAfter := time.Now().Add(time.Duration(validityDays) * 24 * time.Hour)

    crit := map[string]string{}
    extensions := map[string]string{}
    cert := &ssh.Certificate{
        Serial:           uint64(time.Now().UnixNano()),
        Key:              leafPub,
        CertType:         ssh.UserCert,
        KeyId:            username,
        ValidAfter:       uint64(notBefore.Unix()),
        ValidBefore:      uint64(notAfter.Unix()),
        ValidPrincipals:  []string{username},
        Permissions: ssh.Permissions{Extensions: extensions, Critical: crit},
    }

    if err := cert.Sign(rand.Reader, signer); err != nil {
        return "", nil, time.Time{}, time.Time{}, "", "", fmt.Errorf("sign cert: %w", err)
    }

    certLine, err := ssh.MarshalAuthorizedKey(cert)
    if err != nil {
        return "", nil, time.Time{}, time.Time{}, "", "", err
    }

    serial := fmt.Sprintf("%d", cert.Serial)
    caLine := ssh.MarshalAuthorizedKey(caPub)
    // caLine already contains a trailing newline; keep as string
    return string(certLine), caPub, notBefore, notAfter, serial, string(caLine), nil
}

// parsePEMToSigner tries to convert PEM-encoded private keys into an ssh.Signer.
func parsePEMToSigner(pemBytes []byte) (ssh.Signer, error) {
    // Try direct PEM private key parsing (OpenSSH, PEM PKCS8, PKCS1, etc.)
    if signer, err := ssh.ParsePrivateKey(pemBytes); err == nil {
        return signer, nil
    }
    // Fallback: try PEM block parsing
    block, _ := pem.Decode(pemBytes)
    if block == nil {
        return nil, fmt.Errorf("invalid PEM block")
    }
    // Support PKCS#8 PRIVATE KEY
    if block.Type == "PRIVATE KEY" {
        // Try to parse PKCS#8
        if key, err := x509.ParsePKCS8PrivateKey(block.Bytes); err == nil {
            return ssh.NewSignerFromKey(key)
        } else {
            return nil, err
        }
    }
    // Support generic PRIVATE KEY as PKCS#1/SEC1 could be Intentionally skipped here
    return nil, errors.New("unsupported private key format in PEM")
}
