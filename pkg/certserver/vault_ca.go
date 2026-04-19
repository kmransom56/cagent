package certserver

import (
    vaultapi "github.com/hashicorp/vault/api"
    "fmt"
    "crypto/rand"
    "time"
    "golang.org/x/crypto/ssh"
    "encoding/pem"
    "crypto/x509"
    "errors"
)

// This file provides a Vault-based CAStore integration. It reads a PEM-encoded
// CA private key from Vault and signs SSH user certificates.

// NewVaultCAStore constructs a Vault-backed CA store.
type VaultCAStore struct {
    Client     *vaultapi.Client
    SecretPath string
}

// Ensure Vault CA store implements CAStore
var _ CAStore = (*VaultCAStore)(nil)

// SignSSH signs an SSH user certificate using the CA private key loaded from Vault.
func (v *VaultCAStore) SignSSH(leafPub ssh.PublicKey, username string, validityDays int) (string, ssh.PublicKey, time.Time, time.Time, string, string, error) {
    if v == nil || v.Client == nil {
        return "", nil, time.Time{}, time.Time{}, "", "", fmt.Errorf("vault client not configured")
    }
    sec, err := v.Client.Logical().Read(v.SecretPath)
    if err != nil {
        return "", nil, time.Time{}, time.Time{}, "", "", fmt.Errorf("vault read: %w", err)
    }
    if sec == nil {
        return "", nil, time.Time{}, time.Time{}, "", "", fmt.Errorf("vault secret not found at %s", v.SecretPath)
    }
    var pemBytes []byte
    // Support common kv shapes
    if data, ok := sec.Data["data"].(map[string]interface{}); ok {
        if s, ok := data["ca_private_key_pem"].(string); ok {
            pemBytes = []byte(s)
        }
    }
    if len(pemBytes) == 0 {
        if s, ok := sec.Data["ca_private_key_pem"].(string); ok {
            pemBytes = []byte(s)
        }
    }
    if len(pemBytes) == 0 {
        return "", nil, time.Time{}, time.Time{}, "", "", errors.New("ca private key not found in Vault secret")
    }
    signer, err := parsePEMToSigner(pemBytes)
    if err != nil {
        return "", nil, time.Time{}, time.Time{}, "", "", fmt.Errorf("parse ca signer: %w", err)
    }
    // leaf signer is the client's pubkey; we sign the leaf cert with CA
    caPub := signer.PublicKey()
    notBefore := time.Now().Add(-1 * time.Minute)
    notAfter := time.Now().Add(time.Duration(validityDays) * 24 * time.Hour)
    cert := &ssh.Certificate{
        Serial:          uint64(time.Now().UnixNano()),
        Key:             leafPub,
        CertType:        ssh.UserCert,
        KeyId:           username,
        ValidAfter:      uint64(notBefore.Unix()),
        ValidBefore:     uint64(notAfter.Unix()),
        ValidPrincipals: []string{username},
        Permissions:     ssh.Permissions{},
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
    return string(certLine), caPub, notBefore, notAfter, serial, string(caLine), nil
}
