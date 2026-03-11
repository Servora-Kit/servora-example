package jwks

import (
	"encoding/json"
	"net/http"
)

func NewJWKSHandler(km *KeyManager) http.HandlerFunc {
	resp := km.JWKSResponse()
	data, _ := json.Marshal(resp)

	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Cache-Control", "public, max-age=3600")
		w.Write(data)
	}
}

type oidcDiscovery struct {
	Issuer                           string   `json:"issuer"`
	JWKSURI                          string   `json:"jwks_uri"`
	IDTokenSigningAlgValuesSupported []string `json:"id_token_signing_alg_values_supported"`
}

func NewOIDCDiscoveryHandler(issuerURL string) http.HandlerFunc {
	disc := oidcDiscovery{
		Issuer:                           issuerURL,
		JWKSURI:                          issuerURL + "/.well-known/jwks.json",
		IDTokenSigningAlgValuesSupported: []string{"RS256"},
	}
	data, _ := json.Marshal(disc)

	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Cache-Control", "public, max-age=3600")
		w.Write(data)
	}
}
