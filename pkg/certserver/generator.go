package certserver

import (
    "encoding/pem"
    "time"
    "strings"
    "github.com/hashicorp/vault/api"
    "golang.org/x/crypto/ssh"
)

// CertServer orchestrates certificate generation using a CAStore.
type CertServer struct {
    CA CAStore
}

func NewCertServer(ca CAStore) *CertServer {
    return &CertServer{CA: ca}
}

// GenerateFromRequest validates input and signs an SSH user certificate using the CA.
// The leaf public key must be provided in the CertificateRequest as PublicKey.
func (s *CertServer) GenerateFromRequest(req CertificateRequest) (CertificateResponse, error) {
    // Basic validation
    if strings.TrimSpace(req.PublicKey) == "" {
        return CertificateResponse{}, ErrBadRequest("publicKey is required")
    }
    // parse leaf public key
    pubKey, _, _, _, err := ssh.ParseAuthorizedKey([]byte(req.PublicKey))
    if err != nil {
        return CertificateResponse{}, err
    }
    if s.CA == nil {
        return CertificateResponse{}, ErrBadRequest("certificate authority not configured")
    }

    certLine, caPub, notBefore, notAfter, serial, caLine, err := s.CA.SignSSH(pubKey, req.ServerName, req.ValidityDays)
    if err != nil {
        return CertificateResponse{}, err
    }
    resp := CertificateResponse{
        CertificateLine: certLine,
        NotBefore:       notBefore.Format(time.RFC3339),
        NotAfter:        notAfter.Format(time.RFC3339),
        Serial:          serial,
        CAKeyLine:       caLine,
    }
    _ = caPub // unused: kept to show the CA public key is available for response if needed
    return resp, nil
}

// ErrBadRequest is a tiny helper to wrap errors that mean client input was wrong
type BadRequestError string

func (e BadRequestError) Error() string { return string(e) }

func ErrBadRequest(msg string) error { return BadRequestError(msg) }
