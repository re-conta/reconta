// Package turnstile verifica tokens do Cloudflare Turnstile no backend,
// evitando cadastros automatizados por bots.
package turnstile

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const verifyURL = "https://challenges.cloudflare.com/turnstile/v0/siteverify"

// Verifier valida tokens do widget Turnstile contra a API do Cloudflare.
type Verifier struct {
	secretKey string
	client    *http.Client
}

// NewVerifier cria um Verifier. Se secretKey for vazio, Verify sempre retorna
// true (proteção desabilitada), permitindo rodar sem Turnstile configurado.
func NewVerifier(secretKey string) *Verifier {
	return &Verifier{
		secretKey: secretKey,
		client:    &http.Client{Timeout: 5 * time.Second},
	}
}

// Enabled indica se a verificação está ativa (secretKey configurada).
func (v *Verifier) Enabled() bool {
	return v.secretKey != ""
}

type siteverifyResponse struct {
	Success bool `json:"success"`
}

// Verify checa o token enviado pelo widget do cliente. remoteIP é opcional e
// ajuda o Cloudflare a avaliar o risco da requisição.
func (v *Verifier) Verify(ctx context.Context, token, remoteIP string) (bool, error) {
	if !v.Enabled() {
		return true, nil
	}
	if strings.TrimSpace(token) == "" {
		return false, nil
	}

	form := url.Values{}
	form.Set("secret", v.secretKey)
	form.Set("response", token)
	if remoteIP != "" {
		form.Set("remoteip", remoteIP)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, verifyURL, strings.NewReader(form.Encode()))
	if err != nil {
		return false, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := v.client.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	var result siteverifyResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return false, err
	}
	return result.Success, nil
}
