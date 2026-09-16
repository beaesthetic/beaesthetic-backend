package http

import (
	"crypto/subtle"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/petretiandrea/beaesthetic-backend/identity-service/internal/oauth"
	"github.com/petretiandrea/beaesthetic-backend/identity-service/internal/token"
)

type Config struct {
	InternalAPIKey string
}

func NewHandler(cfg Config, issuer *token.Issuer, exchange *oauth.ExchangeService) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", health)
	mux.HandleFunc("GET /v1/internal/keys", func(w http.ResponseWriter, r *http.Request) {
		if !authorize(r, cfg.InternalAPIKey) {
			unauthorized(w)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"keys": issuer.PublicKeys()})
	})
	mux.HandleFunc("POST /v1/internal/tokens", func(w http.ResponseWriter, r *http.Request) {
		if !authorize(r, cfg.InternalAPIKey) {
			unauthorized(w)
			return
		}
		defer r.Body.Close()
		var request token.IssueRequest
		decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 16<<10))
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&request); err != nil {
			badRequest(w, err)
			return
		}
		accessToken, claims, err := issuer.Issue(request)
		if err != nil {
			badRequest(w, err)
			return
		}
		writeJSON(w, http.StatusCreated, map[string]any{
			"access_token": accessToken,
			"token_type":   "Bearer",
			"expires_in":   int64(claims.ExpiresAt.Sub(claims.IssuedAt).Seconds()),
			"claims":       claims,
		})
	})
	mux.HandleFunc("POST /v1/internal/tokens/verify", func(w http.ResponseWriter, r *http.Request) {
		if !authorize(r, cfg.InternalAPIKey) {
			unauthorized(w)
			return
		}
		defer r.Body.Close()
		var request struct {
			Token    string `json:"token"`
			Audience string `json:"audience"`
		}
		decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 32<<10))
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&request); err != nil {
			badRequest(w, err)
			return
		}
		claims, err := issuer.Verify(request.Token, request.Audience)
		if err != nil {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid_token"})
			return
		}
		writeJSON(w, http.StatusOK, claims)
	})
	mux.HandleFunc("POST /oauth/token", func(w http.ResponseWriter, r *http.Request) {
		if exchange == nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "temporarily_unavailable"})
			return
		}
		if err := r.ParseForm(); err != nil {
			badRequest(w, err)
			return
		}
		accessToken, claims, err := exchange.Exchange(r.Context(), oauth.Request{
			GrantType: r.Form.Get("grant_type"), SubjectTokenType: r.Form.Get("subject_token_type"), Authenticator: r.Form.Get("authenticator"), Authorizer: r.Form.Get("authorizer"), SubjectToken: r.Form.Get("subject_token"),
			Audience: r.Form.Get("audience"), OrganizationID: r.Form.Get("organization_id"),
		})
		if err != nil {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid_grant"})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"access_token": accessToken, "token_type": "Bearer", "expires_in": int64(claims.ExpiresAt.Sub(claims.IssuedAt).Seconds())})
	})
	return mux
}

func health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func authorize(r *http.Request, expected string) bool {
	provided := r.Header.Get("X-Internal-API-Key")
	return expected != "" && subtle.ConstantTimeCompare([]byte(provided), []byte(expected)) == 1
}

func unauthorized(w http.ResponseWriter) {
	writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
}

func badRequest(w http.ResponseWriter, err error) {
	message := "invalid request"
	if errors.Is(err, http.ErrBodyReadAfterClose) {
		message = "invalid request body"
	}
	writeJSON(w, http.StatusBadRequest, map[string]string{"error": message})
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
