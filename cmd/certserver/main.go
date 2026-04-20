package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"

	certserverpkg "github.com/docker/docker-agent/pkg/certserver"
	vaultapi "github.com/hashicorp/vault/api"
	"golang.org/x/crypto/ssh"
)

func main() {
	// Basic wiring: Vault-based CA material
	vaultAddr := os.Getenv("VAULT_ADDR")
	vaultToken := os.Getenv("VAULT_TOKEN")
	secretPath := os.Getenv("VAULT_CA_PATH")
	if vaultAddr == "" || vaultToken == "" || secretPath == "" {
		log.Fatal("VAULT_ADDR, VAULT_TOKEN and VAULT_CA_PATH must be set for certserver")
	}
	// Configure Vault client
	cfg := vaultapi.DefaultConfig()
	cfg.Address = vaultAddr
	client, err := vaultapi.NewClient(cfg)
	if err != nil {
		log.Fatalf("failed to create Vault client: %v", err)
	}
	client.SetToken(vaultToken)
	caStore := &certserverpkg.VaultCAStore{Client: client, SecretPath: secretPath}
	svc := certserverpkg.NewCertServer(caStore)

	http.HandleFunc("/api/generate-cert", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		var req certserverpkg.CertificateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if req.ValidityDays <= 0 {
			req.ValidityDays = 365
		}
		// Expect a public key be provided
		leafPub, _, _, _, err := ssh.ParseAuthorizedKey([]byte(req.PublicKey))
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		res, err := svc.GenerateFromRequest(req)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(res)
		_ = leafPub
	})

	log.Println("certserver listening on :8080 (Vault-backed CA)")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
