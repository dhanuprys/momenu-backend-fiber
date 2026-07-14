package utils

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/url"

	"github.com/dhanuprys/momenu-backend-fiber/internal/config"
)

const turnstileVerifyURL = "https://challenges.cloudflare.com/turnstile/v0/siteverify"

type TurnstileResponse struct {
	Success    bool     `json:"success"`
	ErrorCodes []string `json:"error-codes"`
}

// VerifyTurnstile verifies the Cloudflare Turnstile token
func VerifyTurnstile(token string) error {
	secret := config.AppConfig.TurnstileSecret
	if secret == "" {
		return errors.New("turnstile secret is not configured")
	}

	res, err := http.PostForm(turnstileVerifyURL, url.Values{
		"secret":   {secret},
		"response": {token},
	})
	if err != nil {
		return err
	}
	defer res.Body.Close()

	var turnstileRes TurnstileResponse
	if err := json.NewDecoder(res.Body).Decode(&turnstileRes); err != nil {
		return err
	}

	if !turnstileRes.Success {
		return errors.New("captcha verification failed")
	}

	return nil
}
